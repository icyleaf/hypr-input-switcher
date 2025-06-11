package inputmethod

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"hypr-input-switcher/pkg/logger"
)

type Fcitx5 struct {
	rimeInputMethod string
}

func NewFcitx5(rimeInputMethod string) *Fcitx5 {
	return &Fcitx5{
		rimeInputMethod: rimeInputMethod,
	}
}

func (f *Fcitx5) GetCurrentInputMethod() string {
	cmd := exec.Command("fcitx5-remote", "-n")
	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("Failed to get current input method: %v", err)
		return "unknown"
	}

	currentIM := strings.TrimSpace(string(output))
	logger.Debugf("Current fcitx5 input method: %s", currentIM)

	// If it's rime, try to get current schema
	if currentIM == f.rimeInputMethod {
		return f.getCurrentRimeSchema()
	}

	// If it's keyboard-us or similar, return english
	if strings.Contains(currentIM, "keyboard") {
		return "english"
	}

	return "english" // Default fallback
}

func (f *Fcitx5) getCurrentRimeSchema() string {
	// Try using rime_api first
	cmd := exec.Command("rime_api", "get_current_schema")
	if output, err := cmd.Output(); err == nil {
		schema := strings.TrimSpace(string(output))
		logger.Debugf("Current rime schema: %s", schema)
		return f.mapSchemaToInputMethod(schema)
	}

	// If rime_api is not available, we can't determine the exact schema
	logger.Debug("rime_api not available, returning default")
	return "chinese" // Default assumption
}

func (f *Fcitx5) mapSchemaToInputMethod(schema string) string {
	// This mapping should be configurable, but for now hardcode some common ones
	schemaMap := map[string]string{
		"rime_frost":    "chinese",
		"luna_pinyin":   "chinese",
		"double_pinyin": "chinese",
		"jaroomaji":     "japanese",
		"hiragana":      "japanese",
	}

	if inputMethod, exists := schemaMap[schema]; exists {
		return inputMethod
	}

	// Default fallback based on schema name patterns
	if strings.Contains(strings.ToLower(schema), "japan") ||
		strings.Contains(strings.ToLower(schema), "hiragana") ||
		strings.Contains(strings.ToLower(schema), "katakana") {
		return "japanese"
	}

	if strings.Contains(strings.ToLower(schema), "chin") ||
		strings.Contains(strings.ToLower(schema), "pinyin") {
		return "chinese"
	}

	return "chinese" // Default fallback
}

func (f *Fcitx5) SwitchInputMethod(targetMethod string) error {
	logger.Infof("Switching to input method: %s", targetMethod)

	if targetMethod == "english" {
		return f.switchToEnglish()
	}

	return f.switchToRime(targetMethod)
}

func (f *Fcitx5) switchToEnglish() error {
	logger.Debug("Switching to English input method")
	cmd := exec.Command("fcitx5-remote", "-c")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch to English: %w", err)
	}
	return nil
}

func (f *Fcitx5) switchToRime(targetMethod string) error {
	logger.Debugf("Switching to Rime input method for: %s", targetMethod)

	// First activate input method
	cmd := exec.Command("fcitx5-remote", "-o")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to activate input method: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Switch to rime input method
	cmd = exec.Command("fcitx5-remote", "-s", f.rimeInputMethod)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch to rime: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Switch rime schema
	return f.switchRimeSchema(targetMethod)
}

func (f *Fcitx5) switchRimeSchema(targetMethod string) error {
	// This should use the schema mapping from config
	// For now, use some defaults
	schemaMap := map[string]string{
		"chinese":  "rime_frost",
		"japanese": "jaroomaji",
	}

	schema, exists := schemaMap[targetMethod]
	if !exists {
		logger.Warningf("No schema mapping for method: %s", targetMethod)
		return nil // Don't fail, just keep current schema
	}

	logger.Debugf("Switching Rime schema to: %s", schema)

	// Try using rime_api to switch schema
	cmd := exec.Command("rime_api", "select_schema", schema)
	if err := cmd.Run(); err != nil {
		logger.Warningf("Failed to switch rime schema via rime_api: %v", err)
		// Could try other methods here like writing to rime config files
		return nil // Don't fail the whole operation
	}

	logger.Infof("Successfully switched rime schema to: %s", schema)
	return nil
}

func (f *Fcitx5) IsAvailable() bool {
	// Check if fcitx5-remote is available
	_, err := exec.LookPath("fcitx5-remote")
	return err == nil
}
