package main

import (
	"os"

	"github.com/kittors/AgentFlow/internal/app"
)

func main() {
	application := app.New(os.Stdout, os.Stderr)
	os.Exit(application.Run(os.Args[1:]))
}
