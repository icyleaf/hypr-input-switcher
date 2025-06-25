package inputmethod

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"hypr-input-switcher/internal/config"
	"hypr-input-switcher/pkg/logger"
)

type Switcher struct {
	currentClient *ClientInfo
	currentIM     string
	config        *config.Config
	fcitx5        *Fcitx5
	rime          *Rime
	notifier      interface {
		ShowInputMethodSwitch(inputMethod string, clientInfo *config.WindowInfo)
	} // Add notifier interface
}

type ClientInfo struct {
	Address string `json:"address"`
	Class   string `json:"class"`
	Title   string `json:"title"`
}

func NewSwitcher(cfg *config.Config) *Switcher {
	switcher := &Switcher{
		currentClient: &ClientInfo{},
		currentIM:     "",
		config:        cfg,
	}

	// Initialize input method handlers
	if cfg.Fcitx5.Enabled {
		switcher.fcitx5 = NewFcitx5(cfg.Fcitx5.RimeInputMethod)
		switcher.rime = NewRime(cfg.RimeSchemas)
	}

	return switcher
}

// SetNotifier sets the notifier for the switcher
func (s *Switcher) SetNotifier(notifier interface {
	ShowInputMethodSwitch(inputMethod string, clientInfo *config.WindowInfo)
}) {
	s.notifier = notifier
}

func (s *Switcher) MonitorAndSwitch(ctx context.Context) error {
	logger.Info("Starting Hyprland input method switcher...")

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping input method switcher...")
			return nil

		case <-ticker.C:
			if err := s.processCurrentWindow(); err != nil {
				logger.Debugf("Error processing current window: %v", err)
			}
		}
	}
}

func (s *Switcher) processCurrentWindow() error {
	clientInfo, err := s.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	// If window changed (different address means different window)
	if clientInfo.Address != s.currentClient.Address {
		s.currentClient = clientInfo

		// Get current input method status
		currentIM := s.GetCurrent()

		// Determine target input method
		targetIM := s.getTargetInputMethod(clientInfo)

		logger.Infof("Window changed: %s - %s (address: %s)", clientInfo.Class, clientInfo.Title, clientInfo.Address)
		logger.Infof("Current IM: %s -> Target IM: %s", currentIM, targetIM)

		// If input method needs to be switched
		if currentIM != targetIM && currentIM != "unknown" {
			if err := s.Switch(targetIM); err != nil {
				return fmt.Errorf("failed to switch input method to %s: %w", targetIM, err)
			}

			logger.Infof("Switched input method to: %s", targetIM)
			s.currentIM = targetIM

			// Show notification if notifier is available and enabled
			if s.notifier != nil && s.config.Notifications.ShowOnSwitch {
				// Convert ClientInfo to config.WindowInfo
				windowInfo := &config.WindowInfo{
					Class: clientInfo.Class,
					Title: clientInfo.Title,
				}
				s.notifier.ShowInputMethodSwitch(targetIM, windowInfo)
			}
		}
	}

	return nil
}

func (s *Switcher) getCurrentClient() (*ClientInfo, error) {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl command failed: %w", err)
	}

	var clientInfo ClientInfo
	if err := json.Unmarshal(output, &clientInfo); err != nil {
		return nil, fmt.Errorf("failed to parse hyprctl output: %w", err)
	}

	return &clientInfo, nil
}

func (s *Switcher) GetCurrent() string {
	if !s.config.Fcitx5.Enabled || s.fcitx5 == nil {
		return "unknown"
	}

	// Get current fcitx5 input method
	currentIM := s.fcitx5.GetCurrent()

	// If it's Rime, get the specific input method based on schema
	if currentIM == "rime" && s.rime != nil {
		return s.rime.GetCurrentInputMethod(s.config.DefaultInputMethod)
	}

	return currentIM
}

func (s *Switcher) getTargetInputMethod(clientInfo *ClientInfo) string {
	if clientInfo == nil {
		return s.config.DefaultInputMethod
	}

	className := clientInfo.Class
	title := clientInfo.Title

	logger.Debugf("Matching rules for class: %s, title: %s", className, title)

	// Check client rules
	for _, rule := range s.config.ClientRules {
		// Match class (required)
		if rule.Class == "" || !s.matchPattern(rule.Class, className) {
			continue
		}

		// If title is empty or not specified, class match is enough
		if rule.Title == "" {
			logger.Debugf("Matched rule: class=%s -> %s", rule.Class, rule.InputMethod)
			return rule.InputMethod
		}

		// If title is specified, both class and title must match
		if s.matchPattern(rule.Title, title) {
			logger.Debugf("Matched rule: class=%s, title=%s -> %s", rule.Class, rule.Title, rule.InputMethod)
			return rule.InputMethod
		}
	}

	logger.Debugf("No matching rule found, using default: %s", s.config.DefaultInputMethod)
	return s.config.DefaultInputMethod
}

func (s *Switcher) matchPattern(pattern, text string) bool {
	if pattern == "" || text == "" {
		return false
	}

	// Try as regex first
	if matched, err := regexp.MatchString(pattern, text); err == nil {
		logger.Debugf("Regex match '%s' against '%s': %v", pattern, text, matched)
		return matched
	}

	// Fallback to case-insensitive string contains matching
	matched := strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
	logger.Debugf("String contains match '%s' against '%s': %v", pattern, text, matched)
	return matched
}

func (s *Switcher) Switch(targetMethod string) error {
	if !s.config.Fcitx5.Enabled || s.fcitx5 == nil {
		return fmt.Errorf("fcitx5 is not enabled")
	}

	logger.Infof("Switching to input method: %s", targetMethod)

	if targetMethod == "english" {
		return s.fcitx5.SwitchToEnglish()
	}

	// For non-English methods, switch to Rime first, then set schema
	if err := s.fcitx5.SwitchToRime(); err != nil {
		return fmt.Errorf("failed to switch to Rime: %w", err)
	}

	// Wait a bit for the switch to take effect
	time.Sleep(100 * time.Millisecond)

	// Switch to specific schema if Rime is available
	if s.rime != nil {
		return s.rime.SwitchSchema(targetMethod)
	}

	return nil
}

// IsReady checks if the switcher is ready to operate
func (s *Switcher) IsReady() bool {
	// Check if hyprctl is available
	if _, err := exec.LookPath("hyprctl"); err != nil {
		logger.Error("hyprctl not found in PATH")
		return false
	}

	// Check if fcitx5 is available and enabled
	if s.config.Fcitx5.Enabled {
		if s.fcitx5 == nil || !s.fcitx5.IsAvailable() {
			logger.Error("fcitx5 is enabled but not available")
			return false
		}
	}

	return true
}

// GetStatus returns current status information
func (s *Switcher) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"current_client": s.currentClient, // Now contains the window address
		"current_im":     s.currentIM,
		"fcitx5_enabled": s.config.Fcitx5.Enabled,
		"ready":          s.IsReady(),
	}

	if s.fcitx5 != nil {
		status["fcitx5_available"] = s.fcitx5.IsAvailable()
	}

	if s.rime != nil {
		status["rime_available"] = s.rime.IsAvailable()
	}

	return status
}
