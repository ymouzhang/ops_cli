package query

import (
	"github.com/spf13/viper"
)

type PrometheusQuery struct {
	Name  string `mapstructure:"name"`
	Query string `mapstructure:"query"`
}

type QueryConfig struct {
	Port   int               `mapstructure:"port"`
	PromQL []PrometheusQuery `mapstructure:"promql"`
}

type Config struct {
	Query struct {
		QueryTime string      `mapstructure:"query_time"`
		Ops       QueryConfig `mapstructure:"ops"`
		General   QueryConfig `mapstructure:"general"`
	} `mapstructure:"query"`
	QueryRange struct {
		Start   string      `mapstructure:"start"`
		End     string      `mapstructure:"end"`
		Ops     QueryConfig `mapstructure:"ops"`
		General QueryConfig `mapstructure:"general"`
	} `mapstructure:"query_range"`
}

var globalConfig Config

func LoadConfig(queryConfig string) error {
	v := viper.New()

	if queryConfig != "" {
		v.SetConfigFile(queryConfig)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName("query")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(&globalConfig)
}

func GetConfig() *Config {
	return &globalConfig
}

// Helper functions to get specific configurations
func GetQueryConfig(role string) (QueryConfig, string) {
	if role == "ops" {
		return globalConfig.Query.Ops, globalConfig.Query.QueryTime
	}
	return globalConfig.Query.General, globalConfig.Query.QueryTime
}

func GetQueryRangeConfig(role string) (QueryConfig, string, string) {
	if role == "ops" {
		return globalConfig.QueryRange.Ops, globalConfig.QueryRange.Start, globalConfig.QueryRange.End
	}
	return globalConfig.QueryRange.General, globalConfig.QueryRange.Start, globalConfig.QueryRange.End
}

func loadQueries(section, role string) ([]PrometheusQuery, string, string) {
	if section == "query" {
		cfg, queryTime := GetQueryConfig(role)
		return cfg.PromQL, queryTime, ""
	} else if section == "query_range" {
		cfg, start, end := GetQueryRangeConfig(role)
		return cfg.PromQL, start, end
	}
	return nil, "", ""
}
