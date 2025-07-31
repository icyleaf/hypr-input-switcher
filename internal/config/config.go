package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hypr-input-switcher/pkg/logger"

	"github.com/fsnotify/fsnotify"
)

type Manager struct {
	configPath     string
	config         *Config
	mutex          sync.RWMutex
	watcher        *fsnotify.Watcher
	callbacks      []func(*Config)
	callbacksMutex sync.RWMutex // Add this field

	// Debounce related fields
	debounceTimer *time.Timer
	debounceMutex sync.Mutex
	debounceDelay time.Duration
}

func NewManager(configPath string) (*Manager, error) {
	// If no config path provided, use default
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// Expand environment variables in the path
	configPath = os.ExpandEnv(configPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for config: %w", err)
	}
	configPath = absPath

	logger.Debugf("Config manager initialized with path: %s", configPath)

	return &Manager{
		configPath: configPath,
		callbacks:  make([]func(*Config), 0),

		debounceDelay: 200 * time.Millisecond, // 200ms debounce delay
	}, nil
}

// GetConfigPath returns the current config file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

func (m *Manager) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		logger.Infof("Config file not found at %s, creating default config", m.configPath)
		if err := m.createDefaultConfig(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	config, err := LoadConfig(m.configPath)
	if err != nil {
		return nil, err
	}

	m.mutex.Lock()
	m.config = config
	m.mutex.Unlock()

	logger.Infof("Configuration loaded from: %s", m.configPath)
	return config, nil
}

func (m *Manager) GetConfig() *Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.config
}

// createDefaultConfig creates a default configuration file by copying from configs/default.yaml
func (m *Manager) createDefaultConfig() error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	// Find the default config file
	defaultConfigPath, err := findDefaultConfigFile()
	if err != nil {
		return fmt.Errorf("failed to find default config file: %w", err)
	}

	// Copy the default config file
	if err := copyFile(defaultConfigPath, m.configPath); err != nil {
		return fmt.Errorf("failed to copy default config: %w", err)
	}

	logger.Infof("Default configuration created at: %s", m.configPath)
	return nil
}

// findDefaultConfigFile finds the default config file in possible locations
func findDefaultConfigFile() (string, error) {
	// Possible locations for the default config file
	possiblePaths := []string{
		"configs/default.yaml",                        // Relative to current directory
		"./configs/default.yaml",                      // Explicit relative path
		"/usr/share/hypr-input-switcher/default.yaml", // System-wide installation
		"/etc/hypr-input-switcher/default.yaml",       // System config
	}

	// Get executable directory and add it to possible paths
	if execDir, err := getExecutableDir(); err == nil {
		possiblePaths = append([]string{
			filepath.Join(execDir, "configs", "default.yaml"),
			filepath.Join(execDir, "..", "configs", "default.yaml"), // For development
		}, possiblePaths...)
	}

	// Try each possible path
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			logger.Debugf("Found default config at: %s", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("default config file not found in any of the expected locations: %v", possiblePaths)
}

// getExecutableDir returns the directory containing the executable
func getExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Sync to ensure the file is written to disk
	return destFile.Sync()
}

// AddCallback adds a callback function to be called when config changes
func (m *Manager) AddCallback(callback func(*Config)) {
	m.callbacksMutex.Lock()
	defer m.callbacksMutex.Unlock()
	m.callbacks = append(m.callbacks, callback)
}

// StartWatching starts watching the config file for changes
func (m *Manager) StartWatching() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	m.watcher = watcher

	// Add config file to watcher
	err = watcher.Add(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to watch config file %s: %w", m.configPath, err)
	}

	// Also watch the directory in case the file gets replaced
	configDir := filepath.Dir(m.configPath)
	err = watcher.Add(configDir)
	if err != nil {
		logger.Warningf("Failed to watch config directory %s: %v", configDir, err)
	}

	go m.watchLoop()

	logger.Debugf("Started watching config file: %s", m.configPath)
	return nil
}

// StopWatching stops watching the config file
func (m *Manager) StopWatching() error {
	// Clean up debounce timer
	m.debounceMutex.Lock()
	if m.debounceTimer != nil {
		m.debounceTimer.Stop()
		m.debounceTimer = nil
	}
	m.debounceMutex.Unlock()

	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}

func (m *Manager) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			// Only handle write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				logger.Debugf("Config file changed: %s", event.Name)
				m.handleFileChangeDebounced()
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			logger.Errorf("Config file watch error: %v", err)
		}
	}
}

// Debounced file change handler
func (m *Manager) handleFileChangeDebounced() {
	m.debounceMutex.Lock()
	defer m.debounceMutex.Unlock()

	// Cancel previous timer
	if m.debounceTimer != nil {
		m.debounceTimer.Stop()
	}

	// Set new timer
	m.debounceTimer = time.AfterFunc(m.debounceDelay, func() {
		m.handleFileChange()
	})
}

// Handle actual file change
func (m *Manager) handleFileChange() {
	logger.Debug("Reloading configuration...")

	newConfig, err := m.Load()
	if err != nil {
		logger.Errorf("Failed to reload config: %v", err)
		return
	}

	logger.Debug("Configuration reloaded successfully")

	// Call all callback functions
	m.callbacksMutex.RLock()
	callbacks := make([]func(*Config), len(m.callbacks))
	copy(callbacks, m.callbacks)
	m.callbacksMutex.RUnlock()

	for _, callback := range callbacks {
		go callback(newConfig)
	}
}

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if can't get home directory
		return "./config.yaml"
	}
	return filepath.Join(homeDir, ".config", "hypr-input-switcher", "config.yaml")
}
