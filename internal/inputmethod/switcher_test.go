package inputmethod

import (
	"testing"

	"hypr-input-switcher/internal/config"
)

type recordingNotifier struct {
	calls int
}

func (n *recordingNotifier) ShowInputMethodSwitch(string, *config.WindowInfo) {
	n.calls++
}

func TestGetTargetInputMethod(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "english",
		ClientRules: []config.ClientRule{
			{Class: "^code$", Title: `README\.md$`, InputMethod: "chinese"},
			{Class: "^code$", InputMethod: "english"},
			{Class: "google-chrome", InputMethod: config.KeepInputMethod},
			{Class: "firefox", InputMethod: "japanese"},
		},
	})

	tests := []struct {
		name   string
		client *ClientInfo
		want   string
	}{
		{
			name:   "first matching class and title rule wins",
			client: &ClientInfo{Class: "code", Title: "README.md"},
			want:   "chinese",
		},
		{
			name:   "class-only rule is used when title does not match",
			client: &ClientInfo{Class: "code", Title: "main.go"},
			want:   "english",
		},
		{
			name:   "keep rule is returned",
			client: &ClientInfo{Class: "google-chrome", Title: "Issue 6"},
			want:   config.KeepInputMethod,
		},
		{
			name:   "regular expression rule matches",
			client: &ClientInfo{Class: "org.mozilla.firefox", Title: "Home"},
			want:   "japanese",
		},
		{
			name:   "unmatched client uses default",
			client: &ClientInfo{Class: "kitty", Title: "shell"},
			want:   "english",
		},
		{
			name: "nil client uses default",
			want: "english",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := switcher.getTargetInputMethod(tt.client); got != tt.want {
				t.Fatalf("getTargetInputMethod() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessWindowChangeKeepsCurrentInputMethod(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
		client *ClientInfo
	}{
		{
			name: "keep default",
			config: &config.Config{
				DefaultInputMethod: config.KeepInputMethod,
				Notifications:      config.NotificationConfig{ShowOnSwitch: true},
			},
			client: &ClientInfo{Address: "0x1", Class: "kitty", Title: "shell"},
		},
		{
			name: "keep client rule",
			config: &config.Config{
				DefaultInputMethod: "english",
				ClientRules: []config.ClientRule{
					{Class: "google-chrome", InputMethod: config.KeepInputMethod},
				},
				Notifications: config.NotificationConfig{ShowOnSwitch: true},
			},
			client: &ClientInfo{Address: "0x2", Class: "google-chrome", Title: "Issue 6"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switcher := NewSwitcher(tt.config)
			switcher.currentIM = "chinese"
			notifier := &recordingNotifier{}
			switcher.SetNotifier(notifier)

			if err := switcher.processWindowChange(tt.client); err != nil {
				t.Fatalf("processWindowChange() error = %v", err)
			}

			if switcher.currentClient != tt.client {
				t.Fatal("processWindowChange() did not update the current client")
			}
			if switcher.currentIM != "chinese" {
				t.Fatalf("currentIM = %q, want unchanged value %q", switcher.currentIM, "chinese")
			}
			if notifier.calls != 0 {
				t.Fatalf("notification calls = %d, want 0", notifier.calls)
			}
		})
	}
}

func TestSwitchKeepIsNoOp(t *testing.T) {
	switcher := NewSwitcher(&config.Config{})

	if err := switcher.Switch(config.KeepInputMethod); err != nil {
		t.Fatalf("Switch(%q) error = %v", config.KeepInputMethod, err)
	}
}
