package query

import (
	"github.com/spf13/cobra"
	"ops_cli/internal/checker"
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

	var results []checker.CheckResult

	switch queryType {
	case "query":
		queryChecker := query.NewQueryChecker(cfg)
		results = queryChecker.Check()
	case "query_range":
		queryRangeChecker := query.NewQueryRangeChecker(cfg)
		results = queryRangeChecker.Check()
	default:
		cmd.Println("Invalid query type. Use 'query' or 'query_range'.")
		return
	}

	output.FormatCheckResults(results)
}
