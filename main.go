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

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var cli tree.CMD
	kongCtx := kong.Parse(
		&cli,
		kong.BindTo(ctx, new(context.Context)),
	)

	if err := kongCtx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
