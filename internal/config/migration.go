package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const currentConfigVersion = 2

type OldConfig struct {
	Version       int                    `yaml:"version"`
	ClientRules   map[string]string      `yaml:"client_rules"`
	Notifications map[string]interface{} `yaml:"notifications"`
}

type NewConfig struct {
	Version       int                    `yaml:"version"`
	ClientRules   []ClientRule           `yaml:"client_rules"`
	Notifications map[string]interface{} `yaml:"notifications"`
}

func MigrateConfig(oldConfigPath string, newConfigPath string) error {
	oldConfig, err := loadOldConfig(oldConfigPath)
	if err != nil {
		return err
	}

	newConfig := NewConfig{
		Version:       currentConfigVersion,
		ClientRules:   migrateClientRules(oldConfig.ClientRules),
		Notifications: oldConfig.Notifications,
	}

	return saveNewConfig(newConfigPath, newConfig)
}

func loadOldConfig(path string) (OldConfig, error) {
	var config OldConfig
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read old config: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to unmarshal old config: %w", err)
	}

	if config.Version < 1 || config.Version > currentConfigVersion {
		return config, fmt.Errorf("unsupported config version: %d", config.Version)
	}

	return config, nil
}

func migrateClientRules(oldRules map[string]string) []ClientRule {
	var newRules []ClientRule
	for appName, inputMethod := range oldRules {
		newRules = append(newRules, ClientRule{
			Class:       appName,
			Title:       "",
			InputMethod: inputMethod,
		})
	}
	return newRules
}

func saveNewConfig(path string, config NewConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal new config: %w", err)
	}

	if err := ioutil.WriteFile(path, data, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write new config: %w", err)
	}

	return nil
}
