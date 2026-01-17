package utils

import (
	"testing"
)

func TestNormalizeForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{"empty string", "", 10, ""},
		{"short text", "hello", 10, "hello"},
		{"long text", "this is a very long text that should be truncated", 20, "this is a very long ..."},
		{"with newlines", "hello\nworld", 20, "hello world"},
		{"with tabs", "hello\tworld", 20, "hello world"},
		{"with carriage returns", "hello\rworld", 20, "helloworld"},
		{"multiple spaces", "hello    world", 20, "hello world"},
		{"leading/trailing spaces", "  hello  ", 20, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeForDisplay(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("NormalizeForDisplay(%q, %d) = %q, want %q", tt.text, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{"empty string", "", 10, ""},
		{"short string", "hello", 10, "hello"},
		{"long string", "hello world", 5, "he..."},
		{"maxLen less than 3", "hello", 1, "h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.s, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		min      int
		max      int
		expected int
	}{
		{"within range", 5, 0, 10, 5},
		{"below min", -5, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.val, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.val, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		name     string
		val      float64
		min      float64
		max      float64
		expected float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampFloat(tt.val, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("ClampFloat(%f, %f, %f) = %f, want %f", tt.val, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a greater", 10, 5, 10},
		{"b greater", 5, 10, 10},
		{"equal", 5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a greater", 10, 5, 5},
		{"b greater", 5, 10, 5},
		{"equal", 5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSanitizeForTUI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"normal text", "hello world", "hello world"},
		{"with newline", "hello\nworld", "hello world"},
		{"with carriage return", "hello\rworld", "hello world"},
		{"with tab", "hello\tworld", "hello world"},
		{"with control chars", "hello\x00world", "helloworld"},
		{"with ANSI codes", "hello\x1b[31mworld", "hello[31mworld"},
		{"mixed", "hello\n\t\x00world", "hello  world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForTUI(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForTUI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
