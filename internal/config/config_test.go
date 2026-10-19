package config

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/alexzouiten/ovh-ddns/internal/dns"
)

func TestLoadConfig(t *testing.T) {
	yaml := `---
log_level: debug
dry_run: true
interval: 10m
timeout: 10s
ip_provider: https://ipprovider.com
records:
  - zone: mydomain.com
    subdomain: sub
    types: ["A", "AAAA"]
    ttl: 4
auth:
  app_key: test-app-key
  app_secret: test-app-secret
  consumer_key: test-consumer-key
`

	f, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatalf("failed to create config temp file: %v", err)
	}
	defer os.Remove(f.Name())

	if err := os.Setenv(configPathEnvVar, f.Name()); err != nil {
		t.Fatalf("failed to set config path environment variable: %v", err)
	}
	defer os.Unsetenv(configPathEnvVar)

	if _, err := f.Write([]byte(yaml)); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	expected := Config{
		LogLevel:   slog.LevelDebug,
		DryRun:     true,
		Interval:   time.Minute * 10,
		Timeout:    time.Second * 10,
		IpProvider: "https://ipprovider.com",
		Records: []Record{
			{
				Zone:      "mydomain.com",
				Subdomain: "sub",
				Types:     []dns.RecordType{dns.RecordA, dns.RecordAAAA},
				Ttl:       4,
			},
		},
		Auth: Auth{
			AppKey:      "test-app-key",
			AppSecret:   "test-app-secret",
			ConsumerKey: "test-consumer-key",
		},
	}

	actual, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if actual.LogLevel != expected.LogLevel {
		t.Errorf("expected debug log level, got %s", actual.LogLevel)
	}

	if actual.DryRun != expected.DryRun {
		t.Errorf("expected dry_run to be true, got false")
	}

	if actual.Interval != expected.Interval {
		t.Errorf("expected update interval of %s, got %s", expected.Interval, actual.Interval)
	}

	if actual.Timeout != expected.Timeout {
		t.Errorf("expected update timeout of %s, got %s", expected.Timeout, actual.Timeout)
	}

	if actual.IpProvider != expected.IpProvider {
		t.Errorf("expected IP provider URL '%s', got '%s'", expected.IpProvider, actual.IpProvider)
	}

	if actual.Auth.AppKey != expected.Auth.AppKey {
		t.Errorf("expected app key '%s', got '%s'", expected.Auth.AppKey, actual.Auth.AppKey)
	}

	if actual.Auth.AppSecret != expected.Auth.AppSecret {
		t.Errorf("expected app secret '%s', got '%s'", expected.Auth.AppSecret, actual.Auth.AppSecret)
	}

	if actual.Auth.ConsumerKey != expected.Auth.ConsumerKey {
		t.Errorf("expected consumer key '%s', got '%s'", expected.Auth.ConsumerKey, actual.Auth.ConsumerKey)
	}

	expectedRecord := expected.Records[0]
	actualRecord := actual.Records[0]

	if actualRecord.Zone != expectedRecord.Zone {
		t.Errorf("expected dns zone '%s', got '%s'", expectedRecord.Zone, actualRecord.Zone)
	}

	if actualRecord.Subdomain != expectedRecord.Subdomain {
		t.Errorf("expected dns subdomain '%s', got '%s'", expectedRecord.Subdomain, actualRecord.Subdomain)
	}

	if actualRecord.Types[0] != expectedRecord.Types[0] && actualRecord.Types[1] != expectedRecord.Types[1] {
		t.Errorf("expected dns record types %v, got %v", expectedRecord.Types, actualRecord.Types)
	}

	if actualRecord.Ttl != expectedRecord.Ttl {
		t.Errorf("expected dns ttl %d, got %d", expectedRecord.Ttl, actualRecord.Ttl)
	}
}

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

func TestValidateIpProvider_Success(t *testing.T) {
	var actual []error
	actual = validateIpProvider("https://validwebsite.com", actual)
	if len(actual) != 0 {
		t.Errorf("expected no error, got %v", actual)
	}
}

func TestValidateIpProvider_Empty(t *testing.T) {
	var actual []error
	actual = validateIpProvider("", actual)
	if len(actual) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(actual), actual)
	}
}

func TestValidateIpProvider_InvalidUrl(t *testing.T) {
	var actual []error
	actual = validateIpProvider("randomtld", actual)
	if len(actual) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(actual), actual)
	}
}
