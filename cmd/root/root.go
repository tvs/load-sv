package root

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// TODO(tvs): Set up config management

		configureLogger(cmd)

		return nil
	},
}

// Cmd returns the root command
func Cmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command. This is called by main.main() and ony
// needs to be called once.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		l := zerolog.Ctx(rootCmd.Context())
		l.Error().Err(err).Msg("an error occurred during execution")
		return err
	}
	return nil
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

func configureLogger(cmd *cobra.Command) {
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
}
