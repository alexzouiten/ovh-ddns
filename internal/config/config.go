package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alexzouiten/ovh-ddns/internal/dns"
	"go.yaml.in/yaml/v3"
)

const (
	configPathEnvVar  = "OVH_DDNS_CONFIG_PATH"
	defaultConfigPath = "./config.yaml"
)

type Config struct {
	LogLevel slog.Level    `yaml:"log_level"`
	DryRun   bool          `yaml:"dry_run"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Records  []Record      `yaml:"records"`
	Auth     Auth          `yaml:"auth"`
}

type Record struct {
	Zone      string
	Subdomain string
	Types     []dns.RecordType
	Ttl       int
}

type Auth struct {
	AppKey      string `yaml:"app_key"`
	AppSecret   string `yaml:"app_secret"`
	ConsumerKey string `yaml:"consumer_key"`
}

func newConfig() Config {
	return Config{
		LogLevel: slog.LevelInfo,
		DryRun:   false,
		Interval: time.Minute * 5,
		Timeout:  time.Second * 30,
	}
}

func LoadConfig() (Config, error) {
	cfg := newConfig()

	yml, err := readConfigFile(configPath())
	if err != nil {
		return cfg, err
	}

	if err := parseYaml(&cfg, yml); err != nil {
		return cfg, err
	}

	if errs := validate(&cfg); len(errs) > 0 {
		for _, err := range errs {
			slog.Error("validation error", "error", err)
		}
		return cfg, errors.New("validation error")
	}

	return cfg, nil
}

func configPath() string {
	if yamlPathEnvVar := os.Getenv(configPathEnvVar); yamlPathEnvVar != "" {
		return yamlPathEnvVar
	}
	return defaultConfigPath
}

// readConfigFile parses the yaml config file and overwrite the default configuration with its content.
// If the config file doesn't exist, then the default config is returned.
func readConfigFile(yamlPath string) ([]byte, error) {
	f, err := os.Open(yamlPath)
	if err != nil {
		return nil, err
	}

	// read at most 1MB
	buf := make([]byte, 1_000_000)
	n, err := f.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s', %w", yamlPath, err)
	}
	return buf[:n], nil
}

func parseYaml(cfg *Config, buf []byte) error {
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return fmt.Errorf("failed to parse yaml config: %w", err)
	}
	return nil
}

func validate(cfg *Config) []error {
	var err []error

	err = validateInterval(cfg.Interval, err)
	err = validateAuth(cfg.Auth, err)
	err = validateRecords(cfg.Records, err)

	return err
}

func validateRecords(records []Record, err []error) []error {
	if len(records) == 0 {
		err = append(err, errors.New("need at least one configured domain"))
	}

	return err
}

func validateAuth(auth Auth, err []error) []error {
	if auth.AppKey == "" {
		err = append(err, errors.New("app key cannot be empty"))
	}

	if auth.AppSecret == "" {
		err = append(err, errors.New("app secret cannot be empty"))
	}

	if auth.ConsumerKey == "" {
		err = append(err, errors.New("consumer key cannot be empty"))
	}

	return err
}

func validateInterval(interval time.Duration, err []error) []error {
	if interval < time.Minute {
		err = append(err, errors.New("update interval cannot be shorter than 1min"))
	}

	return err
}
