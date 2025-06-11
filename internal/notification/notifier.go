package notification

import (
	"fmt"
	"os/exec"
	"unicode"

	"hypr-input-switcher/internal/config"
	"hypr-input-switcher/pkg/logger"

	"github.com/gen2brain/beeep"
)

type Notifier struct {
	config           *config.Config
	availableMethods []string
	selectedMethod   string
}

func NewNotifier(config *config.Config) *Notifier {
	notifier := &Notifier{
		config: config,
	}

	// Detect available notification methods
	notifier.detectMethods()

	return notifier
}

// Show shows a notification
func (n *Notifier) Show(title, message, icon string) {
	if !n.config.Notifications.Enabled {
		return
	}

	// If force method is specified, use it
	if n.config.Notifications.ForceMethod != "" {
		if n.send(n.config.Notifications.ForceMethod, title, message, icon) {
			return
		}
		logger.Warningf("Force method %s failed, falling back to auto-detection", n.config.Notifications.ForceMethod)
	}

	// Try sending notification in configured priority order
	for _, method := range n.availableMethods {
		if n.send(method, title, message, icon) {
			return
		}
	}

	// Final fallback: use beeep (cross-platform)
	logger.Debug("All notification methods failed, using beeep fallback")

	// For beeep, if it's emoji, add to title
	if n.isEmoji(icon) {
		titleWithIcon := fmt.Sprintf("%s %s", icon, title)
		beeep.Notify(titleWithIcon, message, "")
	} else {
		beeep.Notify(title, message, icon)
	}
}

// ShowInputMethodSwitch shows input method switch notification
func (n *Notifier) ShowInputMethodSwitch(inputMethod string, clientInfo *config.WindowInfo) {
	displayName := n.getDisplayName(inputMethod)
	icon := n.getIcon(inputMethod)

	title := "Input Method Switched"
	message := fmt.Sprintf("Switched to %s", displayName)

	// If configured to show app name
	if n.config.Notifications.ShowAppName && clientInfo != nil {
		appName := clientInfo.Class
		if appName == "" {
			appName = "Unknown"
		}
		message += fmt.Sprintf(" for %s", appName)
	}

	n.Show(title, message, icon)
}

// detectMethods detects available notification methods
func (n *Notifier) detectMethods() {
	configMethods := n.config.Notifications.Methods
	if len(configMethods) == 0 {
		// Use default method list
		configMethods = []string{"notify-send", "dunstify", "hyprctl", "swaync-client", "mako"}
	}

	disabledMethods := make(map[string]bool)
	for _, method := range n.config.Notifications.DisabledMethods {
		disabledMethods[method] = true
	}

	var availableMethods []string

	for _, method := range configMethods {
		// Skip disabled methods
		if disabledMethods[method] {
			logger.Debugf("Notification method %s is disabled", method)
			continue
		}

		if n.isMethodAvailable(method) {
			availableMethods = append(availableMethods, method)
			logger.Debugf("Notification method %s is available", method)
		} else {
			logger.Debugf("Notification method %s is not available", method)
		}
	}

	n.availableMethods = availableMethods

	if len(availableMethods) > 0 {
		n.selectedMethod = availableMethods[0]
		logger.Infof("Available notification methods: %v, selected: %s", availableMethods, n.selectedMethod)
	} else {
		logger.Warning("No notification methods available, will use fallback")
	}
}

// isMethodAvailable checks if a notification method is available
func (n *Notifier) isMethodAvailable(method string) bool {
	switch method {
	case "notify-send":
		return n.commandExists("notify-send")
	case "dunstify":
		return n.commandExists("dunstify")
	case "hyprctl":
		return n.commandExists("hyprctl") && n.isHyprlandRunning()
	case "swaync-client":
		return n.commandExists("swaync-client")
	case "mako":
		return n.commandExists("mako") && n.isMakoRunning()
	default:
		return false
	}
}

// commandExists checks if a command exists in PATH
func (n *Notifier) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// isHyprlandRunning checks if Hyprland is running
func (n *Notifier) isHyprlandRunning() bool {
	cmd := exec.Command("hyprctl", "version")
	return cmd.Run() == nil
}

// isMakoRunning checks if mako is running
func (n *Notifier) isMakoRunning() bool {
	cmd := exec.Command("pgrep", "mako")
	return cmd.Run() == nil
}

// send sends notification using specified method
func (n *Notifier) send(method, title, message, icon string) bool {
	duration := fmt.Sprintf("%d", n.config.Notifications.Duration)

	var cmd *exec.Cmd

	switch method {
	case "notify-send":
		// For notify-send, if it's emoji, use as title prefix; otherwise as icon
		if n.isEmoji(icon) {
			titleWithIcon := fmt.Sprintf("%s %s", icon, title)
			cmd = exec.Command("notify-send", "-t", duration, titleWithIcon, message)
		} else {
			cmd = exec.Command("notify-send", "-t", duration, "-i", icon, title, message)
		}

	case "dunstify":
		// Dunst also supports similar handling
		if n.isEmoji(icon) {
			titleWithIcon := fmt.Sprintf("%s %s", icon, title)
			cmd = exec.Command("dunstify", "-t", duration, titleWithIcon, message)
		} else {
			cmd = exec.Command("dunstify", "-t", duration, "-i", icon, title, message)
		}

	case "hyprctl":
		// Hyprland native notification, add emoji to message
		var notificationText string
		if n.isEmoji(icon) {
			notificationText = fmt.Sprintf("rgb(ffffff) %s %s: %s", icon, title, message)
		} else {
			notificationText = fmt.Sprintf("rgb(ffffff) %s: %s", title, message)
		}
		cmd = exec.Command("hyprctl", "notify", "2", duration, notificationText)

	case "swaync-client":
		// SwayNC supports emoji
		var fullMessage string
		if n.isEmoji(icon) {
			fullMessage = fmt.Sprintf("%s %s: %s", icon, title, message)
		} else {
			fullMessage = fmt.Sprintf("%s: %s", title, message)
		}
		cmd = exec.Command("swaync-client", "-t", "-m", fullMessage)

	case "mako":
		// Mako via notify-send interface, same handling as notify-send
		if n.isEmoji(icon) {
			titleWithIcon := fmt.Sprintf("%s %s", icon, title)
			cmd = exec.Command("notify-send", "-t", duration, titleWithIcon, message)
		} else {
			cmd = exec.Command("notify-send", "-t", duration, "-i", icon, title, message)
		}

	default:
		logger.Warningf("Unknown notification method: %s", method)
		return false
	}

	if err := cmd.Run(); err != nil {
		logger.Debugf("Notification method %s failed: %v", method, err)
		return false
	}

	logger.Debugf("Notification sent via %s: %s - %s", method, title, message)
	return true
}

// isEmoji checks if the given string contains emoji characters
func (n *Notifier) isEmoji(s string) bool {
	if s == "" {
		return false
	}

	// Check if the string contains emoji characters
	for _, r := range s {
		// Check for various emoji ranges
		if (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
			(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
			(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
			(r >= 0x1F1E0 && r <= 0x1F1FF) || // Regional indicators (flags)
			(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
			(r >= 0x2700 && r <= 0x27BF) || // Dingbats
			(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
			(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
			(r >= 0x1F018 && r <= 0x1F270) || // Various asian characters like 🈚️, 🈳
			unicode.Is(unicode.So, r) { // Other symbols
			return true
		}
	}

	return false
}

// GetStatus returns notification system status
func (n *Notifier) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":           n.config.Notifications.Enabled,
		"available_methods": n.availableMethods,
		"selected_method":   n.selectedMethod,
		"force_method":      n.config.Notifications.ForceMethod,
		"disabled_methods":  n.config.Notifications.DisabledMethods,
	}
}

func (n *Notifier) getDisplayName(method string) string {
	if displayName, exists := n.config.DisplayNames[method]; exists {
		return displayName
	}

	return method
}

func (n *Notifier) getIcon(method string) string {
	if icon, exists := n.config.Icons[method]; exists {
		return icon
	}
	// Default emoji icon
	switch method {
	case "english":
		return "🇺🇸"
	case "chinese":
		return "🇨🇳"
	case "japanese":
		return "🇯🇵"
	default:
		return "🇺🇸"
	}
}
