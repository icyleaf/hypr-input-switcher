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

func TestGetTargetInputMethodLayerRules(t *testing.T) {
	newSwitcher := func() *Switcher {
		return NewSwitcher(&config.Config{
			DefaultInputMethod: "english",
			LayerRules: []config.LayerRule{
				{Namespace: "icyleaf-calculator", InputMethod: "english"},
				{Namespace: "omarchy-emojis", InputMethod: "japanese"},
			},
			ClientRules: []config.ClientRule{
				{Class: "^kitty$", InputMethod: "chinese"},
			},
		})
	}

	kitty := &ClientInfo{Class: "kitty", Title: "shell"}

	tests := []struct {
		name   string
		open   []string
		client *ClientInfo
		want   string
	}{
		{
			name:   "open layer overrides matching client rule",
			open:   []string{"icyleaf-calculator"},
			client: kitty,
			want:   "english",
		},
		{
			name:   "no open layer falls back to client rule",
			open:   nil,
			client: kitty,
			want:   "chinese",
		},
		{
			name:   "unmatched open layer does not override client rule",
			open:   []string{"omarchy-osd"},
			client: kitty,
			want:   "chinese",
		},
		{
			name:   "nil client uses layer rule when layer is open",
			open:   []string{"omarchy-emojis"},
			client: nil,
			want:   "japanese",
		},
		{
			name:   "first matching layer rule wins",
			open:   []string{"omarchy-emojis", "icyleaf-calculator"},
			client: kitty,
			want:   "english",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switcher := newSwitcher()
			for _, ns := range tt.open {
				switcher.layers.open(ns)
			}
			if got := switcher.getTargetInputMethod(tt.client); got != tt.want {
				t.Fatalf("getTargetInputMethod() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTargetInputMethodLayerNamespaceIsExactMatch(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "english",
		LayerRules: []config.LayerRule{
			{Namespace: "bar", InputMethod: "japanese"},
		},
		ClientRules: []config.ClientRule{
			{Class: "^kitty$", InputMethod: "chinese"},
		},
	})

	switcher.layers.open("omarchy-bar")

	if got := switcher.getTargetInputMethod(&ClientInfo{Class: "kitty"}); got != "chinese" {
		t.Fatalf("getTargetInputMethod() = %q, want %q (namespace must match exactly)", got, "chinese")
	}
}

func TestGetTargetInputMethodLayerRuleKeep(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "english",
		LayerRules: []config.LayerRule{
			{Namespace: "icyleaf-calculator", InputMethod: config.KeepInputMethod},
		},
		ClientRules: []config.ClientRule{
			{Class: "^kitty$", InputMethod: "chinese"},
		},
	})

	switcher.layers.open("icyleaf-calculator")

	if got := switcher.getTargetInputMethod(&ClientInfo{Class: "kitty"}); got != config.KeepInputMethod {
		t.Fatalf("getTargetInputMethod() = %q, want %q", got, config.KeepInputMethod)
	}
}

func TestLayerRefcountGatesOverride(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "english",
		LayerRules: []config.LayerRule{
			{Namespace: "icyleaf-calculator", InputMethod: "english"},
		},
		ClientRules: []config.ClientRule{
			{Class: "^kitty$", InputMethod: "chinese"},
		},
	})
	client := &ClientInfo{Class: "kitty", Title: "shell"}

	switcher.layers.open("icyleaf-calculator")
	switcher.layers.open("icyleaf-calculator")
	if got := switcher.getTargetInputMethod(client); got != "english" {
		t.Fatalf("with two instances open: getTargetInputMethod() = %q, want %q", got, "english")
	}

	switcher.layers.close("icyleaf-calculator")
	if got := switcher.getTargetInputMethod(client); got != "english" {
		t.Fatalf("with one instance still open: getTargetInputMethod() = %q, want %q", got, "english")
	}

	switcher.layers.close("icyleaf-calculator")
	if got := switcher.getTargetInputMethod(client); got != "chinese" {
		t.Fatalf("with all instances closed: getTargetInputMethod() = %q, want %q", got, "chinese")
	}
}

func TestLayerTrackerIgnoresUnknownClose(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "english",
		LayerRules: []config.LayerRule{
			{Namespace: "icyleaf-calculator", InputMethod: "english"},
		},
		ClientRules: []config.ClientRule{
			{Class: "^kitty$", InputMethod: "chinese"},
		},
	})

	switcher.layers.close("never-opened")

	if got := switcher.getTargetInputMethod(&ClientInfo{Class: "kitty"}); got != "chinese" {
		t.Fatalf("getTargetInputMethod() = %q, want %q", got, "chinese")
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

func TestVerifySwitchReachedTarget(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		observed    string
		wantReached bool
	}{
		{name: "exact match", target: "english", observed: "english", wantReached: true},
		{name: "rime schema match", target: "chinese", observed: "chinese", wantReached: true},
		{name: "mismatch surfaces failure", target: "english", observed: "rime", wantReached: false},
		{name: "unknown is not success", target: "english", observed: "unknown", wantReached: false},
		{name: "empty observation is not success", target: "english", observed: "", wantReached: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifySwitchReachedTarget(tt.target, tt.observed); got != tt.wantReached {
				t.Fatalf("verifySwitchReachedTarget(%q, %q) = %v, want %v", tt.target, tt.observed, got, tt.wantReached)
			}
		})
	}
}
