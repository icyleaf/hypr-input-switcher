package inputmethod

import (
	"os/exec"
	"strings"
	"time"

	"hypr-input-switcher/pkg/logger"

	"github.com/godbus/dbus/v5"
)

type Fcitx5 struct {
	rimeInputMethod string
}

func NewFcitx5(rimeInputMethod string) *Fcitx5 {
	return &Fcitx5{
		rimeInputMethod: rimeInputMethod,
	}
}

// IsAvailable checks if fcitx5 is available
func (f *Fcitx5) IsAvailable() bool {
	// Check if fcitx5-remote is available
	if _, err := exec.LookPath("fcitx5-remote"); err != nil {
		return false
	}

	// Try to get current input method to verify fcitx5 is running
	cmd := exec.Command("fcitx5-remote", "-n")
	_, err := cmd.Output()
	return err == nil
}

// GetCurrent gets current input method via fcitx5
func (f *Fcitx5) GetCurrent() string {
	// Try D-Bus first
	if currentIM := f.getCurrentViaDBus(); currentIM != "unknown" {
		return currentIM
	}

	// Fallback to fcitx5-remote
	cmd := exec.Command("fcitx5-remote", "-n")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	currentIM := strings.TrimSpace(string(output))

	// If it's rime, return the specific identifier
	if currentIM == f.rimeInputMethod {
		return "rime"
	}

	return "english"
}

// getCurrentViaDBus gets current input method via D-Bus
func (f *Fcitx5) getCurrentViaDBus() string {
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

	// If it's rime, return rime identifier
	if currentIM == f.rimeInputMethod {
		return "rime"
	}

	if strings.Contains(currentIM, "keyboard") {
		return "english"
	}

	return "english"
}

// SwitchToEnglish switches to English input method
func (f *Fcitx5) SwitchToEnglish() error {
	// Try D-Bus first
	if err := f.switchToEnglishViaDBus(); err == nil {
		return nil
	}

	// Fallback to fcitx5-remote
	logger.Debug("Using fcitx5-remote fallback for English switch")
	cmd := exec.Command("fcitx5-remote", "-c")
	return cmd.Run()
}

// switchToEnglishViaDBus switches to English via D-Bus
func (f *Fcitx5) switchToEnglishViaDBus() error {
	logger.Debug("Switching to English input method via D-Bus")

	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	call := obj.Call("org.fcitx.Fcitx.Controller1.Deactivate", 0)
	if call.Err != nil {
		return call.Err
	}

	logger.Info("Successfully switched to English via D-Bus")
	return nil
}

// SwitchToRime switches to Rime input method
func (f *Fcitx5) SwitchToRime() error {
	// Try D-Bus first
	if err := f.switchToRimeViaDBus(); err == nil {
		return nil
	}

	// Fallback to fcitx5-remote
	return f.switchToRimeFallback()
}

// switchToRimeViaDBus switches to Rime via D-Bus
func (f *Fcitx5) switchToRimeViaDBus() error {
	logger.Debug("Switching to Rime input method via D-Bus")

	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Step 1: Activate input method
	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	call := obj.Call("org.fcitx.Fcitx.Controller1.Activate", 0)
	if call.Err != nil {
		return call.Err
	}

	time.Sleep(100 * time.Millisecond)

	// Step 2: Switch to rime input method
	call = obj.Call("org.fcitx.Fcitx.Controller1.SetCurrentIM", 0, f.rimeInputMethod)
	if call.Err != nil {
		return call.Err
	}

	logger.Info("Successfully switched to Rime via D-Bus")
	return nil
}

// switchToRimeFallback switches to Rime using fallback method
func (f *Fcitx5) switchToRimeFallback() error {
	logger.Debug("Using fcitx5-remote fallback method for Rime")

	cmd := exec.Command("fcitx5-remote", "-o")
	if err := cmd.Run(); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	cmd = exec.Command("fcitx5-remote", "-s", f.rimeInputMethod)
	return cmd.Run()
}
