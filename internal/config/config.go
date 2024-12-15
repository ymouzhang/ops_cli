package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	IPs []IPConfig `mapstructure:"ips"`
	Log LogConfig  `mapstructure:"log"`
}

type IPConfig struct {
	IP       string `mapstructure:"ip"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	Role     string `mapstructure:"role"`
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
	defaultPorts := map[string]int{
		"prometheus":  9090,
		"grafana":     3000,
		"pushgateway": 9091,
	}

	if role == "ops" {
		switch component {
		case "prometheus":
			return 32190, nil
		case "grafana":
			return 32123, nil
		case "pushgateway":
			return 32191, nil
		}
	}

	if port, exists := defaultPorts[component]; exists {
		return port, nil
	}

	return 0, fmt.Errorf("unknown component: %s", component)
}
