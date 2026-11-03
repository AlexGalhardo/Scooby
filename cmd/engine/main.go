package main

import (
	"github.com/AlexGalhardo/Scooby/bootstrap"
	"github.com/AlexGalhardo/vukem"
	"os"
)

func main() { os.Args = []string{os.Args[0], "trading:work"}; os.Exit(vukem.Main(bootstrap.New)) }
