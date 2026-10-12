package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

const (
	configPathEnvVar = "OVH_DDNS_CONFIG_PATH"
)

var (
	// create a logger just to get the same formatting as the rest of the program
	// when logging that we parse the yaml config in readConfigFile()
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
)

type Auth struct {
	AppKey      string `yaml:"app_key"`
	AppSecret   string `yaml:"app_secret"`
	ConsumerKey string `yaml:"consumer_key"`
}

type Update struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

type Config struct {
	LogLevel slog.Level `yaml:"log_level"`
	DryRun   bool       `yaml:"dry_run"`
	Domains  []string   `yaml:"domains"`
	Auth     Auth       `yaml:"auth"`
	Update   Update     `yaml:"update"`
}

func newConfig() Config {
	return Config{
		LogLevel: slog.LevelInfo,
		DryRun:   false,
		Update: Update{
			Interval: time.Minute * 5,
			Timeout:  time.Second * 30,
		},
	}
}

func LoadConfig() (Config, error) {
	cfg := newConfig()

	if err := readConfigFile(&cfg, configPath()); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func configPath() string {
	yamlPath := "./config.yaml"

	if yamlPathEnvVar := os.Getenv(configPathEnvVar); yamlPathEnvVar != "" {
		yamlPath = yamlPathEnvVar
	}

	return yamlPath
}

// readConfigFile parses the yaml config file and overwrite the default configuration with its content.
// If the config file doesn't exist, then the default config is returned.
func readConfigFile(cfg *Config, yamlPath string) error {
	logger.Info("loading yaml configuration", "path", yamlPath)

	f, err := os.Open(yamlPath)
	if err != nil {
		return err
	}

	// read at most 1MB
	buf := make([]byte, 1_000_000)
	n, err := f.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read config file '%s', %w", yamlPath, err)
	}

	if err := yaml.Unmarshal(buf[:n], &cfg); err != nil {
		return fmt.Errorf("failed to parse yaml config: %w", err)
	}

	if errs := validate(cfg); len(errs) > 0 {
		for _, err := range errs {
			slog.Error("validation error", "error", err)
		}
		return errors.New("validation error")
	}

	return nil
}

func validate(cfg *Config) []error {
	var err []error

	err = validateDomains(cfg.Domains, err)
	err = validateAuth(cfg.Auth, err)
	err = validateUpdate(cfg.Update, err)

	return err
}

func validateDomains(domains []string, err []error) []error {
	if len(domains) == 0 {
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

func validateUpdate(update Update, err []error) []error {
	if update.Interval < time.Minute {
		err = append(err, errors.New("update interval cannot be shorter than 1min"))
	}

	return err
}
