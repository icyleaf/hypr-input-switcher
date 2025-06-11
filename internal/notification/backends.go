package notification

import (
	"fmt"
	"os/exec"
	"time"
)

type NotificationBackend interface {
	Show(title, message string, icon string) error
}

type NotifySendBackend struct {
	duration time.Duration
}

func NewNotifySendBackend(duration time.Duration) *NotifySendBackend {
	return &NotifySendBackend{duration: duration}
}

func (b *NotifySendBackend) Show(title, message string, icon string) error {
	cmd := exec.Command("notify-send", "-t", fmt.Sprintf("%d", int(b.duration.Milliseconds())), "-i", icon, title, message)
	return cmd.Run()
}

type DunstBackend struct {
	duration time.Duration
}

func NewDunstBackend(duration time.Duration) *DunstBackend {
	return &DunstBackend{duration: duration}
}

func (b *DunstBackend) Show(title, message string, icon string) error {
	cmd := exec.Command("dunstify", "-t", fmt.Sprintf("%d", int(b.duration.Milliseconds())), "-i", icon, title, message)
	return cmd.Run()
}

type HyprlandBackend struct {
	duration time.Duration
}

func NewHyprlandBackend(duration time.Duration) *HyprlandBackend {
	return &HyprlandBackend{duration: duration}
}

func (b *HyprlandBackend) Show(title, message string, icon string) error {
	cmd := exec.Command("hyprctl", "notify", "2", fmt.Sprintf("%d", int(b.duration.Milliseconds())), fmt.Sprintf("rgb(ffffff) %s: %s", title, message))
	return cmd.Run()
}

type MakoBackend struct {
	duration time.Duration
}

func NewMakoBackend(duration time.Duration) *MakoBackend {
	return &MakoBackend{duration: duration}
}

func (b *MakoBackend) Show(title, message string, icon string) error {
	cmd := exec.Command("mako", "--timeout", fmt.Sprintf("%d", int(b.duration.Milliseconds())), "--icon", icon, fmt.Sprintf("%s\n%s", title, message))
	return cmd.Run()
}

type NotificationManager struct {
	backends []NotificationBackend
}

func NewNotificationManager(backends []NotificationBackend) *NotificationManager {
	return &NotificationManager{backends: backends}
}

func (m *NotificationManager) ShowNotification(title, message, icon string) {
	for _, backend := range m.backends {
		if err := backend.Show(title, message, icon); err == nil {
			return
		}
	}
}