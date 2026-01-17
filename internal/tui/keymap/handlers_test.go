package keymap

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
)

func TestIsNavigationKey(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		expected bool
	}{
		{"arrow up", tea.KeyMsg{Type: tea.KeyUp}, true},
		{"arrow down", tea.KeyMsg{Type: tea.KeyDown}, true},
		{"tab", tea.KeyMsg{Type: tea.KeyTab}, true},
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}, true},
		{"escape", tea.KeyMsg{Type: tea.KeyEsc}, true},
		{"vim k", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, true},
		{"vim j", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, true},
		{"non-navigation key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, false},
		{"non-key message", tea.WindowSizeMsg{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNavigationKey(tt.msg)
			if result != tt.expected {
				t.Errorf("IsNavigationKey(%v) = %v, want %v", tt.msg, result, tt.expected)
			}
		})
	}
}

func TestHandleMouse_WheelUp(t *testing.T) {
	tests := []struct {
		name           string
		showValue      bool
		focus          string
		cursor         int
		expectedCursor int
		expectedScroll int
	}{
		{"table focus", false, constants.FocusTable, 5, 4, 0},
		{"value focus scrolls", true, constants.FocusValue, 5, 5, -1},
		{"cursor at 0", false, constants.FocusTable, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.MouseMsg{
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonWheelUp,
			}
			params := MouseParams{
				ShowValue: tt.showValue,
				Focus:     tt.focus,
				Cursor:    tt.cursor,
				MaxCursor: 10,
			}

			result := HandleMouse(msg, params)
			if result.Cursor != tt.expectedCursor {
				t.Errorf("HandleMouse() Cursor = %d, want %d", result.Cursor, tt.expectedCursor)
			}
			if result.ScrollValue != tt.expectedScroll {
				t.Errorf("HandleMouse() ScrollValue = %d, want %d", result.ScrollValue, tt.expectedScroll)
			}
		})
	}
}

func TestHandleMouse_WheelDown(t *testing.T) {
	tests := []struct {
		name           string
		showValue      bool
		focus          string
		cursor         int
		maxCursor      int
		expectedCursor int
		expectedScroll int
	}{
		{"table focus", false, constants.FocusTable, 5, 10, 6, 0},
		{"value focus scrolls", true, constants.FocusValue, 5, 10, 5, 1},
		{"cursor at max", false, constants.FocusTable, 10, 10, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.MouseMsg{
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonWheelDown,
			}
			params := MouseParams{
				ShowValue: tt.showValue,
				Focus:     tt.focus,
				Cursor:    tt.cursor,
				MaxCursor: tt.maxCursor,
			}

			result := HandleMouse(msg, params)
			if result.Cursor != tt.expectedCursor {
				t.Errorf("HandleMouse() Cursor = %d, want %d", result.Cursor, tt.expectedCursor)
			}
			if result.ScrollValue != tt.expectedScroll {
				t.Errorf("HandleMouse() ScrollValue = %d, want %d", result.ScrollValue, tt.expectedScroll)
			}
		})
	}
}

func TestHandleMouse_LeftClick(t *testing.T) {
	tests := []struct {
		name             string
		showValue        bool
		tableWidth       int
		width            int
		x                int
		expectedFocus    string
		expectedDragging bool
	}{
		{"click on table side", true, 50, 100, 25, constants.FocusTable, false},
		{"click on value side", true, 50, 100, 75, constants.FocusValue, false},
		{"click on separator", true, 50, 100, 50, "", true},
		{"no value view", false, 0, 100, 50, constants.FocusTable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.MouseMsg{
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonLeft,
				X:      tt.x,
			}
			params := MouseParams{
				ShowValue:  tt.showValue,
				TableWidth: tt.tableWidth,
				Width:      tt.width,
			}

			result := HandleMouse(msg, params)
			if result.SetFocus != tt.expectedFocus {
				t.Errorf("HandleMouse() SetFocus = %q, want %q", result.SetFocus, tt.expectedFocus)
			}
			if result.DraggingSplit != tt.expectedDragging {
				t.Errorf("HandleMouse() DraggingSplit = %v, want %v", result.DraggingSplit, tt.expectedDragging)
			}
		})
	}
}

func TestHandleMouse_Release(t *testing.T) {
	msg := tea.MouseMsg{
		Action: tea.MouseActionRelease,
	}
	params := MouseParams{
		DraggingSplit: true,
	}

	result := HandleMouse(msg, params)
	if result.DraggingSplit {
		t.Errorf("HandleMouse() DraggingSplit = true, want false on release")
	}
}

func TestHandleMouse_Motion(t *testing.T) {
	tests := []struct {
		name              string
		draggingSplit     bool
		showValue         bool
		x                 int
		width             int
		expectedDragging  bool
		expectRatioChange bool
	}{
		{"dragging with value view", true, true, 50, 100, true, true},
		{"not dragging", false, true, 50, 100, false, false},
		{"no value view", true, false, 50, 100, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.MouseMsg{
				Action: tea.MouseActionMotion,
				X:      tt.x,
			}
			params := MouseParams{
				DraggingSplit: tt.draggingSplit,
				ShowValue:     tt.showValue,
				Width:         tt.width,
				SplitRatio:    0.5,
			}

			result := HandleMouse(msg, params)
			if result.DraggingSplit != tt.expectedDragging {
				t.Errorf("HandleMouse() DraggingSplit = %v, want %v", result.DraggingSplit, tt.expectedDragging)
			}
			if tt.expectRatioChange && result.SplitRatio == 0.5 {
				t.Errorf("HandleMouse() SplitRatio should change when dragging, got %f", result.SplitRatio)
			}
			if !tt.expectRatioChange && result.SplitRatio != 0.5 {
				t.Errorf("HandleMouse() SplitRatio should not change, got %f", result.SplitRatio)
			}
		})
	}
}
