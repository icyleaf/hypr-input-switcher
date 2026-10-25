package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const currentConfigVersion = 3

type OldConfig struct {
	Version            int                `yaml:"version"`
	Description        string             `yaml:"description"`
	DefaultInputMethod string             `yaml:"default_input_method"`
	InputMethods       interface{}        `yaml:"input_methods"`
	ClientRules        interface{}        `yaml:"client_rules"`
	LayerRules         []LayerRule        `yaml:"layer_rules"`
	Fcitx5             Fcitx5Config       `yaml:"fcitx5"`
	RimeSchemas        map[string]string  `yaml:"rime_schemas"`
	Notifications      NotificationConfig `yaml:"notifications"`
	DisplayNames       map[string]string  `yaml:"display_names"`
	Icons              map[string]string  `yaml:"icons"`
}

func MigrateConfig(oldConfigPath string, newConfigPath string) error {
	oldConfig, err := loadOldConfig(oldConfigPath)
	if err != nil {
		return err
	}

	cfg := Config{
		Version:            currentConfigVersion,
		Description:        oldConfig.Description,
		DefaultInputMethod: oldConfig.DefaultInputMethod,
		LayerRules:         oldConfig.LayerRules,
		Notifications:      oldConfig.Notifications,
		DisplayNames:       oldConfig.DisplayNames,
		Icons:              oldConfig.Icons,
		RimeSchemas:        oldConfig.RimeSchemas,
		Fcitx5:             oldConfig.Fcitx5,
		InputMethods:       make(map[string]InputMethodConfig),
	}
	if cfg.Description == "" {
		cfg.Description = "Hyprland Input Method Switcher Configuration"
	}
	if cfg.DefaultInputMethod == "" {
		cfg.DefaultInputMethod = KeepInputMethod
	}

	// Handle ClientRules: map[string]string (v1) or []ClientRule (v2/v3)
	switch rules := oldConfig.ClientRules.(type) {
	case map[interface{}]interface{}:
		for k, v := range rules {
			cfg.ClientRules = append(cfg.ClientRules, ClientRule{
				Class:       fmt.Sprintf("%v", k),
				InputMethod: fmt.Sprintf("%v", v),
			})
		}
	case []interface{}:
		for _, item := range rules {
			if m, ok := item.(map[interface{}]interface{}); ok {
				rule := ClientRule{}
				if c, exists := m["class"]; exists {
					rule.Class = fmt.Sprintf("%v", c)
				}
				if t, exists := m["title"]; exists {
					rule.Title = fmt.Sprintf("%v", t)
				}
				if im, exists := m["input_method"]; exists {
					rule.InputMethod = fmt.Sprintf("%v", im)
				}
				cfg.ClientRules = append(cfg.ClientRules, rule)
			}
		}
	}

	// Handle InputMethods: map[string]string (v2) or map[string]InputMethodConfig (v3)
	if ims, ok := oldConfig.InputMethods.(map[interface{}]interface{}); ok {
		for k, v := range ims {
			label := fmt.Sprintf("%v", k)
			switch val := v.(type) {
			case string:
				imCfg := InputMethodConfig{}
				_ = imCfg.UnmarshalYAML(func(target interface{}) error {
					*(target.(*string)) = val
					return nil
				})
				cfg.InputMethods[label] = imCfg
			case map[interface{}]interface{}:
				imCfg := InputMethodConfig{}
				if eng, exists := val["engine"]; exists {
					imCfg.Engine = fmt.Sprintf("%v", eng)
				}
				if prof, exists := val["profile"]; exists {
					imCfg.Profile = fmt.Sprintf("%v", prof)
				}
				if dn, exists := val["display_name"]; exists {
					imCfg.DisplayName = fmt.Sprintf("%v", dn)
				}
				if ic, exists := val["icon"]; exists {
					imCfg.Icon = fmt.Sprintf("%v", ic)
				}
				cfg.InputMethods[label] = imCfg
			}
		}
	}

	// For v1 configs lacking an input_methods section, synthesize entries from references
	if len(cfg.InputMethods) == 0 {
		ensureMethod := func(label string) {
			if label == "" || label == KeepInputMethod {
				return
			}
			if _, exists := cfg.InputMethods[label]; !exists {
				im := InputMethodConfig{}
				if label == "english" {
					im.Engine = "keyboard-us"
				} else if schema, ok := cfg.RimeSchemas[label]; ok {
					im.Engine = "rime"
					im.Profile = schema
				} else {
					im.Engine = label
				}
				cfg.InputMethods[label] = im
			}
		}

		ensureMethod(cfg.DefaultInputMethod)
		for _, rule := range cfg.ClientRules {
			ensureMethod(rule.InputMethod)
		}
		for _, rule := range cfg.LayerRules {
			ensureMethod(rule.InputMethod)
		}
	}

	// Normalize into cohesive v3 entities
	cfg.Normalize()

	// Clear out deprecated top-level maps now that they are inlined into InputMethods
	cfg.RimeSchemas = nil
	cfg.DisplayNames = nil
	cfg.Icons = nil

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal v3 config: %w", err)
	}

	if err := os.WriteFile(newConfigPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write v3 config: %w", err)
	}

	return nil
}

func loadOldConfig(path string) (OldConfig, error) {
	var config OldConfig
	data, err := os.ReadFile(path)
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
