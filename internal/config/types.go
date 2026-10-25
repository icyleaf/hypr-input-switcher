package config

import "strings"

// KeepInputMethod is a reserved input method value that disables switching.
const KeepInputMethod = "keep"

// InputMethodConfig represents configuration for a specific input method entry
type InputMethodConfig struct {
	Engine      string `yaml:"engine" json:"engine"`
	Profile     string `yaml:"profile,omitempty" json:"profile,omitempty"`
	DisplayName string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Icon        string `yaml:"icon,omitempty" json:"icon,omitempty"`
}

// UnmarshalYAML supports both shorthand string ("engine", "engine:profile") and map structure
func (c *InputMethodConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err == nil {
		parts := strings.SplitN(str, ":", 2)
		c.Engine = parts[0]
		if len(parts) == 2 {
			c.Profile = parts[1]
		}
		return nil
	}

	type rawInputMethodConfig InputMethodConfig
	var raw rawInputMethodConfig
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*c = InputMethodConfig(raw)
	return nil
}

// Config represents the application configuration
type Config struct {
	Version            int                          `yaml:"version" json:"version"`
	Description        string                       `yaml:"description" json:"description"`
	DefaultInputMethod string                       `yaml:"default_input_method" json:"default_input_method"`
	InputMethods       map[string]InputMethodConfig `yaml:"input_methods" json:"input_methods"`
	ClientRules        []ClientRule                 `yaml:"client_rules,omitempty" json:"client_rules,omitempty"`
	LayerRules         []LayerRule                  `yaml:"layer_rules,omitempty" json:"layer_rules,omitempty"`
	Fcitx5             *Fcitx5Config                `yaml:"fcitx5,omitempty" json:"fcitx5,omitempty"`
	RimeSchemas        map[string]string            `yaml:"rime_schemas,omitempty" json:"rime_schemas,omitempty"`
	Notifications      NotificationConfig           `yaml:"notifications" json:"notifications"`
	DisplayNames       map[string]string            `yaml:"display_names,omitempty" json:"display_names,omitempty"`
	Icons              map[string]string            `yaml:"icons,omitempty" json:"icons,omitempty"`
}

// ClientRule represents a client-specific input method rule
type ClientRule struct {
	Class       string `yaml:"class" json:"class"`
	Title       string `yaml:"title,omitempty" json:"title,omitempty"`
	InputMethod string `yaml:"input_method" json:"input_method"`
}

// LayerRule represents a layer-shell namespace specific input method rule.
// Layer surfaces (Omarchy overlay plugins) are not windows and can only be
// matched by their namespace, not by class or title.
type LayerRule struct {
	Namespace   string `yaml:"namespace" json:"namespace"`
	InputMethod string `yaml:"input_method" json:"input_method"`
}

// Fcitx5Config represents fcitx5 configuration
type Fcitx5Config struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	RimeInputMethod string `yaml:"rime_input_method,omitempty" json:"rime_input_method,omitempty"`
	RimeConfigDir   string `yaml:"rime_config_dir,omitempty" json:"rime_config_dir,omitempty"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Duration        int      `yaml:"duration"`
	ShowOnSwitch    bool     `json:"show_on_switch" yaml:"show_on_switch"`
	ShowAppName     bool     `json:"show_app_name" yaml:"show_app_name"`
	Methods         []string `yaml:"methods"`
	DisabledMethods []string `yaml:"disabled_methods"`
	ForceMethod     string   `yaml:"force_method"`
	IconPath        string   `yaml:"icon_path"`
}

// WindowInfo represents active window information
type WindowInfo struct {
	Class      string `json:"class"`
	Title      string `json:"title"`
	IsKeep     bool   `json:"is_keep,omitempty"`
	IsRestored bool   `json:"is_restored,omitempty"`
}
