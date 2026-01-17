package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func NormalizeForDisplay(text string, maxLen int) string {
	normalized := strings.ReplaceAll(text, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")
	normalized = strings.ReplaceAll(normalized, "\r", "")

	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	normalized = strings.TrimSpace(normalized)

	if len(normalized) > maxLen {
		return normalized[:maxLen] + "..."
	}
	return normalized
}

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func TruncateANSI(s string, maxWidth int) string {
	actualWidth := lipgloss.Width(s)
	if actualWidth <= maxWidth {
		return s
	}

	if maxWidth < 3 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= maxWidth-3 {
		return s
	}

	truncated := string(runes[:maxWidth-3]) + "..."

	if lipgloss.Width(truncated) > maxWidth {
		for lipgloss.Width(truncated) > maxWidth && len(truncated) > 0 {
			truncated = truncated[:len(truncated)-1]
		}
	}

	return truncated
}

func Clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func ClampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SanitizeForTUI removes control characters and forces single-line output.
// This prevents binary data, ANSI escape sequences, and control codes from breaking the TUI.
// Only allows printable ASCII characters (32-126) and converts line breaks to spaces.
func SanitizeForTUI(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 32 && c <= 126 {
			b = append(b, c)
		} else if c == '\n' || c == '\r' || c == '\t' {
			b = append(b, ' ')
		}
	}
	return string(b)
}

func WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	wrapped := lipgloss.NewStyle().Width(width).Render(text)
	return strings.Split(wrapped, "\n")
}
