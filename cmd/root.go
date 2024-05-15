/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tvs/load-sv/cmd/load"
	"github.com/tvs/load-sv/cmd/version"
)

// NewCommand represents the base command when called without any subcommands
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "load-sv",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}

	c.AddCommand(
		load.NewCommand(),
		version.NewCommand(),
	)

	return c
}
