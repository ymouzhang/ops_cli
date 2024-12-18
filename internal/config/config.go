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
