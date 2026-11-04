// Package bootstrap wires the application together. There is no hidden
// registration: everything the app uses is listed here.
package bootstrap

import (
	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/inertia"

	"github.com/AlexGalhardo/vukem-trading/app/jobs"
	"github.com/AlexGalhardo/vukem-trading/app/listeners"
	"github.com/AlexGalhardo/vukem-trading/database/migrations"
	"github.com/AlexGalhardo/vukem-trading/database/seeders"
	"github.com/AlexGalhardo/vukem-trading/resources"
	"github.com/AlexGalhardo/vukem-trading/routes"
)

// New builds the application; cmd/app and the tests call it.
func New(opts ...vukem.Option) (*vukem.App, error) {
	base := []vukem.Option{
		vukem.WithName("vukem-trading"),
		vukem.WithEnvDefaults(
			map[string]string{"CACHE_STORE": "database", "SESSION_DRIVER": "database", "MAIL_MAILER": "log"},
		),
		vukem.WithDatabaseDrivers(postgres.Driver()),
		vukem.WithProviders(Providers()...),
		vukem.WithMigrations(migrations.All()),
		vukem.WithSeeders(seeders.All()...),
		vukem.WithListeners(listeners.All()...),
		vukem.WithJobsFunc(jobs.All),
		vukem.WithSchedule(routes.Schedule),
		vukem.WithViews(resources.Views()),
		vukem.WithVite(resources.Public()),
		vukem.WithInertia(inertia.Config{}),
	}

	for _, fn := range routes.All() {
		base = append(base, vukem.WithRoutes(fn))
	}

	app, err := vukem.New(append(base, opts...)...)
	if err != nil {
		return nil, err
	}

	app.AddCommand(Commands()...)

	return app, nil
}
