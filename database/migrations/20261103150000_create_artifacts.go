package migrations

import "github.com/AlexGalhardo/vukem/schema"

func artifactsSchema() schema.Migration {
	return schema.Migration{ID: "20261103150000_create_artifacts", Up: func(s *schema.Schema) error {
		return s.Create("artifacts", func(t *schema.Table) {
			t.String("id").Primary()
			t.Text("payload")
			t.BigInteger("version").Default(0)
			t.String("updated_at")
			t.String("status").Default("ready").Index()
		})
	}, Down: func(s *schema.Schema) error { return s.DropIfExists("artifacts") }}
}
