package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch configuration.
type Config struct {
	Environments []Environment `yaml:"environments"`
	CheckInterval time.Duration `yaml:"check_interval"`
	Alerts        AlertConfig   `yaml:"alerts"`
}

// Environment describes a single monitored environment.
type Environment struct {
	Name     string            `yaml:"name"`
	Provider string            `yaml:"provider"` // e.g. "aws", "gcp", "static"
	Region   string            `yaml:"region,omitempty"`
	Tags     map[string]string `yaml:"tags,omitempty"`
}

// AlertConfig controls how drift alerts are emitted.
type AlertConfig struct {
	SlackWebhook string `yaml:"slack_webhook,omitempty"`
	LogOnly      bool   `yaml:"log_only"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 5 * time.Minute
	}

	return &cfg, nil
}

// Environment returns the environment with the given name, or an error if not found.
func (c *Config) Environment(name string) (*Environment, error) {
	for i := range c.Environments {
		if c.Environments[i].Name == name {
			return &c.Environments[i], nil
		}
	}
	return nil, fmt.Errorf("environment %q not found", name)
}

// EnvironmentNames returns a slice of all configured environment names.
func (c *Config) EnvironmentNames() []string {
	names := make([]string, len(c.Environments))
	for i, env := range c.Environments {
		names[i] = env.Name
	}
	return names
}

func (c *Config) validate() error {
	if len(c.Environments) == 0 {
		return fmt.Errorf("at least one environment must be defined")
	}
	seen := make(map[string]bool)
	for _, env := range c.Environments {
		if env.Name == "" {
			return fmt.Errorf("environment name must not be empty")
		}
		if seen[env.Name] {
			return fmt.Errorf("duplicate environment name %q", env.Name)
		}
		seen[env.Name] = true
		if env.Provider == "" {
			return fmt.Errorf("environment %q must specify a provider", env.Name)
		}
	}
	return nil
}
