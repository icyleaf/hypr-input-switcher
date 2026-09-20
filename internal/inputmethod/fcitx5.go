package inputmethod

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"hypr-input-switcher/pkg/logger"

	"github.com/godbus/dbus/v5"
)

// fcitx5 controller State values, as reported by Controller1.State and
// fcitx5-remote. An inactive controller means the keyboard layout is in effect,
// which is the English state.
const (
	fcitxStateClosed   = 0
	fcitxStateInactive = 1
	fcitxStateActive   = 2
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

// resolveFcitx5InputMethod maps the fcitx5 controller state and the current
// input method name to a high-level input method label. The controller reports
// the name of the configured input method even while it is inactive, so the
// state is what distinguishes an active method from the keyboard layout: only
// an active controller can be a real input method, anything else is English.
func resolveFcitx5InputMethod(state int, currentIM, rimeInputMethod string) string {
	if state != fcitxStateActive {
		return "english"
	}

	if currentIM == rimeInputMethod {
		return "rime"
	}

	return "english"
}

// GetCurrent gets current input method via fcitx5
func (f *Fcitx5) GetCurrent() string {
	return resolveFcitx5InputMethod(f.GetState(), f.getCurrentName(), f.rimeInputMethod)
}

// GetState returns the fcitx5 controller state via D-Bus, falling back to
// fcitx5-remote. fcitxStateClosed is returned when neither can be queried.
func (f *Fcitx5) GetState() int {
	if state, err := f.getStateViaDBus(); err == nil {
		return state
	}

	// Fallback to fcitx5-remote
	cmd := exec.Command("fcitx5-remote")
	output, err := cmd.Output()
	if err != nil {
		return fcitxStateClosed
	}

	state, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		logger.Debugf("Failed to parse fcitx5-remote state %q: %v", string(output), err)
		return fcitxStateClosed
	}

	return state
}

// getStateViaDBus gets the controller state via D-Bus
func (f *Fcitx5) getStateViaDBus() (int, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus: %v", err)
		return fcitxStateClosed, err
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	var state int32
	if err := obj.Call("org.fcitx.Fcitx.Controller1.State", 0).Store(&state); err != nil {
		logger.Debugf("Failed to get fcitx5 state via D-Bus: %v", err)
		return fcitxStateClosed, err
	}

	logger.Debugf("Current fcitx5 state via D-Bus: %d", state)
	return int(state), nil
}

// getCurrentName returns the configured current input method name via D-Bus,
// falling back to fcitx5-remote. It returns "" when neither can be queried.
func (f *Fcitx5) getCurrentName() string {
	if name := f.getCurrentNameViaDBus(); name != "" {
		return name
	}

	// Fallback to fcitx5-remote
	cmd := exec.Command("fcitx5-remote", "-n")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// getCurrentNameViaDBus gets the configured current input method name via D-Bus
func (f *Fcitx5) getCurrentNameViaDBus() string {
	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus: %v", err)
		return ""
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/controller")
	var currentIM string

	err = obj.Call("org.fcitx.Fcitx.Controller1.CurrentInputMethod", 0).Store(&currentIM)
	if err != nil {
		logger.Debugf("Failed to get current input method via D-Bus: %v", err)
		return ""
	}

	logger.Debugf("Current input method via D-Bus: %s", currentIM)
	return currentIM
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

	logger.Debug("Successfully switched to English via D-Bus")
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

	logger.Debug("Successfully switched to Rime via D-Bus")
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
