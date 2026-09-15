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
