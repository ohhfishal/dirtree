package main

import (
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/ohhfishal/dirtree/tree"
	"os"
	"os/signal"
	"syscall"
)

type CLI struct {
	Path  string `arg:"" optional:"" type:"path" help:"Path to tree from."`
	Depth int    `default:"2" help:"Max depth to recurse."`
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var cli CLI
	kongCtx := kong.Parse(
		&cli,
		kong.BindTo(ctx, new(context.Context)),
	)

	if err := kongCtx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func (cmd *CLI) Run() error {
	if cmd.Path == "" {
		cmd.Path = "."
	}
	return tree.Print(cmd.Path, "", 0, cmd.Depth)
}
