package notification

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

type Notification struct {
	Enabled bool
	Duration time.Duration
}

func NewNotification(enabled bool, duration time.Duration) *Notification {
	return &Notification{
		Enabled: enabled,
		Duration: duration,
	}
}

func (n *Notification) Show(title, message, icon string) {
	if !n.Enabled {
		return
	}

	// Try various notification methods
	methods := []string{"notify-send", "dunstify", "hyprctl", "mako", "swaync-client"}
	for _, method := range methods {
		if n.sendNotification(method, title, message, icon) {
			return
		}
		logrus.Debugf("Notification method %s failed", method)
	}
}

func (n *Notification) sendNotification(method, title, message, icon string) bool {
	var cmd *exec.Cmd

	switch method {
	case "notify-send":
		cmd = exec.Command("notify-send", "-t", fmt.Sprintf("%d", int(n.Duration.Milliseconds())), "-i", icon, title, message)
	case "dunstify":
		cmd = exec.Command("dunstify", "-t", fmt.Sprintf("%d", int(n.Duration.Milliseconds())), "-i", icon, title, message)
	case "hyprctl":
		cmd = exec.Command("hyprctl", "notify", "2", fmt.Sprintf("%d", int(n.Duration.Milliseconds())), fmt.Sprintf("rgb(ffffff) %s: %s", title, message))
	case "mako":
		cmd = exec.Command("mako", "--timeout", fmt.Sprintf("%d", int(n.Duration.Milliseconds())), "--icon", icon, fmt.Sprintf("%s\n%s", title, message))
	case "swaync-client":
		cmd = exec.Command("swaync-client", "-t", "-e", fmt.Sprintf("%d", int(n.Duration.Milliseconds())), "-i", icon, "-s", title, "-b", message)
	default:
		return false
	}

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Failed to send notification using %s: %v", method, err)
		return false
	}

	return true
}