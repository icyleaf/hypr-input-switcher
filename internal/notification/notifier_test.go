package notification

import (
	"testing"

	"hypr-input-switcher/internal/config"
)

func TestFormatNotificationContent(t *testing.T) {
	cfg := &config.Config{
		Notifications: config.NotificationConfig{
			ShowAppName: true,
		},
		InputMethods: map[string]config.InputMethodConfig{
			"english": {DisplayName: "English"},
			"chinese": {DisplayName: "中文"},
		},
	}
	cfg.Normalize()
	notifier := &Notifier{config: cfg}

	tests := []struct {
		name        string
		inputMethod string
		windowInfo  *config.WindowInfo
		showAppName bool
		wantTitle   string
		wantMessage string
	}{
		{
			name:        "explicit rule switch with app name",
			inputMethod: "english",
			windowInfo:  &config.WindowInfo{Class: "google-chrome"},
			showAppName: true,
			wantTitle:   "google-chrome",
			wantMessage: "Switched to English",
		},
		{
			name:        "explicit rule switch without app name",
			inputMethod: "english",
			windowInfo:  &config.WindowInfo{Class: "google-chrome"},
			showAppName: false,
			wantTitle:   "Input Method Switched",
			wantMessage: "Switched to English",
		},
		{
			name:        "keep rule kept without history",
			inputMethod: "english",
			windowInfo:  &config.WindowInfo{Class: "ghostty", IsKeep: true, IsRestored: false},
			showAppName: true,
			wantTitle:   "ghostty",
			wantMessage: "Kept as English (keep)",
		},
		{
			name:        "keep rule kept without history and without app name",
			inputMethod: "english",
			windowInfo:  &config.WindowInfo{Class: "ghostty", IsKeep: true, IsRestored: false},
			showAppName: false,
			wantTitle:   "Input Method Kept",
			wantMessage: "Kept as English (keep)",
		},
		{
			name:        "keep rule restored from app history",
			inputMethod: "chinese",
			windowInfo:  &config.WindowInfo{Class: "ghostty", IsKeep: true, IsRestored: true},
			showAppName: true,
			wantTitle:   "ghostty",
			wantMessage: "Restored to 中文 (keep)",
		},
		{
			name:        "keep rule restored from app history without app name",
			inputMethod: "chinese",
			windowInfo:  &config.WindowInfo{Class: "ghostty", IsKeep: true, IsRestored: true},
			showAppName: false,
			wantTitle:   "Input Method Restored",
			wantMessage: "Restored to 中文 (keep)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.Notifications.ShowAppName = tt.showAppName
			gotTitle, gotMsg := notifier.formatNotificationContent(tt.inputMethod, tt.windowInfo)
			if gotTitle != tt.wantTitle {
				t.Errorf("title = %q, want %q", gotTitle, tt.wantTitle)
			}
			if gotMsg != tt.wantMessage {
				t.Errorf("message = %q, want %q", gotMsg, tt.wantMessage)
			}
		})
	}
}
