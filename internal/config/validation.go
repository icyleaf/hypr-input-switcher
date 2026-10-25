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

	config.Normalize()

	return &config, nil
}

// Normalize populates backward-compatible fields between input_methods and legacy maps
func (c *Config) Normalize() {
	if c.Fcitx5.RimeInputMethod == "" {
		c.Fcitx5.RimeInputMethod = "rime"
	}
	if c.DisplayNames == nil {
		c.DisplayNames = make(map[string]string)
	}
	if c.Icons == nil {
		c.Icons = make(map[string]string)
	}
	if c.RimeSchemas == nil {
		c.RimeSchemas = make(map[string]string)
	}

	for label, im := range c.InputMethods {
		// Populate profile from legacy RimeSchemas if not specified
		if im.Profile == "" && c.RimeSchemas[label] != "" {
			im.Profile = c.RimeSchemas[label]
		}
		if im.Profile != "" && c.RimeSchemas[label] == "" {
			c.RimeSchemas[label] = im.Profile
		}

		// Sync DisplayName
		if im.DisplayName == "" && c.DisplayNames[label] != "" {
			im.DisplayName = c.DisplayNames[label]
		}
		if im.DisplayName != "" && c.DisplayNames[label] == "" {
			c.DisplayNames[label] = im.DisplayName
		}

		// Sync Icon
		if im.Icon == "" && c.Icons[label] != "" {
			im.Icon = c.Icons[label]
		}
		if im.Icon != "" && c.Icons[label] == "" {
			c.Icons[label] = im.Icon
		}

		c.InputMethods[label] = im
	}
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
	if len(config.ClientRules) == 0 && len(config.LayerRules) == 0 {
		return errors.New("at least one of client rules or layer rules must be defined")
	}
	return nil
}
