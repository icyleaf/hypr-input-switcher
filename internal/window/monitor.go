package window

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"hypr-input-switcher/pkg/logger"
)

type ClientInfo struct {
	Class string `json:"class"`
	Title string `json:"title"`
}

type Monitor struct {
	currentClient string
}

func NewMonitor() *Monitor {
	return &Monitor{
		currentClient: "",
	}
}

// StartMonitoring starts monitoring window changes with context support
func (m *Monitor) StartMonitoring(ctx context.Context) error {
	logger.Info("Starting window monitor...")

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping window monitor...")
			return nil

		case <-ticker.C:
			if err := m.processCurrentWindow(); err != nil {
				logger.Debugf("Error processing current window: %v", err)
			}
		}
	}
}

// processCurrentWindow handles the current window change logic
func (m *Monitor) processCurrentWindow() error {
	clientInfo, err := m.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	if clientInfo == nil {
		return nil // No active window
	}

	currentClient := fmt.Sprintf("%s:%s", clientInfo.Class, clientInfo.Title)

	if currentClient != m.currentClient {
		m.currentClient = currentClient
		logger.Infof("Window changed: %s - %s", clientInfo.Class, clientInfo.Title)

		// Note: The actual input method switching logic should be handled
		// by the inputmethod.Switcher, not directly here
		// This monitor just detects window changes
	}

	return nil
}

// getCurrentClient gets the current active window information
func (m *Monitor) getCurrentClient() (*ClientInfo, error) {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl command failed: %w", err)
	}

	var clientInfo ClientInfo
	if err := json.Unmarshal(output, &clientInfo); err != nil {
		return nil, fmt.Errorf("failed to parse hyprctl output: %w", err)
	}

	// Return nil if no window is active (empty class and title)
	if clientInfo.Class == "" && clientInfo.Title == "" {
		return nil, nil
	}

	return &clientInfo, nil
}

// GetCurrentClient returns the current client info (public method)
func (m *Monitor) GetCurrentClient() (*ClientInfo, error) {
	return m.getCurrentClient()
}

// GetCurrentClientString returns the current client as a string
func (m *Monitor) GetCurrentClientString() string {
	return m.currentClient
}

// IsHyprlandAvailable checks if Hyprland/hyprctl is available
func (m *Monitor) IsHyprlandAvailable() bool {
	_, err := exec.LookPath("hyprctl")
	if err != nil {
		logger.Error("hyprctl not found in PATH")
		return false
	}

	// Test if hyprctl can actually communicate with Hyprland
	cmd := exec.Command("hyprctl", "version")
	if err := cmd.Run(); err != nil {
		logger.Error("hyprctl cannot communicate with Hyprland")
		return false
	}

	return true
}
