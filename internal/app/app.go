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
}

// NewApplication creates a new application instance
func NewApplication() *Application {
	return &Application{}
}

// Run starts the application
func (app *Application) Run() error {
	// Load configuration
	configManager, err := config.NewManager("")
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

	logger.Info("Starting Hyprland input method switcher...")
	logger.Infof("Config loaded from: %s", "~/.config/hypr-input-switcher/config.json")

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")
		cancel()
	}()

	// Start monitoring
	return app.monitorAndSwitch(ctx)
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
	if clientInfo == nil {
		return app.config.DefaultIM
	}

	className := clientInfo.Class
	title := clientInfo.Title

	// Check client rules
	for _, rule := range app.config.ClientRules {
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

	return app.config.DefaultIM
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

						// Show notification
						if app.config.Notifications.ShowOnSwitch {
							app.notifier.ShowInputMethodSwitch(targetIM, clientInfo)
						}
					}
				}
			}
		}
	}
}
