package root

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tvs/ultravisor/pkg/config"
	"github.com/tvs/ultravisor/pkg/config/configmanager"
	"github.com/tvs/ultravisor/pkg/log"
)

var (
	verbose bool
	output  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ultravisor",
	Short: "A collection of tools for working with the vSphere IaaS Control Plane",
	Long: `Developing software on the vSphere IaaS Control Plane can be difficult.
Ultravisor aims to make a number of difficult or tedious tasks simple.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if rootCmdArgs.Profile != "" {
			config.SetProfile(rootCmdArgs.Profile)
		}

		l := configureLogger(cmd)
		l.Debug().Str("profile", config.CurrentProfile().Name).Msg("Using profile")

		cfg, err := configmanager.Load()
		if err != nil {
			l.Error().Err(err).Msg("Unable to load config")
			SetExitCode(1)
			return
		}
		cfg.SetDefaults()

		l.Debug().Any("config", cfg).Msg("Saving config")
		err = configmanager.Save(cfg)
		if err != nil {
			l.Error().Err(err).Msg("Unable to save config")
			SetExitCode(1)
			return
		}
	},
}

// Cmd returns the root command
func Cmd() *cobra.Command {
	return rootCmd
}

// Execute sets up logging, subcommands, and runs the root command. Errors are
// emitted as logs, but otherwise swallowed and converted into error codes.
// This is called by main.main() and ony needs to be called once.
func Execute() int {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		l := zerolog.Ctx(rootCmd.Context())
		l.Error().Err(err).Msg("An error occurred during execution")
	}

	return exitCode
}

// TODO(tvs): Convert exit codes to type and consts
var exitCode int = 0

// SetExitCode sets the exit code to be thrown by the command or subcommands.
func SetExitCode(c int) {
	exitCode = c
}

// rootCmdArgs holds the flags defined for the root command
var rootCmdArgs struct {
	Profile string
	Verbose bool
	Json    bool
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdArgs.Verbose, "verbose", "v", rootCmdArgs.Verbose, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&rootCmdArgs.Profile, "profile", "p", "default", "profile name, for multiple instances")
	rootCmd.PersistentFlags().BoolVar(&rootCmdArgs.Json, "json", rootCmdArgs.Json, "enable json log output")
}

func configureLogger(cmd *cobra.Command) zerolog.Logger {
	var w io.Writer

	if rootCmdArgs.Json {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
		w = os.Stdout
	} else {
		w = log.NewColorWriter(os.Stdout)
	}

	l := zerolog.New(w).With().Timestamp().Logger()
	if rootCmdArgs.Verbose {
		l = l.Level(zerolog.DebugLevel)
	} else {
		l = l.Level(zerolog.InfoLevel)
	}

	cmd.SetContext(l.WithContext(cmd.Context()))

	return l
}
