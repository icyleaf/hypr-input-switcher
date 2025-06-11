package inputmethod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"hypr-input-switcher/pkg/logger"
)

type RimeConfig struct {
	PreviouslySelectedSchema string `json:"previously_selected_schema"`
}

type Rime struct {
	schemaMapping map[string]string
}

func NewRime(schemaMapping map[string]string) *Rime {
	return &Rime{
		schemaMapping: schemaMapping,
	}
}

// GetCurrentSchema gets the current active rime schema
func (r *Rime) GetCurrentSchema() (string, error) {
	// Try using rime_api first
	cmd := exec.Command("rime_api", "get_current_schema")
	if output, err := cmd.Output(); err == nil {
		schema := strings.TrimSpace(string(output))
		logger.Debugf("Current rime schema via rime_api: %s", schema)
		return schema, nil
	}

	// Then use dbus-send
	if schema, err := r.GetSchemaViaDbus(); err == nil {
		return schema, nil
	}

	// Fallback: try reading from rime config files
	return r.getSchemaFromConfig()
}

func (r *Rime) GetSchemaViaDbus() (string, error) {
	cmd := exec.Command("dbus-send", "--print-reply", "--dest=org.fcitx.Fcitx5.Rime", "/org/fcitx/Fcitx5/Rime", "org.fcitx.Fcitx5.Rime.GetCurrentSchema")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current rime schema via dbus: %w", err)
	}

	schema := strings.TrimSpace(string(output))
	logger.Debugf("Current rime schema via dbus-send: %s", schema)
	return schema, nil
}

// getSchemaFromConfig reads schema from rime config files
func (r *Rime) getSchemaFromConfig() (string, error) {
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime", "user.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "ibus", "rime", "user.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "fcitx5", "rime", "user.yaml"),
	}

	for _, configPath := range configPaths {
		if schema, err := r.readSchemaFromFile(configPath); err == nil && schema != "" {
			logger.Debugf("Found schema from config file %s: %s", configPath, schema)
			return schema, nil
		}
	}

	logger.Debug("Could not determine current rime schema from config files")
	return "", fmt.Errorf("could not determine current rime schema")
}

// readSchemaFromFile reads schema from a specific config file
func (r *Rime) readSchemaFromFile(configPath string) (string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Simple parsing for YAML format
	// Look for "previously_selected_schema: schema_name"
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "previously_selected_schema:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				schema := strings.TrimSpace(parts[1])
				schema = strings.Trim(schema, "\"'") // Remove quotes if present
				return schema, nil
			}
		}
	}

	return "", fmt.Errorf("schema not found in config file")
}

// SwitchSchema switches to the specified rime schema
func (r *Rime) SwitchSchema(schema string) error {
	logger.Infof("Switching rime schema to: %s", schema)

	// Try using rime_api first (fastest method)
	if err := r.switchSchemaViaAPI(schema); err == nil {
		return nil
	}

	// Then try using dbus-send
	if err := r.switchSchemaViaDBus(schema); err == nil {
		return nil
	}

	// Fallback: update config file and deploy
	logger.Warning("rime_api and dbus-send failed, falling back to config file method")
	if err := r.updateConfig(schema); err != nil {
		return fmt.Errorf("failed to update rime config: %w", err)
	}

	return r.deployRime()
}

// switchSchemaViaAPI switches schema using rime_api
func (r *Rime) switchSchemaViaAPI(schema string) error {
	cmd := exec.Command("rime_api", "select_schema", schema)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch schema via rime_api: %w", err)
	}

	logger.Infof("Successfully switched rime schema to: %s (via rime_api)", schema)
	return nil
}

func (r *Rime) switchSchemaViaDBus(schema string) error {
	cmd := exec.Command("dbus-send", "--print-reply", "--dest=org.fcitx.Fcitx5.Rime", "/org/fcitx/Fcitx5/Rime", "org.fcitx.Fcitx5.Rime.SelectSchema", "string:"+schema)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch schema via dbus-send: %w", err)
	}

	logger.Infof("Successfully switched rime schema to: %s (via dbus-send)", schema)
	return nil
}

// updateConfig updates the rime configuration file
func (r *Rime) updateConfig(schema string) error {
	// Try multiple possible config locations
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime", "user.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "ibus", "rime", "user.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "fcitx5", "rime", "user.yaml"),
	}

	var lastErr error
	for _, configPath := range configPaths {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			lastErr = err
			continue
		}

		// Write config file
		if err := r.writeConfigFile(configPath, schema); err != nil {
			lastErr = err
			continue
		}

		logger.Infof("Updated rime config file: %s", configPath)
		return nil
	}

	return fmt.Errorf("failed to update any rime config file: %w", lastErr)
}

// writeConfigFile writes the schema to a config file
func (r *Rime) writeConfigFile(configPath, schema string) error {
	// Read existing config if it exists
	var existingLines []string
	if data, err := os.ReadFile(configPath); err == nil {
		existingLines = strings.Split(string(data), "\n")
	}

	// Update or add the schema line
	updated := false
	for i, line := range existingLines {
		if strings.HasPrefix(strings.TrimSpace(line), "previously_selected_schema:") {
			existingLines[i] = fmt.Sprintf("previously_selected_schema: %s", schema)
			updated = true
			break
		}
	}

	if !updated {
		existingLines = append(existingLines, fmt.Sprintf("previously_selected_schema: %s", schema))
	}

	// Write back to file
	content := strings.Join(existingLines, "\n")
	return os.WriteFile(configPath, []byte(content), 0644)
}

// deployRime deploys rime configuration
func (r *Rime) deployRime() error {
	logger.Debug("Deploying rime configuration")

	// Try different deploy commands
	deployCommands := [][]string{
		{"rime_deployer", "--build"},
		{"fcitx5-remote", "-r"},
		{"ibus-daemon", "-r", "-d"},
	}

	var lastErr error
	for _, cmd := range deployCommands {
		if _, err := exec.LookPath(cmd[0]); err != nil {
			continue // Command not available
		}

		execCmd := exec.Command(cmd[0], cmd[1:]...)
		if err := execCmd.Run(); err != nil {
			lastErr = err
			logger.Debugf("Deploy command %v failed: %v", cmd, err)
			continue
		}

		logger.Infof("Successfully deployed rime using: %v", cmd)
		return nil
	}

	logger.Warning("No rime deploy command succeeded, configuration may not take effect immediately")
	return lastErr // Return last error, but don't fail the operation
}

// IsAvailable checks if rime is available
func (r *Rime) IsAvailable() bool {
	// Check if rime_api is available
	if _, err := exec.LookPath("rime_api"); err == nil {
		return true
	}

	// Check if rime config directories exist
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime"),
		filepath.Join(os.Getenv("HOME"), ".config", "ibus", "rime"),
		filepath.Join(os.Getenv("HOME"), ".config", "fcitx5", "rime"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
