package accounts

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem-trading/app/models"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/auth/hash"
	"github.com/AlexGalhardo/vukem/database"
)

type Accounts struct {
	DB         *database.DB
	Web        *auth.SessionGuard[models.User]
	OwnerEmail string
}

func New(app *vukem.App) (*Accounts, error) {
	db, err := app.DB()
	if err != nil {
		return nil, err
	}
	guard := auth.NewSessionGuard("web", auth.NewORMProvider[models.User](db), hash.Default(), auth.NewRememberStore(db))
	guard.InsecureCookies = !app.Config.IsProduction()
	return &Accounts{DB: db, Web: guard, OwnerEmail: app.Source.Get("OWNER_GOOGLE_EMAIL")}, nil
}
