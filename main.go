package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/opdev/opcap/cmd"
)

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
		return
	}
}
