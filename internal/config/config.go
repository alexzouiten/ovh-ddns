package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
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
	LogLevel   slog.Level    `yaml:"log_level"`
	DryRun     bool          `yaml:"dry_run"`
	Interval   time.Duration `yaml:"interval"`
	Timeout    time.Duration `yaml:"timeout"`
	IpProvider string        `yaml:"ip_provider"`
	Records    []Record      `yaml:"records"`
	Auth       Auth          `yaml:"auth"`
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
		return cfg, errors.New("config validation error")
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
	var errs []error

	errs = validateInterval(cfg.Interval, errs)
	errs = validateAuth(cfg.Auth, errs)
	errs = validateRecords(cfg.Records, errs)
	errs = validateIpProvider(cfg.IpProvider, errs)

	return errs
}

func validateRecords(records []Record, errs []error) []error {
	if len(records) == 0 {
		errs = append(errs, errors.New("need at least one configured record"))
	}

	return errs
}

func validateAuth(auth Auth, errs []error) []error {
	if auth.AppKey == "" {
		errs = append(errs, errors.New("app key cannot be empty"))
	}

	if auth.AppSecret == "" {
		errs = append(errs, errors.New("app secret cannot be empty"))
	}

	if auth.ConsumerKey == "" {
		errs = append(errs, errors.New("consumer key cannot be empty"))
	}

	return errs
}

func validateInterval(interval time.Duration, errs []error) []error {
	if interval < time.Minute {
		errs = append(errs, errors.New("update interval cannot be shorter than 1min"))
	}

	return errs
}

func validateIpProvider(ipProvider string, errs []error) []error {
	if ipProvider == "" {
		errs = append(errs, errors.New("IP provider URL cannot be empty"))
		return errs
	}

	if _, err := url.ParseRequestURI(ipProvider); err != nil {
		errs = append(errs, fmt.Errorf("'%s' is not a valid URL", ipProvider))
	}

	return errs
}
