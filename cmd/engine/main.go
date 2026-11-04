// Command engine runs the paper-trading worker: leader election, replay, dispatch and reconciliation.
package main

import (
	"os"

	"github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/vukem-trading/bootstrap"
)

func main() { os.Args = []string{os.Args[0], "trading:work"}; os.Exit(vukem.Main(bootstrap.New)) }
