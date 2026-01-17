package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	filterInput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/olamilekan000/etcd-tui/internal/config"
	"github.com/olamilekan000/etcd-tui/internal/etcd"
	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
	"github.com/olamilekan000/etcd-tui/internal/tui/keymap"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

func (m *Model) filterKeys() []etcd.KeyValue {
	filterValue := m.Filter.Value()
	if filterValue == "" {
		if len(m.AllKeys) == 0 {
			return []etcd.KeyValue{}
		}
		return m.AllKeys[:len(m.AllKeys):len(m.AllKeys)]
	}

	estimatedCapacity := len(m.AllKeys) / 2
	if estimatedCapacity > 1000 {
		estimatedCapacity = 1000
	}
	filtered := make([]etcd.KeyValue, 0, estimatedCapacity)

	filterLower := strings.ToLower(filterValue)
	filterLen := len(filterLower)

	for i := range m.AllKeys {
		kv := &m.AllKeys[i]

		if len(kv.Key) < filterLen && len(kv.Value) < filterLen {
			continue
		}

		if len(kv.Key) >= filterLen && strings.Contains(strings.ToLower(kv.Key), filterLower) {
			filtered = append(filtered, *kv)
			continue
		}

		if len(kv.Value) >= filterLen && strings.Contains(strings.ToLower(kv.Value), filterLower) {
			filtered = append(filtered, *kv)
		}
	}
	return filtered
}

func (m *Model) applyFilter() {
	oldCursor := m.Cursor
	if oldCursor < 0 {
		oldCursor = 0
	}

	currentFilterValue := m.Filter.Value()

	if currentFilterValue == m.LastFilterValue && len(m.FilteredKeys) > 0 && m.LastFilterValue != "" {
		return
	}

	m.FilteredKeys = m.filterKeys()

	m.LastFilterValue = currentFilterValue

	if len(m.FilteredKeys) == 0 {
		m.Cursor = 0
		m.TableYOffset = 0
		return
	}

	if oldCursor >= len(m.FilteredKeys) {
		m.Cursor = len(m.FilteredKeys) - 1
	} else if oldCursor >= 0 {
		m.Cursor = oldCursor
	} else {
		m.Cursor = 0
	}

	m.fixTableViewport()
}

func (m *Model) getMaxVisibleRows() int {
	if m.Height <= 0 {
		return 10
	}

	if m.CachedMaxVisibleRows > 0 && m.CachedHeight == m.Height {
		return m.CachedMaxVisibleRows
	}

	headerLines := 10
	filterLines := 2
	tableHeaderLines := 2

	if m.Error != nil {
		headerLines++
	}

	maxRows := m.Height - headerLines - filterLines - tableHeaderLines
	if maxRows < 1 {
		maxRows = 1
	}

	m.CachedMaxVisibleRows = maxRows
	m.CachedHeight = m.Height
	return maxRows
}

func (m *Model) fixTableViewport() {
	if len(m.FilteredKeys) == 0 {
		m.TableYOffset = 0
		m.Cursor = 0
		return
	}

	maxRows := m.getMaxVisibleRows()
	if maxRows <= 0 {
		maxRows = 1
	}
	if maxRows > len(m.FilteredKeys) {
		maxRows = len(m.FilteredKeys)
	}

	m.Cursor = utils.Clamp(m.Cursor, 0, len(m.FilteredKeys)-1)

	if m.Cursor < m.TableYOffset {
		m.TableYOffset = m.Cursor
	}

	if m.Cursor >= m.TableYOffset+maxRows {
		m.TableYOffset = m.Cursor - maxRows + 1
	}

	maxYOffset := len(m.FilteredKeys) - maxRows
	if maxYOffset < 0 {
		maxYOffset = 0
	}
	m.TableYOffset = utils.Clamp(m.TableYOffset, 0, maxYOffset)
}

func (m *Model) scrollValue(delta int) {
	displayValue := m.FormattedValue
	if displayValue == "" {
		displayValue = m.SelectedValue
	}

	lines := strings.Split(displayValue, "\n")
	if len(lines) == 0 {
		m.ValueViewport = 0
		return
	}

	headerView := m.Header.View()
	headerLines := len(strings.Split(headerView, "\n"))
	filterView := m.Filter.View()
	filterLines := len(strings.Split(filterView, "\n"))
	usedHeight := headerLines + filterLines
	availableHeight := utils.Max(1, m.Height-usedHeight)

	valueHeaderLines := 2
	availableContentHeight := utils.Max(1, availableHeight-valueHeaderLines)

	needsFooter := len(lines) > availableContentHeight
	maxVisibleLines := availableContentHeight
	if needsFooter {
		maxVisibleLines = availableContentHeight - 1
	}

	maxViewport := utils.Max(0, len(lines)-maxVisibleLines)

	newViewport := m.ValueViewport + delta
	m.ValueViewport = utils.Clamp(newViewport, 0, maxViewport)
}

func (m *Model) updateKeyHelp() {
	m.Header.SetKeyHelp(keymap.GenerateKeyHelp(m.ShowValue))
}

func (m *Model) updateStatus() {
	if m.CopyMessage != "" && time.Since(m.CopyMessageTime) < 2*time.Second {
		m.Status = m.CopyMessage
		m.Filter.SetPrefix(m.Status)
		return
	}

	if m.CopyMessage != "" && time.Since(m.CopyMessageTime) >= 2*time.Second {
		m.CopyMessage = ""
	}

	if m.TotalKeys >= 0 {
		m.Status = fmt.Sprintf("Keys: %d/%d", len(m.AllKeys), m.TotalKeys)
	} else {
		m.Status = fmt.Sprintf("Keys: %d", len(m.AllKeys))
	}
	if m.Filter.HasFilterText() {
		m.Status += fmt.Sprintf(" filtered: %d", len(m.FilteredKeys))
	}

	m.Filter.SetPrefix(m.Status)
}

func (m *Model) adjustSplit(delta float64) {
	newRatio := m.SplitRatio + delta
	m.SplitRatio = utils.ClampFloat(newRatio, constants.MinSplitRatio, constants.MaxSplitRatio)
}

func (m *Model) clearValueView() {
	m.SelectedKey = ""
	m.SelectedValue = ""
	m.FormattedValue = ""
	m.IsJSON = false
	m.ShowValue = false
	m.ValueViewport = 0
	m.Focus = constants.FocusTable
}

func (m *Model) jumpToBottomOfValue() {
	displayValue := m.FormattedValue
	if displayValue == "" {
		displayValue = m.SelectedValue
	}
	lines := strings.Split(displayValue, "\n")
	if len(lines) == 0 {
		m.ValueViewport = 0
		return
	}

	headerView := m.Header.View()
	headerLines := len(strings.Split(headerView, "\n"))
	filterView := m.Filter.View()
	filterLines := len(strings.Split(filterView, "\n"))
	usedHeight := headerLines + filterLines
	availableHeight := utils.Max(1, m.Height-usedHeight)

	valueHeaderLines := 2
	availableContentHeight := utils.Max(1, availableHeight-valueHeaderLines)

	needsFooter := len(lines) > availableContentHeight
	maxVisibleLines := availableContentHeight
	if needsFooter {
		maxVisibleLines = availableContentHeight - 1
	}

	m.ValueViewport = utils.Max(0, len(lines)-maxVisibleLines)
}

func (m *Model) blurFilterAndFocusTable() {
	m.Filter.Blur()
	m.applyFilter()
	m.Focus = constants.FocusTable
	m.updateKeyHelp()
}

func (m *Model) ensureTableFocus() {
	if !m.ShowValue {
		if m.Filter.Focused() {
			m.Filter.Blur()
		}
		m.Focus = constants.FocusTable
	}
}

func (m Model) handleQuit() (tea.Model, tea.Cmd) {
	if m.EtcdRepo != nil {
		m.EtcdRepo.Close()
	}
	return m, tea.Quit
}

func (m Model) handleFilterFocus() (tea.Model, tea.Cmd) {
	m.Filter.Focus()
	m.updateKeyHelp()
	return m, filterInput.Blink
}

func (m Model) handleEscape() (tea.Model, tea.Cmd) {
	if m.Filter.Focused() {
		m.Filter.BlurAndClear()
		m.applyFilter()
		m.updateKeyHelp()
		return m, nil
	}
	if m.ShowValue {
		m.clearValueView()
		m.updateKeyHelp()
	}
	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.Filter.Focused() {
		m.Filter.Blur()
		m.updateKeyHelp()
		return m, nil
	}
	if m.Focus == constants.FocusTable && m.Connected && len(m.FilteredKeys) > 0 && m.Cursor >= 0 && m.Cursor < len(m.FilteredKeys) {
		m.SelectedKey = m.FilteredKeys[m.Cursor].Key
		m.ShowValue = true
		m.ValueViewport = 0
		m.Focus = constants.FocusValue
		m.updateKeyHelp()
		return m, m.EtcdRepo.FetchValue(m.SelectedKey)
	}
	return m, nil
}

func (m Model) handleTab() (tea.Model, tea.Cmd) {
	if m.ShowValue {
		if m.Focus == constants.FocusTable {
			m.Focus = constants.FocusValue
		} else {
			m.Focus = constants.FocusTable
		}
		m.updateKeyHelp()
	}
	return m, nil
}

func (m Model) handleUp() (tea.Model, tea.Cmd) {
	if m.Focus == constants.FocusValue && m.ShowValue {
		m.scrollValue(-1)
		return m, nil
	}
	m.ensureTableFocus()
	if m.Focus != constants.FocusTable {
		return m, nil
	}

	if len(m.FilteredKeys) == 0 {
		m.Cursor = 0
		m.TableYOffset = 0
		return m, nil
	}

	m.Cursor = utils.Max(0, m.Cursor-1)
	m.fixTableViewport()

	return m, nil
}

func (m Model) handleDown() (tea.Model, tea.Cmd) {
	if m.Focus == constants.FocusValue && m.ShowValue {
		m.scrollValue(1)
		return m, nil
	}
	m.ensureTableFocus()
	if m.Focus != constants.FocusTable {
		return m, nil
	}

	if len(m.FilteredKeys) == 0 {
		m.Cursor = 0
		m.TableYOffset = 0
		return m, nil
	}

	m.Cursor = utils.Min(m.Cursor+1, len(m.FilteredKeys)-1)
	m.fixTableViewport()

	return m, nil
}

func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	if !m.Connected {
		return m, nil
	}
	m.AllKeys = []etcd.KeyValue{}
	m.LastFetchedKey = ""
	m.HasMoreKeys = true
	m.TotalKeys = -1
	return m, tea.Batch(
		m.EtcdRepo.FetchKeys("", 100),
		m.EtcdRepo.FetchTotalCount(),
	)
}

func (m Model) handleJumpToTop() (tea.Model, tea.Cmd) {
	if m.Focus == constants.FocusValue && m.ShowValue {
		m.ValueViewport = 0
		return m, nil
	}
	if !m.ShowValue {
		m.Focus = constants.FocusTable
	}
	if m.Focus == constants.FocusTable {
		m.Cursor = 0
		m.TableYOffset = 0
	}
	return m, nil
}

func (m Model) handleJumpToBottom() (tea.Model, tea.Cmd) {
	if m.Focus == constants.FocusValue && m.ShowValue {
		m.jumpToBottomOfValue()
		return m, nil
	}
	m.ensureTableFocus()
	if m.Focus != constants.FocusTable || len(m.FilteredKeys) == 0 {
		return m, nil
	}

	m.Cursor = len(m.FilteredKeys) - 1
	m.fixTableViewport()
	return m, nil
}

func (m Model) handleSplitAdjust(delta float64) (tea.Model, tea.Cmd) {
	if m.ShowValue {
		m.adjustSplit(delta)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m.handleQuit()
	case "/":
		return m.handleFilterFocus()
	case "esc":
		return m.handleEscape()
	case "enter":
		return m.handleEnter()
	case "tab":
		return m.handleTab()
	case "up", "k":
		return m.handleUp()
	case "down", "j":
		return m.handleDown()
	case "r", "R":
		return m.handleRefresh()
	case "g":
		return m.handleJumpToTop()
	case "G":
		return m.handleJumpToBottom()
	case "left", "h":
		return m.handleSplitAdjust(-constants.SplitAdjustInc)
	case "right", "l":
		return m.handleSplitAdjust(constants.SplitAdjustInc)
	case "c", "y":
		return m.handleCopy()
	}
	return m, nil
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		if err != nil {
			return CopyMsg{Success: false, Message: fmt.Sprintf("Copy failed: %v", err)}
		}
		preview := text
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		return CopyMsg{Success: true, Message: fmt.Sprintf("Copied to clipboard: %s", preview)}
	}
}

type CopyMsg struct {
	Success bool
	Message string
}

type ClearCopyMsg struct{}

func (m Model) handleCopy() (tea.Model, tea.Cmd) {
	if len(m.FilteredKeys) == 0 || m.Cursor < 0 || m.Cursor >= len(m.FilteredKeys) {
		return m, nil
	}

	kv := m.FilteredKeys[m.Cursor]
	valueToCopy := kv.Value

	if valueToCopy == "" {
		valueToCopy = kv.Key
	}

	return m, copyToClipboard(valueToCopy)
}

func (m Model) handleConnectionMsg(msg etcd.ConnectionMsg) (tea.Model, tea.Cmd) {
	if msg.Success {
		m.EtcdRepo.SetClient(msg.Client)
		m.Connected = true
		m.Status = "Connected"
		m.Endpoint = config.GetEndpoints()
		m.Header.SetEndpoint(m.Endpoint)
		m.Filter.SetPrefix(m.Status)
		m.LastRefresh = time.Now()
		m.AllKeys = []etcd.KeyValue{}
		m.LastFetchedKey = ""
		m.HasMoreKeys = true
		m.TotalKeys = -1
		m.updateKeyHelp()
		return m, tea.Batch(
			m.EtcdRepo.FetchKeys("", 100),
			m.EtcdRepo.FetchTotalCount(),
		)
	}
	m.Error = msg.Err
	m.Status = "Connection Failed"
	m.updateKeyHelp()
	return m, nil
}

func (m Model) handleKeysMsg(msg etcd.KeysMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.Error = msg.Err
		m.Status = "Error fetching keys"
		m.updateKeyHelp()
		return m, nil
	}

	for i := range msg.Keys {
		msg.Keys[i].Key = utils.SanitizeForTUI(msg.Keys[i].Key)
		msg.Keys[i].Value = utils.SanitizeForTUI(msg.Keys[i].Value)
		msg.Keys[i].ValuePreview = utils.SanitizeForTUI(msg.Keys[i].ValuePreview)
	}

	if len(msg.Keys) > 0 {
		m.LastFetchedKey = msg.Keys[len(msg.Keys)-1].Key
		m.HasMoreKeys = msg.HasMore

		existingKeys := make(map[string]bool, len(m.AllKeys))
		for _, kv := range m.AllKeys {
			existingKeys[kv.Key] = true
		}

		for _, kv := range msg.Keys {
			if !existingKeys[kv.Key] {
				m.AllKeys = append(m.AllKeys, kv)
			}
		}
	} else {
		m.HasMoreKeys = false
	}

	m.FetchingKeys = false

	oldLastFilterValue := m.LastFilterValue
	m.LastFilterValue = ""
	m.applyFilter()
	m.LastFilterValue = oldLastFilterValue

	if len(m.FilteredKeys) > 0 && m.Cursor >= len(m.FilteredKeys) {
		m.Cursor = len(m.FilteredKeys) - 1
	}

	m.fixTableViewport()
	m.updateStatus()
	m.Filter.SetPrefix(m.Status)
	m.LastRefresh = time.Now()
	m.updateKeyHelp()
	return m, nil
}

func (m Model) handleValueMsg(msg etcd.ValueMsg) Model {
	if msg.Err != nil {
		m.Error = msg.Err
		return m
	}
	trimmedValue := strings.TrimSpace(msg.Value)
	m.SelectedValue = trimmedValue
	formatted, isJSON := utils.FormatJSON(trimmedValue)
	m.FormattedValue = formatted
	m.IsJSON = isJSON
	m.ValueViewport = 0
	return m
}
