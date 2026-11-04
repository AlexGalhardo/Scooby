package migrations

import "github.com/AlexGalhardo/vukem/schema"

func preferencesSchema() schema.Migration {
	return schema.Migration{ID: "20261104180000_create_preferences", Up: func(s *schema.Schema) error {
		return s.Create("preferences", func(t *schema.Table) {
			t.String("id").Primary()
			t.Text("payload")
			t.BigInteger("version").Default(0)
			t.String("updated_at")
			t.String("status").Default("ready")
		})
	}, Down: func(s *schema.Schema) error { return s.DropIfExists("preferences") }}
}
