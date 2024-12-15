package check

import (
	"ops_cli/internal/checker"
	"ops_cli/internal/config"
	"ops_cli/pkg/output"

	"github.com/spf13/cobra"
)

// Cmd represents the check command
var Cmd = &cobra.Command{
	Use:   "check [flags]",
	Short: "Check system components",
	Long: `Check the status of various system components including:
- SSH connections to remote hosts
- Prometheus services`,
	Run: runCheck,
}

func init() {
	Cmd.Flags().StringP("component", "c", "", "Component to check (prometheus, system, ssh, all)")
}

func runCheck(cmd *cobra.Command, args []string) {
	component, _ := cmd.Flags().GetString("component")
	cfg := config.GetConfig()

	checkMgr := checker.NewManager(cfg)

	results := checkMgr.Check(component)

	output.FormatCheckResults(results)
}
