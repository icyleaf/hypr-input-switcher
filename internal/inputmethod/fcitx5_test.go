package inputmethod

import "testing"

func TestResolveFcitx5InputMethod(t *testing.T) {
	inputMethods := map[string]string{
		"english":  "keyboard-us",
		"chinese":  "msime",
		"japanese": "rime",
	}

	tests := []struct {
		name      string
		state     int
		currentIM string
		want      string
	}{
		{
			name:      "active rime is rime",
			state:     fcitxStateActive,
			currentIM: "rime",
			want:      "rime",
		},
		{
			name:      "active keyboard layout is english",
			state:     fcitxStateActive,
			currentIM: "keyboard-us",
			want:      "english",
		},
		{
			name:      "active msime maps to chinese",
			state:     fcitxStateActive,
			currentIM: "msime",
			want:      "chinese",
		},
		{
			name:      "inactive controller is english even when it names rime",
			state:     fcitxStateInactive,
			currentIM: "rime",
			want:      "english",
		},
		{
			name:      "inactive controller is english even when it names msime",
			state:     fcitxStateInactive,
			currentIM: "msime",
			want:      "english",
		},
		{
			name:      "closed controller is english",
			state:     fcitxStateClosed,
			currentIM: "rime",
			want:      "english",
		},
		{
			name:      "inactive keyboard layout is english",
			state:     fcitxStateInactive,
			currentIM: "keyboard-us",
			want:      "english",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFcitx5InputMethod(tt.state, tt.currentIM, "rime", inputMethods)
			if got != tt.want {
				t.Fatalf("resolveFcitx5InputMethod(%d, %q) = %q, want %q", tt.state, tt.currentIM, got, tt.want)
			}
		})
	}
}
