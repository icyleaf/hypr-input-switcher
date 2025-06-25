package inputmethod

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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
	}
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

	// Process initial window
	if err := s.processCurrentWindow(); err != nil {
		logger.Debugf("Error processing initial window: %v", err)
	}

	// Start IPC event monitoring
	return s.monitorHyprlandEvents(ctx)
}

func (s *Switcher) monitorHyprlandEvents(ctx context.Context) error {
	// Get Hyprland IPC socket path
	socketPath := s.getHyprlandEventSocket()
	if socketPath == "" {
		return fmt.Errorf("failed to get Hyprland event socket path")
	}

	logger.Infof("Connecting to Hyprland event socket: %s", socketPath)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping input method switcher...")
			return nil
		default:
		}

		// Connect to socket
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			logger.Errorf("Failed to connect to Hyprland event socket: %v", err)
			// Retry after delay
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}

		logger.Info("Connected to Hyprland event socket")

		// Monitor events
		err = s.handleEvents(ctx, conn)
		conn.Close()

		if err != nil && err != context.Canceled {
			logger.Errorf("Event monitoring error: %v", err)
			// Retry after delay
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func (s *Switcher) handleEvents(ctx context.Context, conn net.Conn) error {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse event
		parts := strings.SplitN(line, ">>", 2)
		if len(parts) != 2 {
			continue
		}

		eventType := parts[0]
		eventData := parts[1]

		logger.Debugf("Received event: %s >> %s", eventType, eventData)

		// Handle window focus events - prefer activewindowv2 for better info
		switch eventType {
		case "activewindowv2":
			if err := s.handleActiveWindowV2Event(eventData); err != nil {
				logger.Debugf("Error handling activewindowv2 event: %v", err)
			}
		case "activewindow":
			// Fallback for older Hyprland versions
			if err := s.handleActiveWindowEvent(eventData); err != nil {
				logger.Debugf("Error handling activewindow event: %v", err)
			}
		}
	}

	return scanner.Err()
}

func (s *Switcher) handleActiveWindowV2Event(eventData string) error {
	// eventData format: "windowaddress" (hex address like 0x12345678)
	windowAddress := strings.TrimSpace(eventData)
	if windowAddress == "" {
		logger.Debugf("Empty activewindowv2 event data")
		return nil
	}

	logger.Debugf("Active window changed to address: %s", windowAddress)

	// Check if this is the same window we're already tracking
	if windowAddress == s.currentClient.Address {
		logger.Debugf("Same window address, skipping: %s", windowAddress)
		return nil
	}

	// Get full client info for the active window
	clientInfo, err := s.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	// Verify the event matches current window address
	if clientInfo.Address != windowAddress {
		logger.Debugf("Event address mismatch: got %s, expected %s", windowAddress, clientInfo.Address)
		return nil
	}

	// Process window change
	return s.processWindowChange(clientInfo)
}

func (s *Switcher) handleActiveWindowEvent(eventData string) error {
	// eventData format: "class,title"
	parts := strings.SplitN(eventData, ",", 2)
	if len(parts) < 2 {
		logger.Debugf("Invalid activewindow event data: %s", eventData)
		return nil
	}

	class := parts[0]
	title := parts[1]

	logger.Debugf("Active window changed: class=%s, title=%s", class, title)

	// Get full client info for the active window
	clientInfo, err := s.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	// Verify the event matches current window
	if clientInfo.Class != class {
		logger.Debugf("Event class mismatch: got %s, expected %s", class, clientInfo.Class)
		return nil
	}

	// Check if this is the same window we're already tracking
	if clientInfo.Address == s.currentClient.Address {
		logger.Debugf("Same window address, skipping: %s", clientInfo.Address)
		return nil
	}

	// Process window change
	return s.processWindowChange(clientInfo)
}

func (s *Switcher) processWindowChange(clientInfo *ClientInfo) error {
	// Update current client info
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

	return nil
}

func (s *Switcher) processCurrentWindow() error {
	clientInfo, err := s.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	return s.processWindowChange(clientInfo)
}

func (s *Switcher) getHyprlandEventSocket() string {
	logger.Debugf("Searching for Hyprland IPC socket...")

	// Get XDG_RUNTIME_DIR
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		logger.Debugf("XDG_RUNTIME_DIR not set, falling back to /tmp")
		runtimeDir = "/tmp"
	}
	logger.Debugf("Using runtime directory: %s", runtimeDir)

	// Try environment variables first
	if hyprInstance := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"); hyprInstance != "" {
		logger.Debugf("Found HYPRLAND_INSTANCE_SIGNATURE: %s", hyprInstance)
		socketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", runtimeDir, hyprInstance)
		logger.Debugf("Checking socket path: %s", socketPath)
		if _, err := os.Stat(socketPath); err == nil {
			logger.Infof("Found Hyprland IPC socket via environment: %s", socketPath)
			return socketPath
		} else {
			logger.Debugf("Hyprland IPC Socket not found via environment: %v", err)
		}
	} else {
		logger.Debugf("HYPRLAND_INSTANCE_SIGNATURE not set")
	}

	// Check if hypr directory exists in runtime dir
	hyprDir := fmt.Sprintf("%s/hypr", runtimeDir)
	if _, err := os.Stat(hyprDir); os.IsNotExist(err) {
		logger.Errorf("Hyprland directory %s does not exist. Is Hyprland running?", hyprDir)
		return ""
	}

	// List all directories in runtime/hypr/
	entries, err := os.ReadDir(hyprDir)
	if err != nil {
		logger.Errorf("Failed to read Hyprland directory %s: %v", hyprDir, err)
		return ""
	}

	logger.Debugf("Found %d entries in %s", len(entries), hyprDir)
	for _, entry := range entries {
		if entry.IsDir() {
			socketPath := fmt.Sprintf("%s/%s/.socket2.sock", hyprDir, entry.Name())
			logger.Debugf("Checking socket: %s", socketPath)
			if _, err := os.Stat(socketPath); err == nil {
				logger.Infof("Found Hyprland event socket: %s", socketPath)
				return socketPath
			} else {
				logger.Debugf("Socket not found: %v", err)
			}
		}
	}

	// Fallback: try to find socket using glob pattern in runtime dir
	globPattern := fmt.Sprintf("%s/hypr/*/.socket2.sock", runtimeDir)
	logger.Debugf("Trying glob pattern: %s", globPattern)
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		logger.Errorf("Glob pattern failed: %v", err)
		return ""
	}

	logger.Debugf("Glob found %d matches", len(matches))
	for _, match := range matches {
		logger.Debugf("Glob match: %s", match)
	}

	if len(matches) == 0 {
		logger.Error("No Hyprland event sockets found. Please check:")
		logger.Error("1. Is Hyprland running?")
		logger.Error("2. Are you running this inside a Hyprland session?")
		logger.Errorf("3. Check if %s/hypr directory exists and contains instance directories", runtimeDir)

		// List what's actually in runtime/hypr if it exists
		if entries, err := os.ReadDir(hyprDir); err == nil {
			logger.Errorf("Contents of %s:", hyprDir)
			for _, entry := range entries {
				logger.Errorf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
			}
		}

		// Also check for legacy /tmp/hypr path
		logger.Debugf("Checking legacy /tmp/hypr path...")
		legacyPattern := "/tmp/hypr/*/.socket2.sock"
		if legacyMatches, err := filepath.Glob(legacyPattern); err == nil && len(legacyMatches) > 0 {
			logger.Infof("Found legacy socket: %s", legacyMatches[0])
			return legacyMatches[0]
		}

		return ""
	}

	// Use the first available socket
	logger.Infof("Using first available socket: %s", matches[0])
	return matches[0]
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
