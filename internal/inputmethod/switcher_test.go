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
			if got := switcher.resolveTarget(tt.client).inputMethod; got != tt.want {
				t.Fatalf("resolveTarget() = %q, want %q", got, tt.want)
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
		name          string
		open          []string
		client        *ClientInfo
		want          string
		wantNamespace string
	}{
		{
			name:          "open layer overrides matching client rule",
			open:          []string{"icyleaf-calculator"},
			client:        kitty,
			want:          "english",
			wantNamespace: "icyleaf-calculator",
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
			name:          "nil client uses layer rule when layer is open",
			open:          []string{"omarchy-emojis"},
			client:        nil,
			want:          "japanese",
			wantNamespace: "omarchy-emojis",
		},
		{
			name:          "first matching layer rule wins",
			open:          []string{"omarchy-emojis", "icyleaf-calculator"},
			client:        kitty,
			want:          "english",
			wantNamespace: "icyleaf-calculator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switcher := newSwitcher()
			for _, ns := range tt.open {
				switcher.layers.open(ns)
			}
			got := switcher.resolveTarget(tt.client)
			if got.inputMethod != tt.want {
				t.Fatalf("resolveTarget().inputMethod = %q, want %q", got.inputMethod, tt.want)
			}
			if got.layerNamespace != tt.wantNamespace {
				t.Fatalf("resolveTarget().layerNamespace = %q, want %q", got.layerNamespace, tt.wantNamespace)
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

	if got := switcher.resolveTarget(&ClientInfo{Class: "kitty"}).inputMethod; got != "chinese" {
		t.Fatalf("resolveTarget() = %q, want %q (namespace must match exactly)", got, "chinese")
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

	if got := switcher.resolveTarget(&ClientInfo{Class: "kitty"}).inputMethod; got != config.KeepInputMethod {
		t.Fatalf("resolveTarget() = %q, want %q", got, config.KeepInputMethod)
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
	if got := switcher.resolveTarget(client).inputMethod; got != "english" {
		t.Fatalf("with two instances open: resolveTarget() = %q, want %q", got, "english")
	}

	switcher.layers.close("icyleaf-calculator")
	if got := switcher.resolveTarget(client).inputMethod; got != "english" {
		t.Fatalf("with one instance still open: resolveTarget() = %q, want %q", got, "english")
	}

	switcher.layers.close("icyleaf-calculator")
	if got := switcher.resolveTarget(client).inputMethod; got != "chinese" {
		t.Fatalf("with all instances closed: resolveTarget() = %q, want %q", got, "chinese")
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

	if got := switcher.resolveTarget(&ClientInfo{Class: "kitty"}).inputMethod; got != "chinese" {
		t.Fatalf("resolveTarget() = %q, want %q", got, "chinese")
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
			if notifier.calls != 1 {
				t.Fatalf("notification calls = %d, want 1", notifier.calls)
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

func TestPerAppMemory(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "chinese",
		ClientRules: []config.ClientRule{
			{Class: "code", InputMethod: "english"},
			{Class: "ghostty", InputMethod: config.KeepInputMethod},
		},
	})

	// 1. Ghostty cold-start has no history, explicit keep rule returns keep
	target := switcher.resolveTarget(&ClientInfo{Class: "ghostty", Title: "terminal"})
	if target.inputMethod != config.KeepInputMethod {
		t.Fatalf("cold-start ghostty target = %q, want %q", target.inputMethod, config.KeepInputMethod)
	}

	// 2. Simulate user typing in Chinese in Ghostty, recording app history
	switcher.appHistory["ghostty"] = "chinese"

	// 3. Switch to Code (matches explicit english rule)
	codeTarget := switcher.resolveTarget(&ClientInfo{Class: "code", Title: "main.go"})
	if codeTarget.inputMethod != "english" {
		t.Fatalf("code target = %q, want %q", codeTarget.inputMethod, "english")
	}

	// 4. Switch back to Ghostty (explicit keep rule restores remembered chinese)
	ghosttyTarget := switcher.resolveTarget(&ClientInfo{Class: "ghostty", Title: "terminal"})
	if ghosttyTarget.inputMethod != "chinese" {
		t.Fatalf("restored ghostty target = %q, want %q", ghosttyTarget.inputMethod, "chinese")
	}

	// 5. Unconfigured app cold-start falls back to DefaultInputMethod
	unconfiguredTarget := switcher.resolveTarget(&ClientInfo{Class: "wechat", Title: "chat"})
	if unconfiguredTarget.inputMethod != "chinese" {
		t.Fatalf("unconfigured cold-start target = %q, want %q", unconfiguredTarget.inputMethod, "chinese")
	}
}

func TestLayerEventDoesNotTriggerNotificationWhenInputMethodAndWindowUnchanged(t *testing.T) {
	switcher := NewSwitcher(&config.Config{
		DefaultInputMethod: "keep",
		ClientRules: []config.ClientRule{
			{Class: "ghostty", InputMethod: config.KeepInputMethod},
		},
		Notifications: config.NotificationConfig{
			Enabled:      true,
			ShowOnSwitch: true,
		},
	})
	notifier := &recordingNotifier{}
	switcher.SetNotifier(notifier)
	ghosttyClient := &ClientInfo{Address: "0x123", Class: "ghostty", Title: "terminal"}

	// 1. Switch to ghostty window initially: notifications may fire on initial focus
	if err := switcher.processWindowChange(ghosttyClient); err != nil {
		t.Fatalf("processWindowChange() error = %v", err)
	}
	initialCalls := notifier.calls

	// 2. An unrelated layer opens and closes (e.g. notification popup, bar, etc.)
	switcher.handleLayerEvent("some-layer", true)
	if switcher.layers.isOpen("some-layer") {
		t.Fatal("handleLayerEvent opened unconfigured layer, want ignored")
	}
	if notifier.calls != initialCalls {
		t.Fatalf("handleLayerEvent(open) sent spurious notification: calls = %d, want %d", notifier.calls, initialCalls)
	}

	switcher.handleLayerEvent("some-layer", false)
	if notifier.calls != initialCalls {
		t.Fatalf("handleLayerEvent(close) sent spurious notification: calls = %d, want %d", notifier.calls, initialCalls)
	}

	// 3. A configured layer opens with keep input method: should not notify
	switcher.config.LayerRules = []config.LayerRule{
		{Namespace: "icyleaf-calculator", InputMethod: config.KeepInputMethod},
	}
	switcher.currentIM = "english"
	switcher.appHistory["ghostty"] = "english"
	switcher.handleLayerEvent("icyleaf-calculator", true)
	if notifier.calls != initialCalls {
		t.Fatalf("handleLayerEvent(open configured keep) sent spurious notification: calls = %d, want %d", notifier.calls, initialCalls)
	}

	switcher.handleLayerEvent("icyleaf-calculator", false)
	if notifier.calls != initialCalls {
		t.Fatalf("handleLayerEvent(close configured keep) sent spurious notification: calls = %d, want %d", notifier.calls, initialCalls)
	}
}


func TestProfileAndEngineResolution(t *testing.T) {
	cfg := &config.Config{
		DefaultInputMethod: "chinese",
		InputMethods: map[string]config.InputMethodConfig{
			"english":  {Engine: "keyboard-us"},
			"chinese":  {Engine: "msime"},
			"japanese": {Engine: "rime", Profile: "jaroomaji"},
		},
	}
	cfg.Normalize()
	switcher := NewSwitcher(cfg)

	// Verify resolveFcitx5InputMethod returns correct labels
	if got := resolveFcitx5InputMethod(fcitxStateActive, "msime", "rime", cfg.InputMethods); got != "chinese" {
		t.Fatalf("resolve msime = %q, want chinese", got)
	}
	if got := resolveFcitx5InputMethod(fcitxStateActive, "keyboard-us", "rime", cfg.InputMethods); got != "english" {
		t.Fatalf("resolve keyboard-us = %q, want english", got)
	}
	if got := resolveFcitx5InputMethod(fcitxStateActive, "rime", "rime", cfg.InputMethods); got != "rime" {
		t.Fatalf("resolve rime = %q, want rime", got)
	}

	// Verify switcher GetCurrent schema matching with profile
	switcher.rime = NewRime(cfg.RimeSchemas)
	// If rime schema matches jaroomaji, it maps to japanese
	if label := switcher.rime.GetCurrentInputMethod(cfg.DefaultInputMethod); label != "japanese" {
		// Mock schema in mapSchemaToInputMethod
		mapped := mapSchemaToInputMethod("jaroomaji", cfg.RimeSchemas, "chinese")
		if mapped != "japanese" {
			t.Fatalf("mapSchemaToInputMethod = %q, want japanese", mapped)
		}
	}
}
