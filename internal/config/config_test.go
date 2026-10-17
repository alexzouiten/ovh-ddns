package config

import (
	"os"
	"testing"
)

func TestConfigPath_DefaultValue(t *testing.T) {
	oldConfigPathEnvVar := os.Getenv(configPathEnvVar)
	os.Unsetenv(configPathEnvVar)

	actual := configPath()
	if actual != defaultConfigPath {
		t.Errorf("expected default config path, got '%s'", actual)
	}

	if oldConfigPathEnvVar != "" {
		os.Setenv(configPathEnvVar, oldConfigPathEnvVar)
	}
}

func TestConfigPath_ValueFromEnv(t *testing.T) {
	oldConfigPathEnvVar := os.Getenv(configPathEnvVar)

	expectedConfigPath := "test.yaml"
	os.Setenv(configPathEnvVar, expectedConfigPath)

	actual := configPath()
	if actual != expectedConfigPath {
		t.Errorf("expected config path '%s', got '%s'", expectedConfigPath, actual)
	}

	if oldConfigPathEnvVar == "" {
		os.Unsetenv(configPathEnvVar)
	} else {
		os.Setenv(configPathEnvVar, oldConfigPathEnvVar)
	}
}
