package query

import (
	"github.com/spf13/cobra"
	"ops_cli/internal/config"
	"ops_cli/internal/query"
	"ops_cli/pkg/output"
)

// Cmd represents the query command
var Cmd = &cobra.Command{
	Use:   "query [flags]",
	Short: "Query Prometheus data",
	Long: `Query Prometheus data using the configurations defined in query.yaml:
- Query
- Query Range`,
	Run: runQuery,
}

func init() {
	Cmd.Flags().StringP("type", "t", "query", "Type of query to perform (query, query_range)")
}

func runQuery(cmd *cobra.Command, args []string) {
	queryType, _ := cmd.Flags().GetString("type")
	cfg := config.GetConfig()

	manager := query.NewManager(cfg)
	results := manager.Check(queryType)

	output.FormatCheckResults(results)
}
