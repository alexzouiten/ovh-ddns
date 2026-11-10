package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/alexzouiten/ovh-ddns/internal/model"
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
	Types     []model.RecordType
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
		return cfg, fmt.Errorf("failed to read config file '%s': %w", configPath(), err)
	}

	if err := yaml.Unmarshal(yml, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse yaml config: %w", err)
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

func readConfigFile(yamlPath string) ([]byte, error) {
	f, err := os.Open(yamlPath)
	if err != nil {
		return nil, err
	}

	// read at most 1MB
	buf := make([]byte, 1_000_000)
	n, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
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
		errs = append(errs, ErrNoRecord)
	}

	for _, record := range records {
		errs = validateRecord(record, errs)
	}

	return errs
}

func validateRecord(record Record, errs []error) []error {
	if record.Ttl < 0 {
		errs = append(errs, ErrTtlNegative)
	}

	errs = validateRecordTypes(record.Types, errs)

	return errs
}

func validateRecordTypes(recordTypes []model.RecordType, errs []error) []error {
	aCount, aaaaCount := 0, 0

	for _, rt := range recordTypes {
		switch rt {
		case model.RecordA:
			aCount++
		case model.RecordAAAA:
			aaaaCount++
		default:
			panic("unexpected record type")
		}
	}

	if aCount > 1 || aaaaCount > 1 {
		errs = append(errs, ErrRecordTypeDuplicated)
	}

	return errs
}

func validateAuth(auth Auth, errs []error) []error {
	if auth.AppKey == "" {
		errs = append(errs, ErrAppKeyEmpty)
	}

	if auth.AppSecret == "" {
		errs = append(errs, ErrAppSecretEmpty)
	}

	if auth.ConsumerKey == "" {
		errs = append(errs, ErrConsumerKeyEmpty)
	}

	return errs
}

func validateInterval(interval time.Duration, errs []error) []error {
	if interval < time.Minute {
		errs = append(errs, ErrUpdateIntervalTooShort)
	}

	return errs
}

func validateIpProvider(ipProvider string, errs []error) []error {
	if ipProvider == "" {
		errs = append(errs, ErrIpProviderUrlEmpty)
		return errs
	}

	if _, err := url.ParseRequestURI(ipProvider); err != nil {
		errs = append(errs, fmt.Errorf("'%s' is not a valid URL", ipProvider))
	}

	return errs
}
