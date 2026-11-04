// Package commands holds the console commands: the engine worker and the audited owner recovery.
package commands

import (
	"context"
	"errors"
	"flag"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
)

// OwnerReset is console-only, unavailable over HTTP. It revokes every existing
// session before reopening bootstrap against OWNER_GOOGLE_EMAIL.
type OwnerReset struct{ confirm, reason string }

func (*OwnerReset) Name() string { return "owner:reset" }
func (*OwnerReset) Description() string {
	return "Revoke owner sessions and reopen Google bootstrap (audited)"
}

func (c *OwnerReset) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.confirm, "confirm", "", "must be RESET-OWNER")
	fs.StringVar(&c.reason, "reason", "", "administrative recovery reason")
}

func (c *OwnerReset) Run(ctx context.Context, app *vukem.App, _ []string) error {
	if c.confirm != "RESET-OWNER" || len(c.reason) < 10 || len(c.reason) > 300 ||
		app.Source.Get("OWNER_GOOGLE_EMAIL") == "" {
		return errors.New("requires --confirm RESET-OWNER, --reason (10–300 chars), and configured OWNER_GOOGLE_EMAIL")
	}
	db, err := app.DB()
	if err != nil {
		return err
	}
	return db.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(ctx, "UPDATE users SET session_version=session_version+1"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE owner_identity SET subject='' WHERE id=1"); err != nil {
			return err
		}
		_, err := tx.ExecContext(
			ctx,
			database.Rebind(
				tx.Dialect(),
				"INSERT INTO audit_events(actor,action,account_id,details,created_at) VALUES (?,?,?,?,?)",
			),
			"operator",
			"owner.reset",
			"owner",
			c.reason,
			time.Now().UTC().Format(time.RFC3339),
		)
		return err
	})
}
