package notification

import (
	"fmt"
	"os/exec"
	"strings"

	"hypr-input-switcher/internal/config"
	"hypr-input-switcher/pkg/logger"

	"github.com/gen2brain/beeep"
)

type Notifier struct {
	config  *config.Config
	methods []string
}

func NewNotifier(config *config.Config) *Notifier {
	return &Notifier{
		config:  config,
		methods: detectMethods(),
	}
}

// Show shows a notification
func (n *Notifier) Show(title, message, icon string) {
	if !n.config.Notifications.Enabled {
		return
	}

	// Try various notification methods
	for _, method := range n.methods {
		if n.send(method, title, message, icon) {
			return
		}
	}

	// Fallback to beeep (cross-platform)
	beeep.Notify(title, message, icon)
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

func detectMethods() []string {
	var methods []string

	notificationTools := map[string]string{
		"notify-send":   "libnotify",
		"dunstify":      "dunst",
		"hyprctl":       "hyprland",
		"mako":          "mako",
		"swaync-client": "swaync",
	}

	for tool, method := range notificationTools {
		if _, err := exec.LookPath(tool); err == nil {
			methods = append(methods, method)
		}
	}

	logger.Infof("Available notification methods: %v", methods)
	return methods
}

func (n *Notifier) send(method, title, message, icon string) bool {
	duration := fmt.Sprintf("%d", n.config.Notifications.Duration)

	switch method {
	case "libnotify":
		cmd := exec.Command("notify-send", "-t", duration, "-i", icon, title, message)
		return cmd.Run() == nil

	case "dunst":
		cmd := exec.Command("dunstify", "-t", duration, "-i", icon, title, message)
		return cmd.Run() == nil

	case "hyprland":
		cmd := exec.Command("hyprctl", "notify", "2", duration, fmt.Sprintf("rgb(ffffff) %s: %s", title, message))
		return cmd.Run() == nil

	default:
		return false
	}
}

func (n *Notifier) getDisplayName(method string) string {
	if displayName, exists := n.config.DisplayNames[method]; exists {
		return displayName
	}
	return strings.Title(method)
}

func (n *Notifier) getIcon(method string) string {
	if icon, exists := n.config.Icons[method]; exists {
		return icon
	}
	return "input-keyboard"
}
