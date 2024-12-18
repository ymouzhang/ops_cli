package query

import (
	"gopkg.in/yaml.v3"
	"ops_cli/pkg/log"
	"os"
)

func loadQueries(filePath, section, role string) (int, []PrometheusQuery, string, string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to open query.yaml: %v", err)
		return 0, nil, "", ""
	}
	defer file.Close()

	var data QueryFile

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		log.Error("Failed to decode query.yaml: %v", err)
		return 0, nil, "", ""
	}

	if section == "query" {
		if role == "ops" {
			return data.Query.Ops.Port, data.Query.Ops.PromQL, data.Query.QueryTime, ""
		} else if role == "general" {
			return data.Query.General.Port, data.Query.General.PromQL, data.Query.QueryTime, ""
		}
	} else if section == "query_range" {
		if role == "ops" {
			return data.QueryRange.Ops.Port, data.QueryRange.Ops.PromQL, data.QueryRange.Start, data.QueryRange.End
		} else if role == "general" {
			return data.QueryRange.General.Port, data.QueryRange.General.PromQL, data.QueryRange.Start, data.QueryRange.End
		}
	}

	return 0, nil, "", ""
}
