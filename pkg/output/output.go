package output

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"ops_cli/internal/checker"
	"os"
)

// 抽取公共的表格配置函数
func configureTable(table *tablewriter.Table, withColor bool) {
	// 设置表头
	table.SetHeader([]string{"Component", "Role", "IP", "Item", "Status", "Message"})

	// 设置表格样式
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	if withColor {
		// 设置表头颜色
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		)

		// 设置列颜色
		table.SetColumnColor(
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.FgGreenColor},
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
		)
	}
}

// 抽取公共的添加数据行函数
func addTableRows(table *tablewriter.Table, results []checker.CheckResult, withColor bool) {
	for _, result := range results {
		message := result.Message
		if result.Error != nil {
			if message != "" {
				message += ": "
			}
			message += result.Error.Error()
		}

		row := []string{
			result.Component,
			result.Role,
			result.IP,
			result.Item,
			result.Status,
			message,
		}

		if withColor && result.Status == "Failed" {
			table.Rich(row, []tablewriter.Colors{
				{}, {}, {}, {},
				{tablewriter.FgRedColor},
				{},
			})
		} else {
			table.Append(row)
		}
	}
}

// 渲染表格到指定的writer
func renderTable(w io.Writer, results []checker.CheckResult, withColor bool) {
	table := tablewriter.NewWriter(w)
	configureTable(table, withColor)
	addTableRows(table, results, withColor)

	fmt.Fprintln(w, "\nCheck Results:\n")
	table.Render()
	fmt.Fprintln(w)
}

func FormatCheckResults(results []checker.CheckResult) {
	renderTable(os.Stdout, results, true)
}

func FormatCheckResultsToFile(results []checker.CheckResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	renderTable(file, results, false)
	return nil
}
