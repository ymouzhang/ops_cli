package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	IPs  []IPConfig `mapstructure:"ips"`
	Port PortConfig `mapstructure:"port"`
	Log  LogConfig  `mapstructure:"log"`
}

type IPConfig struct {
	IP       string `mapstructure:"ip"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	Role     string `mapstructure:"role"`
}

type PortConfig struct {
	Default PortDetail `mapstructure:"default"`
	Ops     PortDetail `mapstructure:"ops"`
}

type PortDetail struct {
	Prometheus  int `mapstructure:"prometheus"`
	Grafana     int `mapstructure:"grafana"`
	Pushgateway int `mapstructure:"pushgateway"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

var globalConfig Config

func LoadConfig(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(&globalConfig)
}

func GetConfig() *Config {
	return &globalConfig
}

// Component constants
const (
	ComponentPrometheus  = "prometheus"
	ComponentGrafana     = "grafana"
	ComponentPushgateway = "pushgateway"
)

// Path constants
const (
	PathQuery      = "query"
	PathQueryRange = "query_range"
	PathTargets    = "targets"
	PathHealth     = "health"
	PathFederate   = "federate"
)

// Role constants
const (
	RoleOps = "ops"
)

// ComponentConfig 定义组件配置
type ComponentConfig struct {
	Prefix string
	Paths  map[string]string
}

// componentConfigs 定义所有组件的配置
var componentConfigs = map[string]ComponentConfig{
	ComponentPrometheus: {
		Prefix: "/prometheus",
		Paths: map[string]string{
			PathQuery:      "/api/v1/query",
			PathQueryRange: "/api/v1/query_range",
			PathTargets:    "/api/v1/targets",
			PathHealth:     "/-/healthy",
			PathFederate:   "/federate",
		},
	},
	ComponentGrafana: {
		Prefix: "/grafana",
	},
	ComponentPushgateway: {
		Prefix: "/pushgateway",
	},
}

// URLBuilder 用于构建组件 URL
type URLBuilder struct {
	ip        string
	port      int
	role      string
	component string
	item      string
}

// Build 构建最终的 URL
func (b *URLBuilder) Build() (string, error) {
	// 获取端口
	port, err := GetPort(b.role, b.component)
	if err != nil {
		return "", err
	}
	b.port = port

	baseURL := fmt.Sprintf("http://%s:%d", b.ip, b.port)

	// 获取组件配置
	config, ok := componentConfigs[b.component]
	if !ok {
		return "", fmt.Errorf("unknown component: %s", b.component)
	}

	// 如果是 ops 角色
	if b.role == RoleOps {
		if path, ok := config.Paths[b.item]; ok {
			return baseURL + path, nil
		}
		return baseURL, nil
	}

	// 非 ops 角色，添加组件前缀
	if path, ok := config.Paths[b.item]; ok {
		return baseURL + config.Prefix + path, nil
	}
	return baseURL + config.Prefix, nil
}

// GetUrl 构建组件 URL
func GetUrl(ip string, role string, component string, item string) (string, error) {
	builder := &URLBuilder{
		ip:        ip,
		role:      role,
		component: component,
		item:      item,
	}
	return builder.Build()
}

func GetPort(role string, component string) (int, error) {
	if role == "ops" {
		switch component {
		case "prometheus":
			return globalConfig.Port.Ops.Prometheus, nil
		case "grafana":
			return globalConfig.Port.Ops.Grafana, nil
		case "pushgateway":
			return globalConfig.Port.Ops.Pushgateway, nil
		default:
			return 0, fmt.Errorf("unknown component: %s", component)
		}
	}

	// Default ports
	switch component {
	case "prometheus":
		return globalConfig.Port.Default.Prometheus, nil
	case "grafana":
		return globalConfig.Port.Default.Grafana, nil
	case "pushgateway":
		return globalConfig.Port.Default.Pushgateway, nil
	default:
		return 0, fmt.Errorf("unknown component: %s", component)
	}
}
