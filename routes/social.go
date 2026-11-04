package routes

import (
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/vukem-trading/app/accounts"
)

func Social(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}
	s, err := accounts.NewSocial(app)
	if err != nil {
		return err
	}
	limiters, err := app.RateLimiters()
	if err != nil {
		return err
	}
	throttle := ratelimit.Middleware(limiters("google.oauth", 20, time.Minute), ratelimit.ByIP)
	r.Get("/login", func(c *web.Ctx) error {
		if acc.Web.Check(c) {
			return c.Redirect("/")
		}
		return inertia.Render(
			c,
			"Auth/Login",
			inertia.Props{
				"configured": app.Source.Get("SOCIAL_GOOGLE_CLIENT_ID") != "" &&
					app.Source.Get("OWNER_GOOGLE_EMAIL") != "",
			},
		)
	}).Name("login")
	r.Get("/auth/google/redirect", func(c *web.Ctx) error { return s.Redirect(c, "google") }).
		Use(auth.Guest(acc.Web), throttle)
	r.Get("/auth/google/callback", func(c *web.Ctx) error {
		su, err := s.User(c, "google")
		if err != nil {
			return err
		}
		u, err := acc.FindOrCreateFromSocial(c.Context(), su)
		if err != nil {
			return err
		}
		if err := acc.Web.Login(c, u); err != nil {
			return err
		}
		return c.Redirect("/")
	}).Use(throttle)
	r.Post("/logout", func(c *web.Ctx) error {
		if err := acc.Web.Logout(c); err != nil {
			return err
		}
		return c.Redirect("/login")
	}).Use(auth.Required(acc.Web))
	return nil
}
