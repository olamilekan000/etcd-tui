package keymap

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

type MouseParams struct {
	ShowValue     bool
	Focus         string
	Width         int
	TableWidth    int
	SplitRatio    float64
	DraggingSplit bool
	Cursor        int
	MaxCursor     int
}

type MouseResult struct {
	Cursor        int
	SplitRatio    float64
	DraggingSplit bool
	ScrollValue   int
	SetFocus      string
}

func HandleMouse(msg tea.MouseMsg, params MouseParams) MouseResult {
	result := MouseResult{
		Cursor:        params.Cursor,
		SplitRatio:    params.SplitRatio,
		DraggingSplit: params.DraggingSplit,
		SetFocus:      "",
	}

	if msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if params.ShowValue && params.Focus == constants.FocusValue {
				result.ScrollValue = -1
			} else {
				result.Cursor = utils.Max(0, params.Cursor-1)
			}
		case tea.MouseButtonWheelDown:
			if params.ShowValue && params.Focus == constants.FocusValue {
				result.ScrollValue = 1
			} else {
				result.Cursor = utils.Min(params.MaxCursor, params.Cursor+1)
			}
		case tea.MouseButtonLeft:
			if params.ShowValue {
				separatorZone := 2
				if params.DraggingSplit || (msg.X >= params.TableWidth-separatorZone && msg.X <= params.TableWidth+separatorZone) {
					result.DraggingSplit = true
					mouseX := utils.Clamp(int(msg.X), constants.MinPaneWidth, params.Width-constants.MinPaneWidth-3)
					result.SplitRatio = utils.ClampFloat(float64(mouseX)/float64(params.Width-3), constants.MinSplitRatio, constants.MaxSplitRatio)
				} else if msg.X > params.TableWidth+separatorZone {
					result.SetFocus = constants.FocusValue
				} else if msg.X < params.TableWidth-separatorZone {
					result.SetFocus = constants.FocusTable
				}
			} else {
				result.SetFocus = constants.FocusTable
			}
		}
	}

	if msg.Action == tea.MouseActionRelease {
		result.DraggingSplit = false
	}

	if msg.Action == tea.MouseActionMotion {
		if params.DraggingSplit && params.ShowValue {
			mouseX := utils.Clamp(int(msg.X), constants.MinPaneWidth, params.Width-constants.MinPaneWidth-3)
			result.SplitRatio = utils.ClampFloat(float64(mouseX)/float64(params.Width-3), constants.MinSplitRatio, constants.MaxSplitRatio)
		}
	}

	return result
}

func IsNavigationKey(msg tea.Msg) bool {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false
	}

	switch keyMsg.Type {
	case tea.KeyUp, tea.KeyDown, tea.KeyTab, tea.KeyEnter, tea.KeyEsc,
		tea.KeyPgUp, tea.KeyPgDown, tea.KeyLeft, tea.KeyRight:
		return true
	}

	k := keyMsg.String()
	return k == "k" || k == "j" || k == "g" || k == "G" || k == "h" || k == "l"
}
