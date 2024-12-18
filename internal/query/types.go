package query

type PrometheusQuery struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

type QueryConfig struct {
	Port   int               `yaml:"port"`
	PromQL []PrometheusQuery `yaml:"promql"`
}

type QueryFile struct {
	Query struct {
		QueryTime string      `yaml:"query_time"`
		Ops       QueryConfig `yaml:"ops"`
		General   QueryConfig `yaml:"general"`
	} `yaml:"query"`
	QueryRange struct {
		Start   string      `yaml:"start"`
		End     string      `yaml:"end"`
		Ops     QueryConfig `yaml:"ops"`
		General QueryConfig `yaml:"general"`
	} `yaml:"query_range"`
}
