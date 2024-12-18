package cmd

import (
	"github.com/spf13/cobra"
	"ops_cli/cmd/check"
	"ops_cli/cmd/query"
	"ops_cli/internal/config"
	"ops_cli/pkg/log"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "ops_cli",
	Short: "Operations CLI tool for OPS component checks",
	Long:  `A command line tool for checking and managing OPS components.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := config.LoadConfig(cfgFile); err != nil {
			log.Error("Failed to load config: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SuggestionsMinimumDistance = 1
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	rootCmd.AddCommand(check.Cmd)
	rootCmd.AddCommand(query.Cmd)
}
