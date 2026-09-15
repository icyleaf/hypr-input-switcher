package inputmethod

import "testing"

func TestResolveRimeInputMethod(t *testing.T) {
	schemas := map[string]string{
		"chinese":  "rime_frost",
		"japanese": "jaroomaji",
	}

	tests := []struct {
		name      string
		asciiMode bool
		schema    string
		want      string
	}{
		{
			name:      "ascii mode reports english regardless of schema",
			asciiMode: true,
			schema:    "rime_frost",
			want:      "english",
		},
		{
			name:      "ascii mode wins over an unmapped schema",
			asciiMode: true,
			schema:    "someone_elses_schema",
			want:      "english",
		},
		{
			name:      "chinese schema maps to chinese",
			asciiMode: false,
			schema:    "rime_frost",
			want:      "chinese",
		},
		{
			name:      "japanese schema maps to japanese",
			asciiMode: false,
			schema:    "jaroomaji",
			want:      "japanese",
		},
		{
			name:      "unknown schema falls back to default",
			asciiMode: false,
			schema:    "someone_elses_schema",
			want:      "english",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRimeInputMethod(tt.asciiMode, tt.schema, schemas, "english")
			if got != tt.want {
				t.Fatalf("resolveRimeInputMethod(%v, %q) = %q, want %q", tt.asciiMode, tt.schema, got, tt.want)
			}
		})
	}
}
