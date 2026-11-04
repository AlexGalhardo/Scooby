package main

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem-trading/bootstrap"
	"os"
)

func main() { os.Args = []string{os.Args[0], "trading:work"}; os.Exit(vukem.Main(bootstrap.New)) }
