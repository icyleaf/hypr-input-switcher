package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hypr-input-switcher/pkg/logger"
)

const CurrentConfigVersion = 2

// Manager handles configuration loading and migration
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) (*Manager, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".config", "hypr-input-switcher", "config.json")
	}

	return &Manager{
		configPath: configPath,
	}, nil
}

// Load loads the configuration file
func (m *Manager) Load() (*Config, error) {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		logger.Info("Config file not found, creating default config")
		if err := m.createDefaultConfig(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		m.config = getDefaultConfig()
		return m.config, nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check version compatibility
	if !m.checkConfigVersion(&config) {
		config = *getDefaultConfig()
	}

	m.config = &config
	return m.config, nil
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// checkConfigVersion checks and handles config version compatibility
func (m *Manager) checkConfigVersion(config *Config) bool {
	if config.Version == 0 {
		config.Version = 1 // Default to version 1 for old configs
	}

	if config.Version == CurrentConfigVersion {
		logger.Infof("Config version %d is current", config.Version)
		return true
	} else if config.Version < CurrentConfigVersion {
		logger.Warningf("Config version %d is outdated (current: %d)", config.Version, CurrentConfigVersion)
		return m.handleConfigUpgrade(config)
	} else {
		logger.Errorf("Config version %d is newer than supported (current: %d)", config.Version, CurrentConfigVersion)
		fmt.Printf("Error: Configuration file version %d is newer than supported version %d\n", config.Version, CurrentConfigVersion)
		fmt.Println("Please update hypr-smart-input to the latest version or downgrade your config file.")
		return false
	}
}

// handleConfigUpgrade handles configuration upgrades
func (m *Manager) handleConfigUpgrade(oldConfig *Config) bool {
	fmt.Printf("\n⚠️  Configuration Update Required\n")
	fmt.Printf("Your config file is version %d, but current version is %d\n", oldConfig.Version, CurrentConfigVersion)
	fmt.Printf("Config location: %s\n", m.configPath)

	m.showConfigChanges(oldConfig.Version, CurrentConfigVersion)

	for {
		fmt.Print("\nChoose an option:\n")
		fmt.Print("1. Backup old config and create new one (recommended)\n")
		fmt.Print("2. Continue with old config (may cause issues)\n")
		fmt.Print("3. Exit and manually update config\n")
		fmt.Print("Enter choice (1-3): ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			return m.backupAndUpgradeConfig(oldConfig)
		case "2":
			logger.Warning("Continuing with outdated config - some features may not work")
			return true
		case "3":
			fmt.Println("Please update your configuration file manually.")
			fmt.Println("You can find example config at: https://github.com/your-repo/examples/config.json")
			os.Exit(0)
		default:
			fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
		}
	}
}

// showConfigChanges shows what changes between config versions
func (m *Manager) showConfigChanges(oldVersion, newVersion int) {
	fmt.Printf("\nChanges from version %d to %d:\n", oldVersion, newVersion)

	if oldVersion == 1 && newVersion >= 2 {
		fmt.Println("  • client_rules format changed from dict to list of objects")
		fmt.Println("  • Added support for regex patterns in class and title matching")
		fmt.Println("  • Enhanced notification system with multiple backends")
	}
}

// backupAndUpgradeConfig backs up old config and creates new one
func (m *Manager) backupAndUpgradeConfig(oldConfig *Config) bool {
	// Create backup
	backupPath := m.configPath + ".backup"
	counter := 1
	for {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break
		}
		backupPath = fmt.Sprintf("%s.backup.%d", m.configPath, counter)
		counter++
	}

	// Write backup
	oldData, err := json.MarshalIndent(oldConfig, "", "  ")
	if err != nil {
		logger.Errorf("Failed to marshal old config: %v", err)
		return false
	}

	if err := os.WriteFile(backupPath, oldData, 0644); err != nil {
		logger.Errorf("Failed to create backup: %v", err)
		return false
	}

	fmt.Printf("✅ Old config backed up to: %s\n", backupPath)

	// Create new config with migrated settings
	newConfig := getDefaultConfig()
	m.migrateConfigSettings(oldConfig, newConfig)

	// Save new config
	newData, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		logger.Errorf("Failed to marshal new config: %v", err)
		return false
	}

	if err := os.WriteFile(m.configPath, newData, 0644); err != nil {
		logger.Errorf("Failed to save new config: %v", err)
		return false
	}

	fmt.Printf("✅ New config created at: %s\n", m.configPath)
	fmt.Println("   Some of your old settings have been migrated.")
	fmt.Println("   Please review and adjust the new configuration as needed.")

	return true
}

// migrateConfigSettings migrates settings from old config to new config
func (m *Manager) migrateConfigSettings(oldConfig, newConfig *Config) {
	// Migrate basic settings
	if oldConfig.DefaultIM != "" {
		newConfig.DefaultIM = oldConfig.DefaultIM
	}

	if oldConfig.Fcitx5.Enabled {
		newConfig.Fcitx5 = oldConfig.Fcitx5
	}

	if len(oldConfig.RimeSchemas) > 0 {
		newConfig.RimeSchemas = oldConfig.RimeSchemas
	}

	if len(oldConfig.DisplayNames) > 0 {
		newConfig.DisplayNames = oldConfig.DisplayNames
	}

	if len(oldConfig.Icons) > 0 {
		newConfig.Icons = oldConfig.Icons
	}

	// Migrate client rules - already in correct format in our Go version
	if len(oldConfig.ClientRules) > 0 {
		newConfig.ClientRules = oldConfig.ClientRules
		logger.Infof("Migrated %d client rules", len(oldConfig.ClientRules))
	}
}

// createDefaultConfig creates a default configuration file
func (m *Manager) createDefaultConfig() error {
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := getDefaultConfig()
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Created default config at: %s\n", m.configPath)
	return nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	return &Config{
		Version:     CurrentConfigVersion,
		Description: "Hyprland Input Method Switcher Configuration",
		InputMethods: map[string]string{
			"english":  "keyboard-us",
			"chinese":  "rime",
			"japanese": "rime",
		},
		ClientRules: []ClientRule{
			{"firefox", "", "chinese"},
			{"google-chrome", "", "japanese"},
			{"chromium", "", "chinese"},
			{"wechat", "", "chinese"},
			{"code", "", "english"},
			{"vim", "", "english"},
			{"nvim", "", "english"},
			{"terminal", "", "english"},
			{"kitty", "", "english"},
			{"alacritty", "", "english"},
			{"wezterm", "", "english"},
			{"obsidian", "", "chinese"},
			{"typora", "", "chinese"},
			{"anki", "", "japanese"},
		},
		DefaultIM: "english",
		Fcitx5: Fcitx5Config{
			Enabled:         true,
			RimeInputMethod: "rime",
		},
		RimeSchemas: map[string]string{
			"chinese":  "rime_frost",
			"japanese": "jaroomaji",
		},
		Notifications: NotificationConfig{
			Enabled:      true,
			Duration:     2000,
			ShowOnSwitch: true,
			ShowAppName:  true,
		},
		DisplayNames: map[string]string{
			"english":  "English",
			"chinese":  "中文",
			"japanese": "日本語",
		},
		Icons: map[string]string{
			"english":  "input-keyboard",
			"chinese":  "input-method",
			"japanese": "input-method",
		},
	}
}
