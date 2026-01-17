package view

import (
	"testing"

	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

func TestCalculatePaneWidth(t *testing.T) {
	tests := []struct {
		name       string
		totalWidth int
		ratio      float64
		showValue  bool
		expected   int
	}{
		{"no value view", 100, 0.5, false, 100},
		{"with value view", 100, 0.5, true, 48},
		{"minimum width", 50, 0.1, true, constants.MinPaneWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePaneWidth(tt.totalWidth, tt.ratio, tt.showValue)
			if result != tt.expected {
				t.Errorf("calculatePaneWidth(%d, %f, %v) = %d, want %d",
					tt.totalWidth, tt.ratio, tt.showValue, result, tt.expected)
			}
		})
	}
}

func TestCalculateColumnWidths(t *testing.T) {
	tests := []struct {
		name      string
		available int
		checkFunc func(k, v int) bool
	}{
		{"normal case", 100, func(k, v int) bool {
			return k >= 60 && k <= 80 && v >= 20 && (k+v+4) <= 104
		}},
		{"small width", 30, func(k, v int) bool {
			return k >= 6 && v >= 20 && (k+v+4) <= 34
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v := calculateColumnWidths(tt.available)
			if !tt.checkFunc(k, v) {
				t.Errorf("calculateColumnWidths(%d) = (%d, %d), validation failed",
					tt.available, k, v)
			}
		})
	}
}

func TestCalculateNormalColumnWidths(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		checkFunc func(k, v int) bool
	}{
		{"normal case", 100, func(k, v int) bool {
			available := 100 - 16
			expectedK := int(float64(available) * 0.75)
			return k >= expectedK-5 && k <= expectedK+5 && (k+v) == available
		}},
		{"small width", 40, func(k, v int) bool {
			available := 40 - 16
			return v >= 20 && (k+v) == available
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v := calculateNormalColumnWidths(tt.width)
			if !tt.checkFunc(k, v) {
				t.Errorf("calculateNormalColumnWidths(%d) = (%d, %d), validation failed",
					tt.width, k, v)
			}
		})
	}
}

func TestCalculateFilteredColumnWidths(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		checkFunc func(k, v int) bool
	}{
		{"normal case", 100, func(k, v int) bool {
			expectedK := utils.Max(80, int(float64(100)*0.8))
			return k >= expectedK-5 && k <= expectedK+5 && (k+v+4) <= 104
		}},
		{"small width", 50, func(k, v int) bool {
			return k >= 26 && v >= 20 && (k+v+4) <= 54
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v := calculateFilteredColumnWidths(tt.width)
			if !tt.checkFunc(k, v) {
				t.Errorf("calculateFilteredColumnWidths(%d) = (%d, %d), validation failed",
					tt.width, k, v)
			}
		})
	}
}

func TestPadToWidth(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		width     int
		checkFunc func(result string, width int) bool
	}{
		{"exact width", "hello", 5, func(r string, w int) bool {
			return len(r) == w
		}},
		{"needs padding", "hi", 5, func(r string, w int) bool {
			return len(r) == w && r[:2] == "hi"
		}},
		{"needs truncation", "hello world", 5, func(r string, w int) bool {
			return len(r) <= w
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padToWidth(tt.s, tt.width)
			if !tt.checkFunc(result, tt.width) {
				t.Errorf("padToWidth(%q, %d) = %q, validation failed",
					tt.s, tt.width, result)
			}
		})
	}
}
