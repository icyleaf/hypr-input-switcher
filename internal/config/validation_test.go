package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

const baseConfig = `
version: 2
description: test
default_input_method: english
input_methods:
  english: English
`

func TestLoadConfigAcceptsLayerRulesOnly(t *testing.T) {
	path := writeConfig(t, baseConfig+`
layer_rules:
  - namespace: icyleaf-calculator
    input_method: english
  - namespace: omarchy-emojis
    input_method: english
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg.LayerRules) != 2 {
		t.Fatalf("LayerRules length = %d, want 2", len(cfg.LayerRules))
	}
	if cfg.LayerRules[0].Namespace != "icyleaf-calculator" {
		t.Fatalf("LayerRules[0].Namespace = %q, want %q", cfg.LayerRules[0].Namespace, "icyleaf-calculator")
	}
	if cfg.LayerRules[0].InputMethod != "english" {
		t.Fatalf("LayerRules[0].InputMethod = %q, want %q", cfg.LayerRules[0].InputMethod, "english")
	}
}

func TestLoadConfigAcceptsClientRulesOnly(t *testing.T) {
	path := writeConfig(t, baseConfig+`
client_rules:
  - class: kitty
    input_method: english
`)

	if _, err := LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
}

func TestLoadConfigRejectsConfigWithoutAnyRules(t *testing.T) {
	path := writeConfig(t, baseConfig)

	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig() error = nil, want error when neither client nor layer rules are defined")
	}
}

func TestLoadConfigInputMethodFormatsAndNormalize(t *testing.T) {
	configYaml := `
version: 2
description: test
default_input_method: english
input_methods:
  english: keyboard-us
  chinese: msime
  japanese: rime:jaroomaji
  korean:
    engine: hangul
    profile: 2bulsik
    display_name: 한국어
    icon: 🇰🇷
rime_schemas:
  chinese: rime_pinyin
display_names:
  english: English
icons:
  english: 🇺🇸
client_rules:
  - class: ghostty
    input_method: english
`
	path := writeConfig(t, configYaml)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// 1. Shorthand string "keyboard-us"
	if cfg.InputMethods["english"].Engine != "keyboard-us" {
		t.Fatalf("english engine = %q, want %q", cfg.InputMethods["english"].Engine, "keyboard-us")
	}
	// Legacy display_names/icons merged into english
	if cfg.InputMethods["english"].DisplayName != "English" {
		t.Fatalf("english display_name = %q, want %q", cfg.InputMethods["english"].DisplayName, "English")
	}
	if cfg.InputMethods["english"].Icon != "🇺🇸" {
		t.Fatalf("english icon = %q, want %q", cfg.InputMethods["english"].Icon, "🇺🇸")
	}

	// 2. rime_schemas merged into chinese profile
	if cfg.InputMethods["chinese"].Profile != "rime_pinyin" {
		t.Fatalf("chinese profile = %q, want %q", cfg.InputMethods["chinese"].Profile, "rime_pinyin")
	}

	// 3. Shorthand "rime:jaroomaji" parsed
	if cfg.InputMethods["japanese"].Engine != "rime" || cfg.InputMethods["japanese"].Profile != "jaroomaji" {
		t.Fatalf("japanese = %+v, want engine=rime, profile=jaroomaji", cfg.InputMethods["japanese"])
	}

	// 4. Object parsed and synced back to DisplayNames/Icons
	if cfg.DisplayNames["korean"] != "한국어" {
		t.Fatalf("korean display_name sync = %q, want %q", cfg.DisplayNames["korean"], "한국어")
	}
	if cfg.Icons["korean"] != "🇰🇷" {
		t.Fatalf("korean icon sync = %q, want %q", cfg.Icons["korean"], "🇰🇷")
	}
}
