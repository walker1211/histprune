package main

import (
	"context"
	"os"

	"github.com/walker1211/histprune/internal/app"
	"github.com/walker1211/histprune/internal/cli"
)

func main() {
	runners := app.NewRunners(os.Stdout, os.Stderr)
	os.Exit(cli.Run(context.Background(), os.Args[1:], runners))
}
