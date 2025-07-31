package notification

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"hypr-input-switcher/pkg/logger"
)

// Embed country flag icons into the binary
//
//go:embed icons/*.svg
var embeddedIcons embed.FS

// EmbeddedIconExtractor extracts embedded icons to local directory
type EmbeddedIconExtractor struct {
	iconPath string
}

func NewEmbeddedIconExtractor(iconPath string) *EmbeddedIconExtractor {
	return &EmbeddedIconExtractor{
		iconPath: iconPath,
	}
}

// ExtractEmbeddedIcons extracts embedded icons to local directory
func (e *EmbeddedIconExtractor) ExtractEmbeddedIcons() error {
	// Ensure directory exists
	if err := os.MkdirAll(e.iconPath, 0755); err != nil {
		return fmt.Errorf("failed to create icon directory: %v", err)
	}

	// Read embedded files
	entries, err := embeddedIcons.ReadDir("icons")
	if err != nil {
		return fmt.Errorf("failed to read embedded icons: %v", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no embedded icons found")
	}

	logger.Infof("Extracting %d embedded icons...", len(entries))

	extractedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Read embedded file
		embeddedPath := filepath.Join("icons", entry.Name())
		data, err := embeddedIcons.ReadFile(embeddedPath)
		if err != nil {
			logger.Warningf("Failed to read embedded icon %s: %v", entry.Name(), err)
			continue
		}

		// Write to local file
		localPath := filepath.Join(e.iconPath, entry.Name())
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			logger.Warningf("Failed to write icon %s: %v", entry.Name(), err)
			continue
		}

		logger.Debugf("Extracted icon: %s", entry.Name())
		extractedCount++
	}

	logger.Infof("Successfully extracted %d/%d embedded icons", extractedCount, len(entries))
	return nil
}

// HasEmbeddedIcons checks if embedded icons are available
func (e *EmbeddedIconExtractor) HasEmbeddedIcons() bool {
	entries, err := embeddedIcons.ReadDir("icons")
	return err == nil && len(entries) > 0
}

// ListEmbeddedIcons lists all embedded icon names
func (e *EmbeddedIconExtractor) ListEmbeddedIcons() []string {
	entries, err := embeddedIcons.ReadDir("icons")
	if err != nil {
		return nil
	}

	var icons []string
	for _, entry := range entries {
		if !entry.IsDir() {
			icons = append(icons, entry.Name())
		}
	}
	return icons
}
