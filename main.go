/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/tvs/load-sv/cmd"
)

func main() {
	err := cmd.NewCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}
