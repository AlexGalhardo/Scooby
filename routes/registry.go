package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/router"
)

func All() []func(*router.Router, *vukem.App) error {
	return []func(*router.Router, *vukem.App) error{Web, Social}
}
