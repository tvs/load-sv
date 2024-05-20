/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"os"

	"github.com/tvs/ultravisor/cmd"
)

func main() {
	ctx := context.Background()
	err := cmd.NewCommand().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
