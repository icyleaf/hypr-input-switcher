package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"hypr-input-switcher/internal/config"
	"hypr-input-switcher/internal/inputmethod"
	"hypr-input-switcher/internal/notification"
	"hypr-input-switcher/pkg/logger"
)

// Application represents the main application
type Application struct {
	config        *config.Config
	configManager *config.Manager
	switcher      *inputmethod.Switcher
	notifier      *notification.Notifier
	watchConfig   bool

	// Add fields to manage the monitoring context
	monitorCtx    context.Context
	monitorCancel context.CancelFunc
	monitorMutex  sync.Mutex
	restartChan   chan struct{}
}

// NewApplication creates a new application instance
func NewApplication() *Application {
	return &Application{
		restartChan: make(chan struct{}, 1),
	}
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

	// Set notifier for switcher
	app.switcher.SetNotifier(app.notifier)

	// Register config change callback
	app.configManager.AddCallback(app.onConfigChanged)

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

		// Cancel monitor context
		app.monitorMutex.Lock()
		if app.monitorCancel != nil {
			app.monitorCancel()
		}
		app.monitorMutex.Unlock()

		cancel()
	}()

	// Start monitoring loop with restart capability
	return app.runMonitoringLoop(ctx)
}

// runMonitoringLoop handles the main monitoring with restart capability
func (app *Application) runMonitoringLoop(ctx context.Context) error {
	for {
		// Create a new context for this monitoring session
		app.monitorMutex.Lock()
		app.monitorCtx, app.monitorCancel = context.WithCancel(ctx)
		monitorCtx := app.monitorCtx
		app.monitorMutex.Unlock()

		// Start monitoring in a goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- app.switcher.MonitorAndSwitch(monitorCtx)
		}()

		// Wait for either completion, restart signal, or parent context cancellation
		select {
		case err := <-errChan:
			if err != nil && ctx.Err() == nil {
				return err
			}
			return nil

		case <-app.restartChan:
			logger.Info("Restarting monitoring loop with new configuration...")
			// Cancel current monitoring
			app.monitorMutex.Lock()
			if app.monitorCancel != nil {
				app.monitorCancel()
			}
			app.monitorMutex.Unlock()
			// Wait for current monitoring to stop
			<-errChan
			// Continue to next iteration to restart

		case <-ctx.Done():
			app.monitorMutex.Lock()
			if app.monitorCancel != nil {
				app.monitorCancel()
			}
			app.monitorMutex.Unlock()
			return nil
		}
	}
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

	// Set notifier for switcher
	app.switcher.SetNotifier(app.notifier)

	logger.Info("Configuration applied successfully")

	// Signal to restart monitoring loop
	select {
	case app.restartChan <- struct{}{}:
	default:
		// Channel is full, restart is already pending
	}
}
