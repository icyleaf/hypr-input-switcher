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
	layers        *layerTracker
	notifier      interface {
		ShowInputMethodSwitch(inputMethod string, clientInfo *config.WindowInfo)
	}
}

// layerTracker counts how many instances of each layer-shell namespace are
// currently open. The same namespace can be open on several monitors at once,
// so a namespace only counts as "open" while its refcount is above zero.
type layerTracker struct {
	counts map[string]int
}

func newLayerTracker() *layerTracker {
	return &layerTracker{counts: make(map[string]int)}
}

func (l *layerTracker) open(namespace string) {
	if namespace == "" {
		return
	}
	l.counts[namespace]++
}

func (l *layerTracker) close(namespace string) {
	if namespace == "" {
		return
	}
	if count, ok := l.counts[namespace]; ok {
		if count <= 1 {
			delete(l.counts, namespace)
			return
		}
		l.counts[namespace] = count - 1
	}
}

func (l *layerTracker) isOpen(namespace string) bool {
	return l.counts[namespace] > 0
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
		layers:        newLayerTracker(),
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
	logger.Debug("Starting Hyprland input method switcher...")

	// Process initial window
	if err := s.processCurrentWindow(); err != nil {
		logger.Warningf("Error processing initial window: %v", err)
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

	logger.Debugf("Connecting to Hyprland event socket: %s", socketPath)

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

		logger.Debug("Connected to Hyprland event socket")

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

		logger.Tracef("Received event: %s >> %s", eventType, eventData)

		// Handle window focus events - prefer activewindowv2 for better info
		switch eventType {
		case "activewindowv2":
			if err := s.handleActiveWindowV2Event(eventData); err != nil {
				logger.Warningf("Error handling activewindowv2 event: %v", err)
			}
		case "activewindow":
			// Fallback for older Hyprland versions
			if err := s.handleActiveWindowEvent(eventData); err != nil {
				logger.Warningf("Error handling activewindow event: %v", err)
			}
		case "openlayer":
			s.handleLayerEvent(eventData, true)
		case "closelayer":
			s.handleLayerEvent(eventData, false)
		}
	}

	return scanner.Err()
}

func (s *Switcher) handleActiveWindowV2Event(eventData string) error {
	// eventData format: "windowaddress" (hex address like 0x12345678)
	windowAddress := strings.TrimSpace(eventData)
	if windowAddress == "" {
		logger.Tracef("Empty activewindowv2 event data")
		return nil
	}

	logger.Tracef("Active window changed to address: %s", windowAddress)

	// Check if this is the same window we're already tracking
	if windowAddress == s.currentClient.Address {
		logger.Tracef("Same window address, skipping: %s", windowAddress)
		return nil
	}

	// Get full client info for the active window
	clientInfo, err := s.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	// Verify the event matches current window address
	if clientInfo.Address != windowAddress {
		logger.Tracef("Event address mismatch: got %s, expected %s", windowAddress, clientInfo.Address)
		return nil
	}

	// Process window change
	return s.processWindowChange(clientInfo)
}

func (s *Switcher) handleActiveWindowEvent(eventData string) error {
	// eventData format: "class,title"
	parts := strings.SplitN(eventData, ",", 2)
	if len(parts) < 2 {
		logger.Warningf("Invalid activewindow event data: %s", eventData)
		return nil
	}

	class := parts[0]
	title := parts[1]

	logger.Tracef("Active window changed: class=%s, title=%s", class, title)

	// Get full client info for the active window
	clientInfo, err := s.getCurrentClient()
	if err != nil {
		return fmt.Errorf("failed to get current client: %w", err)
	}

	// Verify the event matches current window
	if clientInfo.Class != class {
		logger.Tracef("Event class mismatch: got %s, expected %s", class, clientInfo.Class)
		return nil
	}

	// Check if this is the same window we're already tracking
	if clientInfo.Address == s.currentClient.Address {
		logger.Tracef("Same window address, skipping: %s", clientInfo.Address)
		return nil
	}

	// Process window change
	return s.processWindowChange(clientInfo)
}

func (s *Switcher) processWindowChange(clientInfo *ClientInfo) error {
	// Update current client info
	s.currentClient = clientInfo

	logger.Debugf("Window changed: %s - %s (address: %s)", clientInfo.Class, clientInfo.Title, clientInfo.Address)

	return s.applyResolution(s.resolveTarget(clientInfo), clientInfo)
}

// handleLayerEvent updates the open-layer bookkeeping and recomputes the target
// input method. Layer surfaces do not become the active window, so opening or
// closing one never produces an activewindow event we could rely on; the
// recompute re-reads the window underneath and re-applies the window rules when
// no layer rule applies.
func (s *Switcher) handleLayerEvent(namespace string, isOpen bool) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		logger.Tracef("Empty layer event namespace")
		return
	}

	if isOpen {
		s.layers.open(namespace)
	} else {
		s.layers.close(namespace)
	}

	logger.Debugf("Layer %s: open=%v (still open: %v)", namespace, isOpen, s.layers.isOpen(namespace))

	clientInfo, err := s.getCurrentClient()
	if err != nil {
		logger.Warningf("Failed to get current client for layer change: %v", err)
		clientInfo = s.currentClient
	}
	if clientInfo != nil {
		s.currentClient = clientInfo
	}

	if err := s.applyResolution(s.resolveTarget(clientInfo), clientInfo); err != nil {
		logger.Warningf("Error applying layer rule for %s: %v", namespace, err)
	}
}

// applyResolution switches to the resolved input method. The keep target
// preserves the active input method without querying or invoking the configured
// input method backend.
func (s *Switcher) applyResolution(resolution targetResolution, clientInfo *ClientInfo) error {
	targetIM := resolution.inputMethod

	if targetIM == config.KeepInputMethod {
		logger.Debug("Keeping current input method")
		return nil
	}

	// Get current input method status
	currentIM := s.GetCurrent()

	logger.Debugf("Current IM: %s -> Target IM: %s", currentIM, targetIM)

	if currentIM == targetIM || currentIM == "unknown" {
		return nil
	}

	if err := s.Switch(targetIM); err != nil {
		return fmt.Errorf("failed to switch input method to %s: %w", targetIM, err)
	}

	logger.Debugf("Switched input method to: %s", targetIM)
	s.currentIM = targetIM

	// Show notification if notifier is available and enabled
	if s.notifier != nil && s.config.Notifications.ShowOnSwitch {
		windowInfo := &config.WindowInfo{}
		if resolution.layerNamespace != "" {
			windowInfo.Class = resolution.layerNamespace
		} else if clientInfo != nil {
			windowInfo.Class = clientInfo.Class
			windowInfo.Title = clientInfo.Title
		}
		s.notifier.ShowInputMethodSwitch(targetIM, windowInfo)
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
	logger.Tracef("Searching for Hyprland IPC socket...")

	// Get XDG_RUNTIME_DIR
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		logger.Tracef("XDG_RUNTIME_DIR not set, falling back to /tmp")
		runtimeDir = "/tmp"
	}
	logger.Tracef("Using runtime directory: %s", runtimeDir)

	// Try environment variables first
	if hyprInstance := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"); hyprInstance != "" {
		logger.Tracef("Found HYPRLAND_INSTANCE_SIGNATURE: %s", hyprInstance)
		socketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", runtimeDir, hyprInstance)
		logger.Tracef("Checking socket path: %s", socketPath)
		if _, err := os.Stat(socketPath); err == nil {
			logger.Debug("Found Hyprland IPC socket via environment")
			return socketPath
		} else {
			logger.Tracef("Hyprland IPC Socket not found via environment: %v", err)
		}
	} else {
		logger.Tracef("HYPRLAND_INSTANCE_SIGNATURE not set")
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

	logger.Tracef("Found %d entries in %s", len(entries), hyprDir)
	for _, entry := range entries {
		if entry.IsDir() {
			socketPath := fmt.Sprintf("%s/%s/.socket2.sock", hyprDir, entry.Name())
			logger.Tracef("Checking socket: %s", socketPath)
			if _, err := os.Stat(socketPath); err == nil {
				logger.Debug("Found Hyprland event socket")
				return socketPath
			} else {
				logger.Tracef("Socket not found: %v", err)
			}
		}
	}

	// Fallback: try to find socket using glob pattern in runtime dir
	globPattern := fmt.Sprintf("%s/hypr/*/.socket2.sock", runtimeDir)
	logger.Tracef("Trying glob pattern: %s", globPattern)
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		logger.Errorf("Glob pattern failed: %v", err)
		return ""
	}

	logger.Tracef("Glob found %d matches", len(matches))
	for _, match := range matches {
		logger.Tracef("Glob match: %s", match)
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
		logger.Trace("Checking legacy /tmp/hypr path...")
		legacyPattern := "/tmp/hypr/*/.socket2.sock"
		if legacyMatches, err := filepath.Glob(legacyPattern); err == nil && len(legacyMatches) > 0 {
			logger.Debug("Found legacy socket")
			return legacyMatches[0]
		}

		return ""
	}

	// Use the first available socket
	logger.Debug("Using first available socket")
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

// targetResolution is the outcome of matching the current focus (an open layer
// surface or the active window) against the configured rules. layerNamespace is
// non-empty when a layer rule decided the result, which lets callers report the
// overlay rather than the window underneath.
type targetResolution struct {
	inputMethod    string
	layerNamespace string
}

func (s *Switcher) resolveTarget(clientInfo *ClientInfo) targetResolution {
	// Open layer surfaces take precedence over the active window: an overlay is
	// the surface the user is currently interacting with, even though Hyprland
	// still reports the window underneath as the active one.
	if namespace, target, ok := s.getLayerTargetInputMethod(); ok {
		return targetResolution{inputMethod: target, layerNamespace: namespace}
	}

	if clientInfo == nil {
		return targetResolution{inputMethod: s.config.DefaultInputMethod}
	}

	className := clientInfo.Class
	title := clientInfo.Title

	logger.Tracef("Matching rules for class: %s, title: %s", className, title)

	// Check client rules
	for _, rule := range s.config.ClientRules {
		// Match class (required)
		if rule.Class == "" || !s.matchPattern(rule.Class, className) {
			continue
		}

		// If title is empty or not specified, class match is enough
		if rule.Title == "" {
			logger.Tracef("Matched rule: class=%s -> %s", rule.Class, rule.InputMethod)
			return targetResolution{inputMethod: rule.InputMethod}
		}

		// If title is specified, both class and title must match
		if s.matchPattern(rule.Title, title) {
			logger.Tracef("Matched rule: class=%s, title=%s -> %s", rule.Class, rule.Title, rule.InputMethod)
			return targetResolution{inputMethod: rule.InputMethod}
		}
	}

	logger.Tracef("No matching rule found, using default: %s", s.config.DefaultInputMethod)
	return targetResolution{inputMethod: s.config.DefaultInputMethod}
}

// getLayerTargetInputMethod returns the namespace and input method dictated by
// the first layer rule whose namespace is currently open. The third return
// value is false when no open namespace matches a rule.
func (s *Switcher) getLayerTargetInputMethod() (string, string, bool) {
	if s.layers == nil {
		return "", "", false
	}

	for _, rule := range s.config.LayerRules {
		if rule.Namespace == "" {
			continue
		}
		if s.layers.isOpen(rule.Namespace) {
			logger.Tracef("Matched layer rule: namespace=%s -> %s", rule.Namespace, rule.InputMethod)
			return rule.Namespace, rule.InputMethod, true
		}
	}

	return "", "", false
}

func (s *Switcher) matchPattern(pattern, text string) bool {
	if pattern == "" || text == "" {
		return false
	}

	// Try as regex first
	if matched, err := regexp.MatchString(pattern, text); err == nil {
		logger.Tracef("Regex match '%s' against '%s': %v", pattern, text, matched)
		return matched
	}

	// Fallback to case-insensitive string contains matching
	matched := strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
	logger.Tracef("String contains match '%s' against '%s': %v", pattern, text, matched)
	return matched
}

// verifySwitchReachedTarget reports whether an observed input method confirms
// that a switch to target actually took effect. The D-Bus calls used to switch
// methods can return success while silently doing nothing (for example when the
// target method is not present in the active fcitx5 input method group), so the
// result is re-checked against the observed value rather than trusted.
func verifySwitchReachedTarget(target, observed string) bool {
	if observed == "" || observed == "unknown" {
		return false
	}
	return target == observed
}

func (s *Switcher) Switch(targetMethod string) error {
	if targetMethod == config.KeepInputMethod {
		logger.Debug("Keeping current input method")
		return nil
	}

	if !s.config.Fcitx5.Enabled || s.fcitx5 == nil {
		return fmt.Errorf("fcitx5 is not enabled")
	}

	logger.Debugf("Switching to input method: %s", targetMethod)

	var err error
	if targetMethod == "english" {
		err = s.switchToEnglish()
	} else {
		err = s.switchToNative(targetMethod)
	}

	if err != nil {
		return err
	}

	if observed := s.GetCurrent(); !verifySwitchReachedTarget(targetMethod, observed) {
		return fmt.Errorf("input method did not switch to %s (still %s)", targetMethod, observed)
	}

	return nil
}

// switchToEnglish puts the input state into English. When Rime is the active
// input method the English state is Rime's ASCII mode: in a fcitx5 group that
// contains Rime alone (a common single-IM setup), Deactivate/SetCurrentIM
// silently do nothing, so ASCII mode is the only way to reach English. For
// other setups the fcitx5 deactivation path is used.
func (s *Switcher) switchToEnglish() error {
	if s.rime != nil && s.fcitx5.GetCurrent() == "rime" {
		err := s.rime.SetAsciiMode(true)
		if err == nil {
			return nil
		}
		logger.Debugf("Failed to enable rime ascii mode, falling back to fcitx5: %v", err)
	}

	return s.fcitx5.SwitchToEnglish()
}

// switchToNative switches to a Rime-backed input method (e.g. chinese,
// japanese): activate Rime, select the target schema, then leave ASCII mode.
// The schema must be selected before clearing ASCII mode because selecting a
// schema resets the ASCII state to that schema's default.
func (s *Switcher) switchToNative(targetMethod string) error {
	if err := s.fcitx5.SwitchToRime(); err != nil {
		return fmt.Errorf("failed to switch to Rime: %w", err)
	}

	// Wait a bit for the switch to take effect
	time.Sleep(100 * time.Millisecond)

	if s.rime == nil {
		return nil
	}

	if err := s.rime.SwitchSchema(targetMethod); err != nil {
		return err
	}

	// Wait for the schema change to settle before clearing ASCII mode, which
	// would otherwise be overwritten by the schema switch.
	time.Sleep(100 * time.Millisecond)

	if err := s.rime.SetAsciiMode(false); err != nil {
		return fmt.Errorf("failed to leave rime ascii mode: %w", err)
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
