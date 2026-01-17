package table

import (
	"fmt"

	"github.com/anurag-roy/bubbletable/components"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/olamilekan000/etcd-tui/internal/etcd"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

type TableRow struct {
	Index int    `table:"#,width:6,sortable:false"`
	Key   string `table:"Key,width:70,sortable:false"`
	Value string `table:"Value,width:24,sortable:false"`
}

type Model struct {
	tableModel tea.Model

	keys      []etcd.KeyValue
	cursor    int
	width     int
	height    int
	showValue bool
	yOffset   int
}

func New() *Model {
	return &Model{
		keys:      []etcd.KeyValue{},
		cursor:    0,
		width:     80,
		height:    24,
		showValue: false,
		yOffset:   0,
	}
}

func (m *Model) SetKeys(keys []etcd.KeyValue) {
	keysChanged := len(m.keys) != len(keys)
	if !keysChanged && len(keys) > 0 {
		if len(m.keys) > 0 {
			keysChanged = keys[0].Key != m.keys[0].Key ||
				keys[len(keys)-1].Key != m.keys[len(m.keys)-1].Key
		}
	}

	m.keys = keys
	if m.cursor >= len(keys) {
		m.cursor = utils.Max(0, len(keys)-1)
	}

	if keysChanged {
		m.rebuildTable()
	}
}

func (m *Model) SetCursor(cursor int) {
	if cursor >= 0 && cursor < len(m.keys) {
		m.cursor = cursor
		if m.tableModel != nil {
		}
	}
}

func (m *Model) Cursor() int {
	return m.cursor
}

func (m *Model) SetSize(width, height int) {
	if m.width != width || m.height != height {
		m.width = width
		m.height = height
		if len(m.keys) > 0 {
			m.rebuildTable()
		}
	}
}

func (m *Model) SetShowValue(showValue bool) {
	if m.showValue != showValue {
		m.showValue = showValue
		if len(m.keys) > 0 {
			m.rebuildTable()
		}
	}
}

func (m *Model) SetYOffset(offset int) {
	if offset >= 0 {
		m.yOffset = offset
	}
}

func (m *Model) YOffset() int {
	return m.yOffset
}

func (m *Model) rebuildTable() {
	if len(m.keys) == 0 {
		m.tableModel = nil
		return
	}

	rows := make([]TableRow, 0, len(m.keys))
	for i, kv := range m.keys {
		rows = append(rows, TableRow{
			Index: i + 1,
			Key:   kv.Key,
			Value: kv.ValuePreview,
		})
	}

	pageSize := utils.Max(1, m.height-2)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > len(rows) {
		pageSize = len(rows)
	}

	tableModel := components.NewTable(rows).
		WithPageSize(pageSize).
		WithSorting(false).
		WithSearch(false)

	m.tableModel = tableModel

	if m.width > 0 && m.height > 0 {
		windowSizeMsg := tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		}
		updatedModel, _ := m.tableModel.Update(windowSizeMsg)
		if updatedModel != nil {
			m.tableModel = updatedModel
		}
	}
}

func (m *Model) Init() tea.Cmd {
	m.rebuildTable()
	if m.tableModel != nil {
		return m.tableModel.Init()
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if windowSizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowSizeMsg.Width
		m.height = windowSizeMsg.Height
		if len(m.keys) > 0 {
			m.rebuildTable()
		}
	}

	if m.tableModel == nil {
		return m, nil
	}

	updatedModel, cmd := m.tableModel.Update(msg)
	if updatedModel != nil {
		m.tableModel = updatedModel
	}

	return m, cmd
}

func (m *Model) View() string {
	if len(m.keys) == 0 {
		return "No keys found. Press 'r' to refresh."
	}

	if m.tableModel == nil {
		m.rebuildTable()
	}

	if m.tableModel == nil {
		return fmt.Sprintf("ERROR: Table model is nil, keys: %d, width: %d, height: %d",
			len(m.keys), m.width, m.height)
	}

	if m.width > 0 && m.height > 0 {
		windowSizeMsg := tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		}
		updatedModel, _ := m.tableModel.Update(windowSizeMsg)
		if updatedModel != nil {
			m.tableModel = updatedModel
		}
	}

	viewStr := m.tableModel.View()

	if viewStr == "" {
		return fmt.Sprintf("DEBUG: Table view is empty, keys: %d, width: %d, height: %d, pageSize: %d",
			len(m.keys), m.width, m.height, utils.Max(1, m.height-2))
	}

	return viewStr
}
