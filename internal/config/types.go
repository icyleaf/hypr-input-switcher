package config

// Config represents the application configuration
type Config struct {
	Version            int                `yaml:"version" json:"version"`
	Description        string             `yaml:"description" json:"description"`
	DefaultInputMethod string             `yaml:"default_input_method" json:"default_input_method"`
	InputMethods       map[string]string  `yaml:"input_methods" json:"input_methods"`
	ClientRules        []ClientRule       `yaml:"client_rules" json:"client_rules"`
	Fcitx5             Fcitx5Config       `yaml:"fcitx5" json:"fcitx5"`
	RimeSchemas        map[string]string  `yaml:"rime_schemas" json:"rime_schemas"`
	Notifications      NotificationConfig `yaml:"notifications" json:"notifications"`
	DisplayNames       map[string]string  `yaml:"display_names" json:"display_names"`
	Icons              map[string]string  `yaml:"icons" json:"icons"`
}

// ClientRule represents a client-specific input method rule
type ClientRule struct {
	Class       string `yaml:"class" json:"class"`
	Title       string `yaml:"title" json:"title"`
	InputMethod string `yaml:"input_method" json:"input_method"`
}

// Fcitx5Config represents fcitx5 configuration
type Fcitx5Config struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	RimeInputMethod string `yaml:"rime_input_method" json:"rime_input_method"`
	RimeConfigDir   string `yaml:"rime_config_dir" json:"rime_config_dir"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	Duration        int      `json:"duration" yaml:"duration"`
	ShowOnSwitch    bool     `json:"show_on_switch" yaml:"show_on_switch"`
	ShowAppName     bool     `json:"show_app_name" yaml:"show_app_name"`
	Methods         []string `json:"methods" yaml:"methods"`
	ForceMethod     string   `json:"force_method" yaml:"force_method"`
	DisabledMethods []string `json:"disabled_methods" yaml:"disabled_methods"`
}

// WindowInfo represents active window information
type WindowInfo struct {
	Class string `json:"class"`
	Title string `json:"title"`
}
