package accounts

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/AlexGalhardo/Scooby/app/models"
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/social"
	"strings"
	"time"
)

func NewSocial(app *vukem.App) (*social.Social, error) {
	key, err := app.DeriveKey("social.state", 32)
	if err != nil {
		return nil, err
	}
	s := social.New(key).WithPreviousKeys(app.DerivePreviousKeys("social.state", 32)...)
	s.InsecureCookies = !app.Config.IsProduction()
	if id := app.Source.Get("SOCIAL_GOOGLE_CLIENT_ID"); id != "" {
		s.Use(social.Google(social.Config{ClientID: id, ClientSecret: app.Source.Get("SOCIAL_GOOGLE_CLIENT_SECRET"), RedirectURL: strings.TrimSuffix(app.Config.URL, "/") + "/auth/google/callback"}))
	}
	if app.Kind == vukem.KindTest && app.Source.Get("SOCIAL_FAKE_ID") != "" {
		s.Use(social.Fake("google", "/auth/google/callback", social.User{ID: app.Source.Get("SOCIAL_FAKE_ID"), Email: app.Source.Get("SOCIAL_FAKE_EMAIL"), Name: "Owner", EmailVerified: app.Source.Get("SOCIAL_FAKE_EMAIL_VERIFIED") == "true"}))
	}
	c, err := app.Cache()
	if err != nil {
		return nil, err
	}
	s.States = c.Store()
	return s, nil
}

// The singleton owner slot makes bootstrap atomic even with concurrent callbacks.
// Once linked, changing the configured email cannot silently replace the subject.
func (a *Accounts) FindOrCreateFromSocial(ctx context.Context, su social.User) (models.User, error) {
	var u models.User
	if su.Provider != "google" || su.ID == "" || !su.EmailVerified {
		return u, errs.New(403, "Conta Google não autorizada")
	}
	err := a.DB.Tx(ctx, func(tx *database.Tx) error {
		var subject string
		var uid int64
		q := "SELECT subject,user_id FROM owner_identity WHERE id=1"
		if tx.Dialect().Name() == "postgres" {
			q += " FOR UPDATE"
		}
		if err := tx.QueryRowContext(ctx, q).Scan(&subject, &uid); err != nil {
			return err
		}
		if subject != "" {
			if subject != su.ID {
				return errs.New(403, "Conta Google não autorizada")
			}
			var err error
			u, err = orm.QueryOf[models.User](tx).Find(ctx, uid)
			return err
		}
		if strings.TrimSpace(a.OwnerEmail) == "" || !strings.EqualFold(strings.TrimSpace(su.Email), strings.TrimSpace(a.OwnerEmail)) {
			return errs.New(403, "Conta Google não autorizada")
		}
		if uid > 0 {
			var err error
			u, err = orm.QueryOf[models.User](tx).Find(ctx, uid)
			if err != nil {
				return err
			}
			u.Email = strings.ToLower(su.Email)
			u.Name = su.Name
			if err := orm.Update(ctx, tx, &u); err != nil {
				return err
			}
		} else {
			u = models.User{Name: su.Name, Email: strings.ToLower(su.Email), Password: rand.Text() + rand.Text(), EmailVerifiedAt: orm.NullTime{Null: sql.Null[time.Time]{V: time.Now().UTC(), Valid: true}}}
			if err := orm.Create(ctx, tx, &u); err != nil {
				return err
			}
		}
		q = database.Rebind(tx.Dialect(), "UPDATE owner_identity SET subject=?,user_id=? WHERE id=1 AND subject=''")
		res, err := tx.ExecContext(ctx, q, su.ID, u.ID)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return errors.New("concurrent owner bootstrap; retry sign in")
		}
		return nil
	})
	if errors.Is(err, sql.ErrNoRows) {
		return u, errs.New(503, "Migrations pendentes")
	}
	return u, err
}
