/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tvs/load-sv/cmd/load"
	"github.com/tvs/load-sv/cmd/version"
	"github.com/tvs/load-sv/pkg/util/log"
)

var (
	verbose bool
	output  string
)

func BindFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&verbose, "v", "v", false, "enable debug logging")
	flags.StringVarP(&output, "output", "o", "", "output format: one of (json, text)")
}

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
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if output != "" {
				if !strings.EqualFold("json", output) && !strings.EqualFold("text", output) {
					return fmt.Errorf("Unknown output type %q", output)
				}
			}

			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			var handler slog.Handler
			switch output {
			case "":
				fallthrough
			case "text":
				handler = tint.NewHandler(os.Stdout, &tint.Options{
					Level:   level,
					NoColor: !isatty.IsTerminal(os.Stdout.Fd()),
				})
			case "json":
				handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level: level,
				})
			}

			logger := slog.New(handler)
			cmd.SetContext(log.ContextWithLogger(cmd.Context(), logger))

			return nil
		},
	}

	BindFlags(c.PersistentFlags())

	c.AddCommand(
		load.NewCommand(),
		version.NewCommand(),
	)

	return c
}
