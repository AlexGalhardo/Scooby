//go:build e2e

// This loopback-only fake OAuth server is excluded from production builds.
package main

import (
	"context"
	"github.com/AlexGalhardo/Scooby/bootstrap"
	"github.com/AlexGalhardo/Scooby/internal/trading"
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	app, err := bootstrap.New(vukem.WithSource(config.MapSource(map[string]string{"APP_ENV": "test", "APP_KEY": config.GenerateKey(), "APP_URL": "http://127.0.0.1:3100", "DATABASE_URL": "file:e2e?mode=memory&cache=shared", "MAIL_MAILER": "array", "SESSION_DRIVER": "cookie", "CACHE_STORE": "memory", "OWNER_GOOGLE_EMAIL": "owner@example.com", "SOCIAL_GOOGLE_CLIENT_ID": "test-client", "SOCIAL_FAKE_ID": "test-owner", "SOCIAL_FAKE_EMAIL": "owner@example.com", "SOCIAL_FAKE_EMAIL_VERIFIED": "true"})), vukem.WithKind(vukem.KindTest), vukem.WithoutHTTP())
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := app.Boot(ctx); err != nil {
		log.Fatal(err)
	}
	m, err := app.Migrator()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := m.Up(ctx); err != nil {
		log.Fatal(err)
	}
	db, err := app.DB()
	if err != nil {
		log.Fatal(err)
	}
	store := trading.Store{DB: db}
	go func() {
		for {
			_ = store.ProcessJob(ctx)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	h, err := app.Handler()
	if err != nil {
		log.Fatal(err)
	}
	if os.Getenv("APP_ENV") == "production" {
		log.Fatal("test server cannot run in production")
	}
	log.Fatal(http.ListenAndServe("127.0.0.1:3100", h))
}
