package main

import (
	"github.com/opdev/opcap/cmd"
	"github.com/opdev/opcap/internal/logger"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logger.Sugar.Fatal(err)
	}
}
