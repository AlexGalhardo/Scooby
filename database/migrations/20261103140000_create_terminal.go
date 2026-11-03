package migrations

import (
	"github.com/AlexGalhardo/vukem/schema"
)

func terminalSchema() schema.Migration {
	return schema.Migration{ID: "20261103140000_create_terminal", Up: func(s *schema.Schema) error {
		if err := s.Create("owner_identity", func(t *schema.Table) { t.Integer("id").Primary(); t.String("subject"); t.BigInteger("user_id") }); err != nil {
			return err
		}
		if err := s.Raw("INSERT INTO owner_identity (id,subject,user_id) VALUES (1,'',0)"); err != nil {
			return err
		}
		for _, name := range []string{"trading_accounts", "datasets", "strategies", "backtests"} {
			if err := s.Create(name, func(t *schema.Table) {
				t.String("id").Primary()
				t.Text("payload")
				t.BigInteger("version").Default(0)
				t.String("updated_at")
				t.String("status").Default("ready").Index()
			}); err != nil {
				return err
			}
		}
		return s.Create("audit_events", func(t *schema.Table) {
			t.ID()
			t.String("actor")
			t.String("action")
			t.String("account_id")
			t.Text("details")
			t.String("created_at").Index()
		})
	}, Down: func(s *schema.Schema) error {
		for _, name := range []string{"audit_events", "backtests", "strategies", "datasets", "trading_accounts", "owner_identity"} {
			if err := s.DropIfExists(name); err != nil {
				return err
			}
		}
		return nil
	}}
}
