package notification

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"hypr-input-switcher/internal/config"
	"hypr-input-switcher/pkg/logger"

	"github.com/gen2brain/beeep"
)

type Notifier struct {
	config            *config.Config
	availableMethods  []string
	selectedMethod    string
	iconPath          string
	embeddedExtractor *EmbeddedIconExtractor
}

func NewNotifier(config *config.Config) *Notifier {
	notifier := &Notifier{
		config: config,
	}

	// Setup icon path
	notifier.setupIconPath()

	// Initialize embedded icon extractor
	notifier.embeddedExtractor = NewEmbeddedIconExtractor(notifier.iconPath)

	// Extract embedded icons
	go notifier.ensureIconsAvailable()

	// Detect available notification methods
	notifier.detectMethods()

	return notifier
}

// setupIconPath sets up the icon path
func (n *Notifier) setupIconPath() {
	if n.config.Notifications.IconPath != "" {
		// Expand environment variables and tilde
		iconPath := os.ExpandEnv(n.config.Notifications.IconPath)

		// Handle tilde expansion manually
		if strings.HasPrefix(iconPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				iconPath = filepath.Join(homeDir, iconPath[2:])
			}
		}

		n.iconPath = iconPath
		logger.Debugf("Using configured icon path: %s", n.iconPath)
		return
	}

	// Default icon paths
	defaultPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".local/share/hypr-input-switcher/icons"),
		"./icons",
		"/usr/share/hypr-input-switcher/icons",
	}

	for _, path := range defaultPaths {
		// Try to create directory
		if err := os.MkdirAll(path, 0755); err == nil {
			n.iconPath = path
			logger.Debugf("Using icon path: %s", path)
			return
		}
	}

	// If all failed, use temporary directory
	n.iconPath = filepath.Join(os.TempDir(), "hypr-input-switcher-icons")
	os.MkdirAll(n.iconPath, 0755)
	logger.Warningf("Using temporary icon path: %s", n.iconPath)
}

// ensureIconsAvailable ensures icons are available
func (n *Notifier) ensureIconsAvailable() {
	// Extract embedded icons
	if err := n.embeddedExtractor.ExtractEmbeddedIcons(); err != nil {
		logger.Warningf("Failed to extract embedded icons: %v", err)
		logger.Info("Will use emoji fallback for notifications")
	} else {
		logger.Info("Successfully extracted embedded icons")
	}
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
		title = appName
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

	// Check if icon is an image file
	isImageIcon := n.isImageFile(icon)

	switch method {
	case "notify-send":
		if isImageIcon {
			// Use image icon
			cmd = exec.Command("notify-send", "-t", duration, "-i", icon, title, message)
		} else if n.isEmoji(icon) {
			// Use emoji
			titleWithIcon := fmt.Sprintf("%s %s", icon, title)
			cmd = exec.Command("notify-send", "-t", duration, titleWithIcon, message)
		} else {
			// Use text icon
			cmd = exec.Command("notify-send", "-t", duration, "-i", icon, title, message)
		}

	case "dunstify":
		if isImageIcon {
			cmd = exec.Command("dunstify", "-t", duration, "-i", icon, title, message)
		} else if n.isEmoji(icon) {
			titleWithIcon := fmt.Sprintf("%s %s", icon, title)
			cmd = exec.Command("dunstify", "-t", duration, titleWithIcon, message)
		} else {
			cmd = exec.Command("dunstify", "-t", duration, "-i", icon, title, message)
		}

	case "hyprctl":
		// Hyprland notification doesn't support images, use emoji or text
		var notificationText string
		if n.isEmoji(icon) {
			notificationText = fmt.Sprintf("%s %s: %s", icon, title, message)
		} else {
			notificationText = fmt.Sprintf("%s: %s", title, message)
		}
		cmd = exec.Command("hyprctl", "notify", "2", duration, "0", notificationText)

	case "swaync-client":
		var fullMessage string
		if n.isEmoji(icon) {
			fullMessage = fmt.Sprintf("%s %s: %s", icon, title, message)
		} else {
			fullMessage = fmt.Sprintf("%s: %s", title, message)
		}
		cmd = exec.Command("swaync-client", "-t", "-m", fullMessage)

	case "mako":
		// Mako via notify-send interface, supports image icons
		if isImageIcon {
			cmd = exec.Command("notify-send", "-t", duration, "-i", icon, title, message)
		} else if n.isEmoji(icon) {
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

	logger.Debugf("Notification sent via %s: %s - %s (icon: %s)", method, title, message, icon)
	return true
}

// isImageFile checks if the path is an image file
func (n *Notifier) isImageFile(path string) bool {
	if path == "" {
		return false
	}

	// Check if it's an absolute path or relative path file
	if filepath.IsAbs(path) || filepath.Dir(path) != "." {
		ext := filepath.Ext(path)
		imageExts := []string{".png", ".svg", ".jpg", ".jpeg", ".ico", ".gif", ".bmp"}

		for _, imgExt := range imageExts {
			if ext == imgExt {
				// Confirm file exists
				if _, err := os.Stat(path); err == nil {
					return true
				}
			}
		}
	}

	return false
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
	hasEmbedded := n.embeddedExtractor.HasEmbeddedIcons()
	embeddedIcons := n.embeddedExtractor.ListEmbeddedIcons()

	return map[string]interface{}{
		"enabled":           n.config.Notifications.Enabled,
		"available_methods": n.availableMethods,
		"selected_method":   n.selectedMethod,
		"force_method":      n.config.Notifications.ForceMethod,
		"disabled_methods":  n.config.Notifications.DisabledMethods,
		"icon_path":         n.iconPath,
		"has_embedded":      hasEmbedded,
		"embedded_icons":    embeddedIcons,
		"embedded_count":    len(embeddedIcons),
	}
}

func (n *Notifier) getDisplayName(method string) string {
	if displayName, exists := n.config.DisplayNames[method]; exists {
		return displayName
	}

	return method
}

// findIconFile finds icon file
func (n *Notifier) findIconFile(method string) string {
	if n.iconPath == "" {
		return ""
	}

	// Supported image formats
	extensions := []string{".png", ".svg", ".jpg", ".jpeg", ".ico"}

	// Possible file names
	possibleNames := []string{
		method,
		n.getLanguageCode(method),
		n.getCountryCode(method),
	}

	for _, name := range possibleNames {
		if name == "" {
			continue
		}

		for _, ext := range extensions {
			iconFile := filepath.Join(n.iconPath, name+ext)
			if _, err := os.Stat(iconFile); err == nil {
				logger.Debugf("Found icon file: %s", iconFile)
				return iconFile
			}
		}
	}

	return ""
}

func (n *Notifier) getIcon(method string) string {
	logger.Debugf("Getting icon for method: %s", method)

	// First check if there's a custom icon in config
	if icon, exists := n.config.Icons[method]; exists {
		logger.Debugf("Found custom icon in config: %s", icon)

		// Check if it's a filename (not emoji or absolute path)
		if n.isImageFileName(icon) {
			// Try to find the file in icon path
			iconFile := n.findIconFileByName(icon)
			if iconFile != "" {
				logger.Debugf("Found configured icon file: %s", iconFile)
				return iconFile
			} else {
				logger.Warningf("Configured icon file not found: %s, falling back to emoji", icon)
				// Fall through to emoji fallback
			}
		} else {
			// It's either emoji, absolute path, or system icon name
			return icon
		}
	}

	// Try to find image file by method name
	if n.iconPath != "" {
		iconFile := n.findIconFile(method)
		if iconFile != "" {
			logger.Debugf("Found icon file by method: %s", iconFile)
			return iconFile
		} else {
			logger.Debugf("No icon file found for method: %s", method)
		}
	}

	// If no image file found, use emoji as fallback
	logger.Debugf("Using emoji fallback for method: %s", method)
	return n.getEmojiIcon(method)
}

// isImageFileName checks if the string looks like an image filename
func (n *Notifier) isImageFileName(s string) bool {
	if s == "" {
		return false
	}

	// Check if it's emoji
	if n.isEmoji(s) {
		return false
	}

	// Check if it's an absolute path
	if filepath.IsAbs(s) {
		return false
	}

	// Check if it has image file extension
	ext := strings.ToLower(filepath.Ext(s))
	imageExts := []string{".png", ".svg", ".jpg", ".jpeg", ".ico", ".gif", ".bmp"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}

	return false
}

// findIconFileByName finds icon file by exact filename in icon path
func (n *Notifier) findIconFileByName(filename string) string {
	if n.iconPath == "" || filename == "" {
		return ""
	}

	// First try exact filename
	fullPath := filepath.Join(n.iconPath, filename)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	}

	// If not found, try finding with different extensions
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	extensions := []string{".png", ".svg", ".jpg", ".jpeg", ".ico", ".gif", ".bmp"}

	for _, ext := range extensions {
		testPath := filepath.Join(n.iconPath, baseName+ext)
		if _, err := os.Stat(testPath); err == nil {
			logger.Debugf("Found icon with different extension: %s -> %s", filename, testPath)
			return testPath
		}
	}

	return ""
}

// getEmojiIcon returns the emoji icon for a method
func (n *Notifier) getEmojiIcon(method string) string {
	switch method {
	case "english", "en", "us":
		return "🇺🇸"
	case "chinese", "zh", "cn":
		return "🇨🇳"
	case "japanese", "ja", "jp":
		return "🇯🇵"
	case "korean", "ko", "kr":
		return "🇰🇷"
	case "german", "de":
		return "🇩🇪"
	case "french", "fr":
		return "🇫🇷"
	case "spanish", "es":
		return "🇪🇸"
	case "russian", "ru":
		return "🇷🇺"
	case "arabic", "ar":
		return "🇸🇦"
	case "hindi", "hi", "in":
		return "🇮🇳"
	default:
		return "🌐"
	}
}

// getLanguageCode gets language code
func (n *Notifier) getLanguageCode(method string) string {
	languageMap := map[string]string{
		"english":  "en",
		"chinese":  "zh",
		"japanese": "ja",
		"korean":   "ko",
		"german":   "de",
		"french":   "fr",
		"spanish":  "es",
		"russian":  "ru",
		"arabic":   "ar",
		"hindi":    "hi",
	}

	if code, exists := languageMap[method]; exists {
		return code
	}

	return method
}

// getCountryCode gets country code
func (n *Notifier) getCountryCode(method string) string {
	countryMap := map[string]string{
		"english":  "us",
		"chinese":  "cn",
		"japanese": "jp",
		"korean":   "kr",
		"german":   "de",
		"french":   "fr",
		"spanish":  "es",
		"russian":  "ru",
		"arabic":   "sa",
		"hindi":    "in",
		"en":       "us",
		"zh":       "cn",
		"ja":       "jp",
		"ko":       "kr",
	}

	if code, exists := countryMap[method]; exists {
		return code
	}

	return ""
}
