package tests

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/vukem-trading/bootstrap"
	"github.com/AlexGalhardo/vukem-trading/internal/engine"
	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

func TestPostgresConcurrencyAndRollback(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for dedicated PostgreSQL integration")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a dedicated *_test database")
	}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": dsn}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	store := trading.Store{DB: db}
	id := "test-" + trading.ID()
	a := trading.NewState(id, "USD", "simulation")
	a.Halted = false
	if err := store.Create(t.Context(), "trading_accounts", id, "ready", "test", a); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM trading_accounts WHERE id=$1", id)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM audit_events WHERE account_id=$1", id)
	})
	var success atomic.Int64
	var wg sync.WaitGroup
	for j := range 16 {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			err := store.Mutate(t.Context(), id, "test", "reserve", func(a *trading.State) error {
				_, err := a.Reserve(
					trading.Strategy{Symbol: "AAPL", StopBPS: 100, TargetBPS: 200},
					"buy",
					strconv.Itoa(j),
					trading.D("100"),
					time.Now(),
				)
				return err
			})
			if err == nil {
				success.Add(1)
			}
		}(j)
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatalf("expected one reserved intent, got %d", success.Load())
	}
	if err := store.Get(t.Context(), "trading_accounts", id, &a); err != nil {
		t.Fatal(err)
	}
	if len(a.Orders) != 1 {
		t.Fatal("lost or duplicated intention")
	}
	before := a.Cash
	err = store.Mutate(
		t.Context(),
		id,
		"test",
		"rollback",
		func(a *trading.State) error { a.Cash = trading.Zero; return errors.New("failure before commit") },
	)
	if err == nil {
		t.Fatal("expected rollback")
	}
	if err := store.Get(t.Context(), "trading_accounts", id, &a); err != nil || !a.Cash.Equal(before) {
		t.Fatal("rollback did not preserve cash", err)
	}
}

func TestPostgresWorkerFailover(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("dedicated PostgreSQL required")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("dedicated *_test database required")
	}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": dsn}})
	db, _ := app.DB()
	store := trading.Store{DB: db}
	if err := store.Initialize(t.Context()); err != nil {
		t.Fatal(err)
	}
	var before trading.State
	if err := store.Get(t.Context(), "trading_accounts", "b3", &before); err != nil {
		t.Fatal(err)
	}
	ctx1, stop1 := context.WithCancel(t.Context())
	defer stop1()
	ctx2, stop2 := context.WithCancel(t.Context())
	defer stop2()
	r1 := make(chan error, 1)
	r2 := make(chan error, 1)
	e := engine.Engine{Store: store, Log: slog.Default()}
	go func() { r1 <- e.Run(ctx1) }()
	waitEpoch := func(want int64) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			var a trading.State
			if store.Get(t.Context(), "trading_accounts", "b3", &a) == nil && a.LeaderEpoch == want {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("epoch %d not reached", want)
	}
	waitEpoch(before.LeaderEpoch + 1)
	go func() { r2 <- e.Run(ctx2) }()
	time.Sleep(350 * time.Millisecond)
	var current trading.State
	_ = store.Get(t.Context(), "trading_accounts", "b3", &current)
	if current.LeaderEpoch != before.LeaderEpoch+1 {
		t.Fatal("standby became leader too early")
	}
	stop1()
	if err := <-r1; err != nil {
		t.Fatal(err)
	}
	waitEpoch(before.LeaderEpoch + 2)
	stop2()
	if err := <-r2; err != nil {
		t.Fatal(err)
	}
}
