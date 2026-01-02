package cli

import (
	"fmt"
	"os"

	"github.com/mmdbasi/mtcli/internal/commands/history"
	"github.com/mmdbasi/mtcli/internal/commands/show"
	"github.com/mmdbasi/mtcli/internal/commands/stats"
	"github.com/mmdbasi/mtcli/internal/commands/test"
	"github.com/mmdbasi/mtcli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "mtcli",
		Short: "A terminal typing test inspired by Monkeytype",
		Long: `mtcli is a command-line typing test tool that helps you improve your typing speed.

It supports multiple modes:
  - Timer mode: Type as many words as you can in a set time
  - Words mode: Type a fixed number of words
  - Quote mode: Type famous quotes

Your results are saved locally so you can track your progress over time.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/mtcli/config.toml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable color output")

	// Add subcommands
	rootCmd.AddCommand(test.NewTestCmd())
	rootCmd.AddCommand(stats.NewStatsCmd())
	rootCmd.AddCommand(history.NewHistoryCmd())
	rootCmd.AddCommand(show.NewShowCmd())
}

func initConfig() {
	if cfgFile != "" {
		config.SetConfigFile(cfgFile)
	}

	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
	}
}

func Execute() error {
	return rootCmd.Execute()
}

