package inputmethod

import (
	"fmt"
	"os/exec"

	"hypr-input-switcher/pkg/logger"

	"github.com/godbus/dbus/v5"
)

type Rime struct {
	schemas map[string]string // inputMethod -> schema mapping
}

func NewRime(schemas map[string]string) *Rime {
	return &Rime{
		schemas: schemas,
	}
}

// IsAvailable checks if Rime is available
func (r *Rime) IsAvailable() bool {
	// Check if we can connect to Rime via D-Bus
	conn, err := dbus.SessionBus()
	if err != nil {
		return false
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	var schemas []string
	err = obj.Call("org.fcitx.Fcitx.Rime1.ListAllSchemas", 0).Store(&schemas)
	return err == nil
}

// GetCurrentSchema gets current Rime schema
func (r *Rime) GetCurrentSchema() string {
	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus: %v", err)
		return "unknown"
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	var currentSchema string

	err = obj.Call("org.fcitx.Fcitx.Rime1.GetCurrentSchema", 0).Store(&currentSchema)
	if err != nil {
		logger.Debugf("Failed to get current rime schema via D-Bus: %v", err)
		return "unknown"
	}

	logger.Debugf("Current rime schema via D-Bus: %s", currentSchema)
	return currentSchema
}

// resolveRimeInputMethod maps Rime state to a configured input method name.
// ASCII mode means the user is typing Latin characters, so it always resolves
// to "english" regardless of the active schema. Otherwise the schema is mapped
// back through the configured schema table, falling back to defaultIM.
func resolveRimeInputMethod(asciiMode bool, schema string, schemas map[string]string, defaultIM string) string {
	if asciiMode {
		return "english"
	}

	if schema == "unknown" {
		return defaultIM
	}

	for imType, schemaName := range schemas {
		if schemaName == schema {
			return imType
		}
	}

	return defaultIM
}

// IsAsciiMode reports whether Rime is currently in ASCII (Latin) mode, which is
// how a Latin/English input state is represented when the fcitx5 group only
// contains Rime.
func (r *Rime) IsAsciiMode() bool {
	conn, err := dbus.SessionBus()
	if err != nil {
		logger.Debugf("Failed to connect to session bus: %v", err)
		return false
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	var asciiMode bool
	if err := obj.Call("org.fcitx.Fcitx.Rime1.IsAsciiMode", 0).Store(&asciiMode); err != nil {
		logger.Debugf("Failed to query rime ascii mode via D-Bus: %v", err)
		return false
	}

	return asciiMode
}

// SetAsciiMode switches Rime between ASCII (Latin) and native input.
func (r *Rime) SetAsciiMode(asciiMode bool) error {
	logger.Debugf("Setting rime ascii mode to %v", asciiMode)

	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	if err := obj.Call("org.fcitx.Fcitx.Rime1.SetAsciiMode", 0, asciiMode).Err; err != nil {
		return err
	}

	return nil
}

// GetCurrentInputMethod returns the input method type based on current Rime state
func (r *Rime) GetCurrentInputMethod(defaultIM string) string {
	return resolveRimeInputMethod(r.IsAsciiMode(), r.GetCurrentSchema(), r.schemas, defaultIM)
}

// SwitchSchema switches to specified schema
func (r *Rime) SwitchSchema(targetMethod string) error {
	schema, exists := r.schemas[targetMethod]
	if !exists {
		return fmt.Errorf("no schema configured for input method: %s", targetMethod)
	}

	// Try D-Bus first
	if err := r.switchSchemaViaDBus(schema); err == nil {
		logger.Debugf("Successfully switched rime schema to: %s (D-Bus)", schema)
		return nil
	}

	// Fallback to dbus-send
	return r.switchSchemaFallback(schema)
}

// switchSchemaViaDBus switches schema via D-Bus
func (r *Rime) switchSchemaViaDBus(schema string) error {
	logger.Debugf("Switching rime schema to: %s via D-Bus", schema)

	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	call := obj.Call("org.fcitx.Fcitx.Rime1.SetSchema", 0, schema)
	return call.Err
}

// switchSchemaFallback switches schema using fallback method
func (r *Rime) switchSchemaFallback(schema string) error {
	logger.Debugf("Switching rime schema to: %s via fallback method", schema)

	cmd := exec.Command("dbus-send",
		"--type=method_call",
		"--dest=org.fcitx.Fcitx5",
		"/rime",
		"org.fcitx.Fcitx.Rime1.SetSchema",
		fmt.Sprintf("string:%s", schema))

	if err := cmd.Run(); err != nil {
		logger.Warningf("Failed to switch rime schema via dbus-send: %v", err)
		return err
	}

	logger.Debugf("Successfully switched rime schema to: %s (dbus-send)", schema)
	return nil
}

// GetAvailableSchemas returns list of available schemas
func (r *Rime) GetAvailableSchemas() ([]string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	obj := conn.Object("org.fcitx.Fcitx5", "/rime")
	var schemas []string
	err = obj.Call("org.fcitx.Fcitx.Rime1.ListAllSchemas", 0).Store(&schemas)
	return schemas, err
}
