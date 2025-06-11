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

	"github.com/godbus/dbus/v5"
)

type Switcher struct {
	currentClient string
	currentIM     string
	config        *config.Config
	fcitx5        *Fcitx5
	rime          *Rime
}

type ClientInfo struct {
	Class string `json:"class"`
	Title string `json:"title"`
}

func NewSwitcher(cfg *config.Config) *Switcher {
	switcher := &Switcher{
		currentClient: "",
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

	currentClient := fmt.Sprintf("%s:%s", clientInfo.Class, clientInfo.Title)

	// If window changed
	if currentClient != s.currentClient {
		s.currentClient = currentClient

		// Get current input method status
		currentIM := s.GetCurrent()

		// Determine target input method
		targetIM := s.getTargetInputMethod(clientInfo)

		logger.Infof("Window changed: %s - %s", clientInfo.Class, clientInfo.Title)
		logger.Infof("Current IM: %s -> Target IM: %s", currentIM, targetIM)

		// If input method needs to be switched
		if currentIM != targetIM && currentIM != "unknown" {
			if err := s.Switch(targetIM); err != nil {
				return fmt.Errorf("failed to switch input method to %s: %w", targetIM, err)
			}

			logger.Infof("Switched input method to: %s", targetIM)
			s.currentIM = targetIM
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
	if !s.config.Fcitx5.Enabled {
		return "unknown"
	}

	// Try D-Bus first
	if currentIM := s.getCurrentViaDBus(); currentIM != "unknown" {
		return currentIM
	}

	// Fallback to fcitx5-remote
	cmd := exec.Command("fcitx5-remote", "-n")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	currentIM := strings.TrimSpace(string(output))

	// If it's rime, get current schema
	if currentIM == s.config.Fcitx5.RimeInputMethod {
		return s.getCurrentRimeSchemaViaDBus()
	}

	return "english"
}

func (s *Switcher) getCurrentViaDBus() string {
	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus: %v", err)
		return "unknown"
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	var currentIM string

	err = obj.Call("org.fcitx.Fcitx.Controller1.CurrentInputMethod", 0).Store(&currentIM)
	if err != nil {
		logger.Debugf("Failed to get current input method via D-Bus: %v", err)
		return "unknown"
	}

	logger.Debugf("Current input method via D-Bus: %s", currentIM)

	// If it's rime, get current schema
	if currentIM == s.config.Fcitx5.RimeInputMethod {
		return s.getCurrentRimeSchemaViaDBus()
	}

	if strings.Contains(currentIM, "keyboard") {
		return "english"
	}

	return "english"
}

func (s *Switcher) getCurrentRimeSchemaViaDBus() string {
	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus: %v", err)
		return s.config.DefaultIM
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	var currentSchema string

	err = obj.Call("org.fcitx.Fcitx.Rime1.CurrentSchema", 0).Store(&currentSchema)
	if err != nil {
		logger.Debugf("Failed to get current rime schema via D-Bus: %v", err)
		return s.config.DefaultIM
	}

	logger.Debugf("Current rime schema via D-Bus: %s", currentSchema)

	// Return corresponding input method type based on schema name
	for imType, schemaName := range s.config.RimeSchemas {
		if schemaName == currentSchema {
			return imType
		}
	}

	return s.config.DefaultIM
}

func (s *Switcher) getTargetInputMethod(clientInfo *ClientInfo) string {
	if clientInfo == nil {
		return s.config.DefaultIM
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

	logger.Debugf("No matching rule found, using default: %s", s.config.DefaultIM)
	return s.config.DefaultIM
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
	if !s.config.Fcitx5.Enabled {
		return fmt.Errorf("fcitx5 is not enabled")
	}

	logger.Infof("Switching to input method: %s", targetMethod)

	if targetMethod == "english" {
		return s.switchToEnglishViaDBus()
	}

	return s.switchToRimeViaDBus(targetMethod)
}

func (s *Switcher) switchToEnglishViaDBus() error {
	logger.Debug("Switching to English input method via D-Bus")

	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus, falling back to fcitx5-remote: %v", err)
		cmd := exec.Command("fcitx5-remote", "-c")
		return cmd.Run()
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	call := obj.Call("org.fcitx.Fcitx.Controller1.Deactivate", 0)
	if call.Err != nil {
		logger.Warningf("Failed to deactivate via D-Bus: %v", call.Err)
		cmd := exec.Command("fcitx5-remote", "-c")
		return cmd.Run()
	}

	logger.Info("Successfully switched to English via D-Bus")
	return nil
}

func (s *Switcher) switchToRimeViaDBus(targetMethod string) error {
	logger.Debugf("Switching to Rime input method for: %s", targetMethod)

	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus, falling back to fcitx5-remote: %v", err)
		return s.switchToRimeFallback(targetMethod)
	}
	defer conn.Close()

	// Step 1: Activate input method
	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	call := obj.Call("org.fcitx.Fcitx.Controller1.Activate", 0)
	if call.Err != nil {
		logger.Warningf("Failed to activate via D-Bus: %v", call.Err)
		return s.switchToRimeFallback(targetMethod)
	}

	time.Sleep(100 * time.Millisecond)

	// Step 2: Switch to rime input method
	call = obj.Call("org.fcitx.Fcitx.Controller1.SetCurrentIM", 0, s.config.Fcitx5.RimeInputMethod)
	if call.Err != nil {
		logger.Warningf("Failed to set current IM via D-Bus: %v", call.Err)
		return s.switchToRimeFallback(targetMethod)
	}

	time.Sleep(100 * time.Millisecond)

	// Step 3: Switch rime schema
	return s.switchRimeSchemaViaDBus(targetMethod)
}

func (s *Switcher) switchToRimeFallback(targetMethod string) error {
	logger.Debug("Using fcitx5-remote fallback method")

	cmd := exec.Command("fcitx5-remote", "-o")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to activate input method: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	cmd = exec.Command("fcitx5-remote", "-s", s.config.Fcitx5.RimeInputMethod)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch to rime: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	return s.switchRimeSchemaViaDBus(targetMethod)
}

func (s *Switcher) switchRimeSchemaViaDBus(targetMethod string) error {
	schema, exists := s.config.RimeSchemas[targetMethod]
	if !exists {
		return nil
	}

	logger.Infof("Switching rime schema to: %s via D-Bus", schema)

	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Warningf("Failed to connect to session bus for schema switch: %v", err)
		return s.switchRimeSchemaFallback(targetMethod)
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	call := obj.Call("org.fcitx.Fcitx.Rime1.SetSchema", 0, schema)
	if call.Err != nil {
		logger.Warningf("Failed to switch rime schema via D-Bus: %v", call.Err)
		return s.switchRimeSchemaFallback(targetMethod)
	}

	logger.Infof("Successfully switched rime schema to: %s (D-Bus)", schema)
	return nil
}

func (s *Switcher) switchRimeSchemaFallback(targetMethod string) error {
	schema, exists := s.config.RimeSchemas[targetMethod]
	if !exists {
		return nil
	}

	logger.Infof("Switching rime schema to: %s via fallback method", schema)

	cmd := exec.Command("dbus-send",
		"--type=method_call",
		"--dest=org.fcitx.Fcitx5",
		"/rime",
		"org.fcitx.Fcitx.Rime1.SetSchema",
		fmt.Sprintf("string:%s", schema))

	if err := cmd.Run(); err != nil {
		logger.Warningf("Failed to switch rime schema via dbus-send: %v", err)
		// TODO: implement updateRimeConfig fallback
		return err
	}

	logger.Infof("Successfully switched rime schema to: %s (dbus-send)", schema)
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
		"current_client": s.currentClient,
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
