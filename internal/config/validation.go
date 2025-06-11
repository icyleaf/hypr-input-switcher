package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("configuration file does not exist")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Version <= 0 {
		return errors.New("invalid configuration version")
	}
	if config.Description == "" {
		return errors.New("description cannot be empty")
	}
	if len(config.InputMethods) == 0 {
		return errors.New("input methods cannot be empty")
	}
	if config.DefaultInputMethod == "" {
		return errors.New("default input method cannot be empty")
	}
	if len(config.ClientRules) == 0 {
		return errors.New("client rules cannot be empty")
	}
	return nil
}
