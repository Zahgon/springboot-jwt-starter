// Package config loads the application settings that were previously supplied
// by Spring Boot's application.yml binding.
package config

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed application.yml
var applicationYAML []byte

// Config mirrors the key names in application.yml exactly.
type Config struct {
	App struct {
		Name string `yaml:"name"`
	} `yaml:"app"`
	JWT struct {
		Header string `yaml:"header"`
		// ExpiresIn is the access-token lifetime in seconds.
		ExpiresIn int `yaml:"expires_in"`
		// MobileExpiresIn is carried for fidelity with the original, which
		// declares it in configuration but never reads it.
		MobileExpiresIn int    `yaml:"mobile_expires_in"`
		Secret          string `yaml:"secret"`
	} `yaml:"jwt"`
}

// Load parses the embedded application.yml.
func Load() (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(applicationYAML, &c); err != nil {
		return nil, fmt.Errorf("config: parsing application.yml: %w", err)
	}
	return &c, nil
}
