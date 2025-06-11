package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"hypr-input-switcher/internal/config"
	"hypr-input-switcher/internal/inputmethod"
	"hypr-input-switcher/internal/notification"
	"hypr-input-switcher/pkg/logger"
)

// Application represents the main application
type Application struct {
	config        *config.Config
	configManager *config.Manager
	currentClient string
	currentIM     string
	switcher      *inputmethod.Switcher
	notifier      *notification.Notifier
	watchConfig   bool
}

// NewApplication creates a new application instance
func NewApplication() *Application {
	return &Application{}
}

// Run starts the application
func (app *Application) Run(configPath string, watchConfig bool) error {
	app.watchConfig = watchConfig

	// Load configuration with specified path
	configManager, err := config.NewManager(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app.config = cfg
	app.configManager = configManager
	app.switcher = inputmethod.NewSwitcher(cfg)
	app.notifier = notification.NewNotifier(cfg)

	// Register config change callback
	app.configManager.AddCallback(app.onConfigChanged)

	logger.Info("Starting Hyprland input method switcher...")

	// Show which config file is being used
	logger.Infof("Using config file: %s", app.configManager.GetConfigPath())

	// Start config file watching if enabled
	if watchConfig {
		if err := app.configManager.StartWatching(); err != nil {
			logger.Errorf("Failed to start config watching: %v", err)
			// Continue without watching
		} else {
			logger.Info("Config file watching enabled")
		}
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")

		// Stop config watching
		if app.watchConfig {
			app.configManager.StopWatching()
		}

		cancel()
	}()

	// Start monitoring
	return app.monitorAndSwitch(ctx)
}

// onConfigChanged handles configuration changes
func (app *Application) onConfigChanged(newConfig *config.Config) {
	logger.Info("Applying new configuration...")

	// Update config
	app.config = newConfig

	// Recreate switcher with new config
	app.switcher = inputmethod.NewSwitcher(newConfig)

	// Recreate notifier with new config
	app.notifier = notification.NewNotifier(newConfig)

	// Reset current client to force re-evaluation
	app.currentClient = ""

	logger.Info("Configuration applied successfully")
}

// getCurrentClient gets current active window information
func (app *Application) getCurrentClient() (*config.WindowInfo, error) {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window: %w", err)
	}

	var windowInfo config.WindowInfo
	if err := json.Unmarshal(output, &windowInfo); err != nil {
		return nil, fmt.Errorf("failed to parse window info: %w", err)
	}

	return &windowInfo, nil
}

// getTargetInputMethod determines target input method based on client information
func (app *Application) getTargetInputMethod(clientInfo *config.WindowInfo) string {
	// Get current config (thread-safe)
	currentConfig := app.configManager.GetConfig()

	if clientInfo == nil {
		return currentConfig.DefaultInputMethod
	}

	className := clientInfo.Class
	title := clientInfo.Title

	// Check client rules
	for _, rule := range currentConfig.ClientRules {
		// Match class (required)
		if rule.Class == "" || !app.matchPattern(rule.Class, className) {
			continue
		}

		// If title is empty or not specified, class match is enough
		if rule.Title == "" {
			return rule.InputMethod
		}

		// If title is specified, both class and title must match
		if app.matchPattern(rule.Title, title) {
			return rule.InputMethod
		}
	}

	return currentConfig.DefaultInputMethod
}

// matchPattern matches pattern, supporting regex and string matching
func (app *Application) matchPattern(pattern, text string) bool {
	if pattern == "" || text == "" {
		return false
	}

	// Try as regex first
	if matched, err := regexp.MatchString(pattern, text); err == nil {
		return matched
	}

	// Fallback to string contains matching
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}

// monitorAndSwitch main monitoring loop
func (app *Application) monitorAndSwitch(ctx context.Context) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			// Get current active window
			clientInfo, err := app.getCurrentClient()
			if err != nil {
				logger.Debugf("Failed to get current client: %v", err)
				continue
			}

			currentClient := fmt.Sprintf("%s:%s", clientInfo.Class, clientInfo.Title)

			// If window changed
			if currentClient != app.currentClient {
				app.currentClient = currentClient

				// Get current input method status
				currentIM := app.switcher.GetCurrent()

				// Determine target input method
				targetIM := app.getTargetInputMethod(clientInfo)

				logger.Infof("Window changed: %s - %s", clientInfo.Class, clientInfo.Title)
				logger.Infof("Current IM: %s -> Target IM: %s", currentIM, targetIM)

				// If input method needs to be switched
				if currentIM != targetIM && currentIM != "unknown" {
					if err := app.switcher.Switch(targetIM); err != nil {
						logger.Errorf("Failed to switch input method to %s: %v", targetIM, err)
					} else {
						logger.Infof("Switched input method to: %s", targetIM)

						// Show notification (use current config)
						currentConfig := app.configManager.GetConfig()
						if currentConfig.Notifications.ShowOnSwitch {
							app.notifier.ShowInputMethodSwitch(targetIM, clientInfo)
						}
					}
				}
			}
		}
	}
}
