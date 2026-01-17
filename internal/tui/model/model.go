package model

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olamilekan000/etcd-tui/internal/config"
	"github.com/olamilekan000/etcd-tui/internal/etcd"
	"github.com/olamilekan000/etcd-tui/internal/tui/components/filter"
	"github.com/olamilekan000/etcd-tui/internal/tui/components/header"
	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
	"github.com/olamilekan000/etcd-tui/internal/tui/keymap"
	"github.com/olamilekan000/etcd-tui/internal/tui/style"
	"github.com/olamilekan000/etcd-tui/internal/tui/view"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

type Model struct {
	EtcdRepo  etcd.Repository
	Connected bool
	Error     error
	Endpoint  string

	AllKeys        []etcd.KeyValue
	FilteredKeys   []etcd.KeyValue
	LastFetchedKey string
	HasMoreKeys    bool
	FetchingKeys   bool

	Cursor        int
	TableYOffset  int
	SelectedKey   string
	SelectedValue string
	Focus         string

	CachedMaxVisibleRows int
	CachedHeight         int

	LastFilterValue string

	FetchingAllKeys bool
	FilterTriggered bool

	PreFilterAllKeys        []etcd.KeyValue
	PreFilterCursor         int
	PreFilterYOffset        int
	PreFilterLastFetchedKey string
	PreFilterHasMoreKeys    bool

	ShowValue      bool
	FormattedValue string
	ValueViewport  int
	IsJSON         bool

	Width         int
	Height        int
	SplitRatio    float64
	DraggingSplit bool

	Status          string
	LastRefresh     time.Time
	TotalKeys       int
	CopyMessage     string
	CopyMessageTime time.Time

	Header header.Model
	Filter filter.Model
}

func New() Model {
	endpoint := config.GetEndpoints()
	if endpoint == "" {
		endpoint = "not set"
	}

	status := "Connecting..."
	keyHelp := "q r / tab ↑↓ g/G enter esc"

	return Model{
		EtcdRepo:         etcd.NewRepository(),
		AllKeys:          []etcd.KeyValue{},
		FilteredKeys:     []etcd.KeyValue{},
		LastFetchedKey:   "",
		HasMoreKeys:      true,
		FetchingKeys:     false,
		Connected:        false,
		Cursor:           0,
		TableYOffset:     0,
		Endpoint:         endpoint,
		Status:           status,
		Focus:            constants.FocusTable,
		SplitRatio:       0.5,
		TotalKeys:        -1,
		FetchingAllKeys:  false,
		FilterTriggered:  false,
		PreFilterAllKeys: []etcd.KeyValue{},

		Header: header.New(constants.LogoString, "", endpoint, constants.Version, keyHelp),
		Filter: filter.New(status),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg { return m.EtcdRepo.Connect() }, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Filter.Focused() {
		updatedModel, cmd := m.handleFilterUpdate(msg)
		if cmd != nil || updatedModel.Filter.Focused() {
			return updatedModel, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		updatedModel, cmd := m.handleKey(msg)
		model := updatedModel.(Model)
		model, paginationCmd := model.checkPagination()
		return model, tea.Batch(cmd, paginationCmd)

	case tea.MouseMsg:
		model, cmd := m.handleMouse(msg)
		model, paginationCmd := model.checkPagination()
		return model, tea.Batch(cmd, paginationCmd)

	case tea.WindowSizeMsg:
		return m.handleResize(msg)

	case etcd.ConnectionMsg, etcd.KeysMsg, etcd.ValueMsg, etcd.CountMsg:
		return m.handleEtcdMsg(msg)

	case CopyMsg, ClearCopyMsg:
		return m.handleClipboardMsg(msg)
	}

	return m, nil
}

func (m Model) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	m.Header.SetWidth(m.Width)

	header := m.Header.View()
	filter := m.Filter.View()

	chromeHeight := lipgloss.Height(header) + lipgloss.Height(filter)
	if m.Error != nil {
		chromeHeight += 1
	}
	contentHeight := utils.Max(1, m.Height-chromeHeight)

	var content string
	tableData := m.getTableViewData(contentHeight)

	if m.ShowValue {
		valueData := m.getValueViewData(contentHeight)
		table := view.RenderTable(tableData)
		valView := view.RenderValueView(valueData)
		content = view.RenderSplitView(table, valView, m.Width, m.SplitRatio, m.DraggingSplit)
	} else {
		content = view.RenderTable(tableData)
	}

	var sections []string
	sections = append(sections, header)
	if m.Error != nil {
		sections = append(sections, m.renderError())
	}
	sections = append(sections, filter)
	sections = append(sections, content)

	result := strings.Join(sections, "")
	lines := strings.Split(result, "\n")
	if len(lines) > m.Height {
		lines = lines[:m.Height]
	}

	return strings.Join(lines, "\n")
}

func (m Model) handleFilterUpdate(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if key == constants.KeyEsc {
			m.Filter.BlurAndClear()
			cmd := (&m).applyFilter(false)
			m.Focus = constants.FocusTable
			m.updateKeyHelp()
			return m, cmd
		}
		if key == constants.KeyEnter {
			m.Filter.Blur()
			m.Focus = constants.FocusTable
			m.updateKeyHelp()
			cmd := (&m).applyFilter(true)
			return m, cmd
		}
		if key == constants.KeyTab {
			paginationCmd := m.blurFilterAndFocusTable()
			result, cmd := m.handleKey(msg)
			return result.(Model), tea.Batch(cmd, paginationCmd)
		}
		if key == constants.KeyUp || key == constants.KeyDown || key == constants.KeyLeft || key == constants.KeyRight {
			m.Filter.Blur()
			m.Focus = constants.FocusTable
			m.updateKeyHelp()
			result, cmd := m.handleKey(msg)
			return result.(Model), cmd
		}

		var filterCmd tea.Cmd
		m.Filter, filterCmd = m.Filter.Update(msg)
		paginationCmd := (&m).applyFilter(false)
		m.updateKeyHelp()
		return m, tea.Batch(filterCmd, paginationCmd)

	default:
		var filterCmd tea.Cmd
		m.Filter, filterCmd = m.Filter.Update(msg)
		return m, filterCmd
	}
}

func (m Model) handleMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	maxCursor := 0
	if len(m.FilteredKeys) > 0 {
		maxCursor = len(m.FilteredKeys) - 1
	}

	tableWidth := m.calculateTableWidth()

	result := keymap.HandleMouse(msg, keymap.MouseParams{
		ShowValue:     m.ShowValue,
		Focus:         m.Focus,
		Width:         m.Width,
		TableWidth:    tableWidth,
		SplitRatio:    m.SplitRatio,
		DraggingSplit: m.DraggingSplit,
		Cursor:        m.Cursor,
		MaxCursor:     maxCursor,
	})

	if len(m.FilteredKeys) == 0 {
		m.Cursor = 0
		m.TableYOffset = 0
	} else if result.Cursor < 0 {
		m.Cursor = 0
		m.TableYOffset = 0
	} else if result.Cursor >= len(m.FilteredKeys) {
		m.Cursor = len(m.FilteredKeys) - 1
		m.fixTableViewport()
	} else {
		m.Cursor = result.Cursor
		m.fixTableViewport()
	}

	m.SplitRatio = result.SplitRatio
	m.DraggingSplit = result.DraggingSplit

	if result.ScrollValue != 0 {
		m.scrollValue(result.ScrollValue)
	}

	if result.SetFocus != "" {
		m.Focus = result.SetFocus
		m.updateKeyHelp()
	}

	return m, nil
}

func (m Model) handleResize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.Width = msg.Width
	m.Height = msg.Height
	m.Header.SetWidth(m.Width)
	m.CachedMaxVisibleRows = 0
	m.CachedHeight = 0
	m.fixTableViewport()
	m.updateKeyHelp()
	return m, nil
}

func (m Model) handleEtcdMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case etcd.ConnectionMsg:
		result, cmd := m.handleConnectionMsg(msg)
		return result.(Model), cmd

	case etcd.KeysMsg:
		result, cmd := m.handleKeysMsg(msg)
		return result.(Model), cmd

	case etcd.ValueMsg:
		m = m.handleValueMsg(msg)
		return m, nil

	case etcd.CountMsg:
		if msg.Err == nil {
			m.TotalKeys = msg.Count
			m.updateStatus()
			m.Filter.SetPrefix(m.Status)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleClipboardMsg(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CopyMsg:
		if msg.Success {
			m.CopyMessage = msg.Message
		} else {
			m.CopyMessage = "Error: " + msg.Message
		}
		m.CopyMessageTime = time.Now()
		m.updateStatus()
		m.updateKeyHelp()
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return ClearCopyMsg{}
		})

	case ClearCopyMsg:
		m.CopyMessage = ""
		m.updateStatus()
		m.updateKeyHelp()
		return m, nil
	}
	return m, nil
}

func (m Model) checkPagination() (Model, tea.Cmd) {
	if m.FilterTriggered || m.FetchingAllKeys {
		return m, nil
	}

	if !m.FetchingKeys && m.HasMoreKeys && len(m.AllKeys) > 0 {
		threshold := len(m.AllKeys) - 10
		if threshold < 0 {
			threshold = 0
		}

		if m.Filter.HasFilterText() {
			if len(m.AllKeys) <= threshold+10 {
				m.FetchingKeys = true
				return m, m.EtcdRepo.FetchKeys(m.LastFetchedKey, 100)
			}
		} else {
			if m.Cursor >= threshold && m.Cursor < len(m.FilteredKeys) && len(m.FilteredKeys) == len(m.AllKeys) {
				m.FetchingKeys = true
				return m, m.EtcdRepo.FetchKeys(m.LastFetchedKey, 100)
			}
		}
	}
	return m, nil
}

func (m Model) renderError() string {
	if m.Error != nil {
		return style.Error.Render(fmt.Sprintf("⚠ Error: %v", m.Error))
	}
	return ""
}

func (m Model) getTableViewData(contentHeight int) view.TableViewData {
	return view.TableViewData{
		FilteredKeys: m.FilteredKeys,
		Cursor:       m.Cursor,
		YOffset:      m.TableYOffset,
		Focus:        view.FocusArea(m.Focus),
		ShowValue:    m.ShowValue,
		Width:        m.Width,
		Height:       contentHeight,
		SplitRatio:   m.SplitRatio,
		Filter:       m.Filter,
	}
}

func (m Model) getValueViewData(contentHeight int) view.ValueViewData {
	return view.ValueViewData{
		SelectedKey:    m.SelectedKey,
		SelectedValue:  m.SelectedValue,
		FormattedValue: m.FormattedValue,
		IsJSON:         m.IsJSON,
		ValueViewport:  m.ValueViewport,
		Focus:          view.FocusArea(m.Focus),
		Width:          m.Width,
		Height:         contentHeight,
		SplitRatio:     m.SplitRatio,
		DraggingSplit:  m.DraggingSplit,
	}
}

func (m Model) calculateTableWidth() int {
	if !m.ShowValue {
		return m.Width
	}
	w := int(float64(m.Width-3) * m.SplitRatio)
	return utils.Max(constants.MinPaneWidth, w)
}
