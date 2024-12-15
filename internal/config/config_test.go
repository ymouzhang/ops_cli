package config

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	err := LoadConfig("/workspaces/ops_cli/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
}

func TestGetConfig(t *testing.T) {
	config := GetConfig()
	if config == nil {
		t.Fatalf("GetConfig returned nil")
	}

	if len(config.IPs) != 2 {
		t.Errorf("Expected 2 IPs, got %d", len(config.IPs))
	}
	level := "info"
	if config.Log.Level != level {
		t.Errorf("Expected log level %s, got %s", level, config.Log.Level)
	}

	fmt.Println(config.Log.Level)
}
