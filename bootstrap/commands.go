package bootstrap

import (
	"github.com/AlexGalhardo/Scooby/app/commands"
	"github.com/AlexGalhardo/vukem"
)

func Commands() []vukem.Command { return []vukem.Command{commands.Worker{}, &commands.OwnerReset{}} }
