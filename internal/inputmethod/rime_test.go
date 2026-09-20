package inputmethod

import "testing"

func TestMapSchemaToInputMethod(t *testing.T) {
	schemas := map[string]string{
		"chinese":  "rime_frost",
		"japanese": "jaroomaji",
	}

	tests := []struct {
		name   string
		schema string
		want   string
	}{
		{
			name:   "chinese schema maps to chinese",
			schema: "rime_frost",
			want:   "chinese",
		},
		{
			name:   "japanese schema maps to japanese",
			schema: "jaroomaji",
			want:   "japanese",
		},
		{
			name:   "unknown schema falls back to default",
			schema: "someone_elses_schema",
			want:   "english",
		},
		{
			name:   "unavailable schema falls back to default",
			schema: "unknown",
			want:   "english",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapSchemaToInputMethod(tt.schema, schemas, "english")
			if got != tt.want {
				t.Fatalf("mapSchemaToInputMethod(%q) = %q, want %q", tt.schema, got, tt.want)
			}
		})
	}
}
