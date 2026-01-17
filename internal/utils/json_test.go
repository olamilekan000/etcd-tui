package utils

import (
	"strings"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantJSON  bool
		wantValid bool
	}{
		{"valid JSON object", `{"key":"value"}`, true, true},
		{"valid JSON array", `[1,2,3]`, true, true},
		{"nested JSON", `{"a":{"b":"c"}}`, true, true},
		{"invalid JSON", "not json", false, false},
		{"empty string", "", false, false},
		{"plain text", "hello world", false, false},
		{"JSON with whitespace", `  {"key":"value"}  `, true, true},
		{"malformed JSON", `{"key":}`, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, isJSON := FormatJSON(tt.input)

			if isJSON != tt.wantJSON {
				t.Errorf("FormatJSON(%q) isJSON = %v, want %v", tt.input, isJSON, tt.wantJSON)
			}

			if tt.wantJSON {
				if !strings.Contains(result, "\n") && len(result) > len(tt.input) {
					t.Errorf("FormatJSON(%q) should return formatted JSON, got %q", tt.input, result)
				}
			} else {
				trimmed := strings.TrimSpace(tt.input)
				if result != trimmed {
					t.Errorf("FormatJSON(%q) = %q, want %q", tt.input, result, trimmed)
				}
			}
		})
	}
}
