package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/opdev/opcap/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
