package commands

import (
	"context"
	"errors"
	"flag"
	"github.com/AlexGalhardo/Scooby/internal/adapters"
	"github.com/AlexGalhardo/Scooby/internal/engine"
	"github.com/AlexGalhardo/Scooby/internal/trading"
	"github.com/AlexGalhardo/vukem"
	"net"
	"net/http"
	"time"
)

type Worker struct{}

func (Worker) Name() string        { return "trading:work" }
func (Worker) Description() string { return "Run paper trading, reconciliation and backtest jobs" }
func (Worker) Flags(*flag.FlagSet) {}
func (Worker) Run(ctx context.Context, app *vukem.App, args []string) error {
	db, err := app.DB()
	if err != nil {
		return err
	}
	store := trading.Store{DB: db}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	port := app.Source.Get("ENGINE_HEALTH_PORT")
	if port == "" {
		port = app.Source.Get("PORT")
	}
	if port == "" {
		port = "3001"
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ok := true
		for _, id := range []string{"b3", "us"} {
			var a trading.State
			if err := store.Get(r.Context(), "trading_accounts", id, &a); err != nil || time.Since(a.Heartbeat) > 45*time.Second {
				ok = false
			}
		}
		if !ok {
			http.Error(w, "not ready", 503)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ready"))
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cancel()
		}
	}()
	defer func() {
		shutdown, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = server.Shutdown(shutdown)
	}()
	return (engine.Engine{Store: store, Broker: adapters.Alpaca{Key: app.Source.Get("ALPACA_API_KEY"), Secret: app.Source.Get("ALPACA_API_SECRET")}, Log: app.Log, Alert: func(ctx context.Context, message string) {
		if err := adapters.SendAlert(ctx, app.Source.Get("TELEGRAM_BOT_TOKEN"), app.Source.Get("TELEGRAM_CHAT_ID"), message); err != nil {
			app.Log.Warn("alert failed", "error", err)
		}
	}}).Run(workerCtx)
}
