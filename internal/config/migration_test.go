package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateConfigV1ToV3(t *testing.T) {
	v1Yaml := `version: 1
description: "v1 config"
default_input_method: "english"
fcitx5:
  enabled: true
  rime_input_method: "rime"
rime_schemas:
  japanese: "jaroomaji"
display_names:
  english: "English"
  japanese: "日本語"
icons:
  english: "🇺🇸"
client_rules:
  kitty: "english"
  firefox: "japanese"
`
	dir := t.TempDir()
	v1Path := filepath.Join(dir, "v1.yaml")
	v3Path := filepath.Join(dir, "v3.yaml")

	if err := os.WriteFile(v1Path, []byte(v1Yaml), 0o644); err != nil {
		t.Fatalf("failed to write v1 config: %v", err)
	}

	if err := MigrateConfig(v1Path, v3Path); err != nil {
		t.Fatalf("MigrateConfig() error = %v", err)
	}

	cfg, err := LoadConfig(v3Path)
	if err != nil {
		t.Fatalf("LoadConfig(v3Path) error = %v", err)
	}

	if cfg.Version != 3 {
		t.Errorf("cfg.Version = %d, want 3", cfg.Version)
	}
	if len(cfg.ClientRules) != 2 {
		t.Errorf("len(ClientRules) = %d, want 2", len(cfg.ClientRules))
	}

	v3Raw, err := os.ReadFile(v3Path)
	if err != nil {
		t.Fatalf("ReadFile(v3Path) error = %v", err)
	}
	if strings.Contains(string(v3Raw), "fcitx5:") {
		t.Errorf("v3 yaml should not contain fcitx5 section, got:\n%s", string(v3Raw))
	}
}

func TestMigrateConfigV2ToV3(t *testing.T) {
	v2Yaml := `version: 2
description: "v2 config"
default_input_method: "keep"
input_methods:
  english: "keyboard-us"
  chinese: "msime"
  japanese: "rime"
fcitx5:
  enabled: true
  rime_input_method: "rime"
rime_schemas:
  japanese: "jaroomaji"
display_names:
  english: "English"
  chinese: "中文"
  japanese: "日本語"
icons:
  chinese: "🇨🇳"
client_rules:
  - class: "kitty"
    input_method: "english"
  - class: "google-chrome"
    input_method: "chinese"
`
	dir := t.TempDir()
	v2Path := filepath.Join(dir, "v2.yaml")
	v3Path := filepath.Join(dir, "v3.yaml")

	if err := os.WriteFile(v2Path, []byte(v2Yaml), 0o644); err != nil {
		t.Fatalf("failed to write v2 config: %v", err)
	}

	if err := MigrateConfig(v2Path, v3Path); err != nil {
		t.Fatalf("MigrateConfig() error = %v", err)
	}

	cfg, err := LoadConfig(v3Path)
	if err != nil {
		t.Fatalf("LoadConfig(v3Path) error = %v", err)
	}

	if cfg.Version != 3 {
		t.Errorf("cfg.Version = %d, want 3", cfg.Version)
	}
	if cfg.InputMethods["chinese"].Engine != "msime" {
		t.Errorf("chinese engine = %q, want msime", cfg.InputMethods["chinese"].Engine)
	}
	if cfg.InputMethods["chinese"].DisplayName != "中文" {
		t.Errorf("chinese display_name = %q, want 中文", cfg.InputMethods["chinese"].DisplayName)
	}
	if cfg.InputMethods["chinese"].Icon != "🇨🇳" {
		t.Errorf("chinese icon = %q, want 🇨🇳", cfg.InputMethods["chinese"].Icon)
	}
	if cfg.InputMethods["japanese"].Engine != "rime" {
		t.Errorf("japanese engine = %q, want rime", cfg.InputMethods["japanese"].Engine)
	}
	if cfg.InputMethods["japanese"].Profile != "jaroomaji" {
		t.Errorf("japanese profile = %q, want jaroomaji", cfg.InputMethods["japanese"].Profile)
	}
	if len(cfg.ClientRules) != 2 {
		t.Errorf("len(ClientRules) = %d, want 2", len(cfg.ClientRules))
	}

	v3Raw, err := os.ReadFile(v3Path)
	if err != nil {
		t.Fatalf("ReadFile(v3Path) error = %v", err)
	}
	if strings.Contains(string(v3Raw), "fcitx5:") {
		t.Errorf("v3 yaml should not contain fcitx5 section, got:\n%s", string(v3Raw))
	}
}

func TestManagerAutoMigrate(t *testing.T) {
	v2Yaml := `version: 2
description: "v2 config"
default_input_method: "keep"
input_methods:
  english: "keyboard-us"
  chinese: "msime"
  japanese: "rime"
rime_schemas:
  japanese: "jaroomaji"
client_rules:
  - class: "kitty"
    input_method: "english"
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(v2Yaml), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		t.Fatalf("mgr.Load() error = %v", err)
	}

	if cfg.Version != 3 {
		t.Errorf("loaded cfg.Version = %d, want 3", cfg.Version)
	}

	// Verify backup file was created
	backupPath := filepath.Join(dir, "config.yaml.v2.bak")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("backup file %s does not exist", backupPath)
	}

	// Verify file on disk was upgraded to version: 3
	fileCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig(configPath) error = %v", err)
	}
	if fileCfg.Version != 3 {
		t.Errorf("on-disk file version = %d, want 3", fileCfg.Version)
	}
}
