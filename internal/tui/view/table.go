package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/olamilekan000/etcd-tui/internal/etcd"
	"github.com/olamilekan000/etcd-tui/internal/tui/components/filter"
	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
	"github.com/olamilekan000/etcd-tui/internal/tui/style"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

type FocusArea string

const (
	FocusTable FocusArea = "table"
	FocusValue FocusArea = "value"
	ColumnGap            = "      "
)

type TableViewData struct {
	FilteredKeys []etcd.KeyValue
	Cursor       int
	YOffset      int
	Focus        FocusArea
	ShowValue    bool
	Width        int
	Height       int
	SplitRatio   float64
	Filter       filter.Model
}

type ValueViewData struct {
	SelectedKey    string
	SelectedValue  string
	FormattedValue string
	IsJSON         bool
	ValueViewport  int
	Focus          FocusArea
	Width          int
	Height         int
	SplitRatio     float64
	DraggingSplit  bool
}

func RenderTable(data TableViewData) string {
	width := calculatePaneWidth(data.Width, data.SplitRatio, data.ShowValue)

	if len(data.FilteredKeys) == 0 {
		return renderEmptyState(data, width)
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(renderTableHeader(data, width))

	maxRows := utils.Max(1, data.Height-3)
	startIdx, endIdx := calculateVisibleIndices(data, maxRows)

	for i := startIdx; i < endIdx; i++ {
		row := renderRow(i, data, width)
		b.WriteString(ensureValidRow(row, i))
	}

	if len(data.FilteredKeys) > endIdx {
		b.WriteString(fmt.Sprintf("\n... and %d more keys", len(data.FilteredKeys)-endIdx))
	}

	return finalizeOutput(b.String(), width)
}

func calculatePaneWidth(totalWidth int, ratio float64, showValue bool) int {
	if !showValue {
		return totalWidth
	}
	w := int(float64(totalWidth-3) * ratio)
	return utils.Max(constants.MinPaneWidth, w)
}

func renderEmptyState(data TableViewData, width int) string {
	msg := "No keys found. Press 'r' to refresh."
	if data.Filter.HasFilterText() {
		msg = "No keys match the filter."
	}
	msg = utils.Truncate(msg, width)
	return msg + "\n\n"
}

func renderTableHeader(data TableViewData, width int) string {
	if data.ShowValue {
		title := "Keys"
		headerRendered := style.TableHeader.Render(title)
		if lipgloss.Width(headerRendered) > width-2 {
			headerRendered = utils.Truncate(headerRendered, width-2)
		}
		return headerRendered + "\n" + strings.Repeat("─", width-2) + "\n"
	}

	availableWidth := width - 3
	var kWidth, vWidth int
	if data.Filter.HasFilterText() {
		kWidth, vWidth = calculateFilteredColumnWidths(availableWidth)
	} else {
		kWidth, vWidth = calculateNormalColumnWidths(availableWidth)
	}

	keyH := style.TableHeader.Render("Key")
	valH := style.TableHeader.Render("Value")

	if !data.Filter.HasFilterText() {
		rowNumberSpace := "     "
		keyH = rowNumberSpace + keyH
	}

	if lipgloss.Width(keyH) > kWidth {
		keyH = utils.Truncate(keyH, kWidth)
	}
	if lipgloss.Width(valH) > vWidth {
		valH = utils.Truncate(valH, vWidth)
	}

	// Pad to fixed widths for perfect alignment
	keyH = padToWidth(keyH, kWidth)
	valH = padToWidth(valH, vWidth)

	return keyH + ColumnGap + valH + "\n" + strings.Repeat("─", width-2) + "\n"
}

func calculateVisibleIndices(data TableViewData, maxRows int) (startIdx, endIdx int) {
	validCursor := data.Cursor
	if validCursor < 0 {
		validCursor = 0
	}
	if len(data.FilteredKeys) > 0 && validCursor >= len(data.FilteredKeys) {
		validCursor = len(data.FilteredKeys) - 1
	}

	startIdx = data.YOffset
	if startIdx < 0 {
		startIdx = 0
	}

	maxStartIdx := utils.Max(0, len(data.FilteredKeys)-maxRows)
	if startIdx > maxStartIdx {
		startIdx = maxStartIdx
	}

	if startIdx >= len(data.FilteredKeys) {
		startIdx = utils.Max(0, len(data.FilteredKeys)-1)
	}

	if startIdx < 0 || startIdx >= len(data.FilteredKeys) {
		startIdx = 0
	}

	endIdx = utils.Min(len(data.FilteredKeys), startIdx+maxRows)
	return startIdx, endIdx
}

func ensureValidRow(row string, idx int) string {
	if row == "" || strings.TrimSpace(row) == "" {
		return fmt.Sprintf("  %4d empty row\n", idx+1)
	}
	return row
}

func finalizeOutput(output string, width int) string {
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(output, "\n")
	var finalLines []string
	for _, line := range lines {
		if lipgloss.Width(line) > width {
			line = utils.Truncate(line, width)
		}
		finalLines = append(finalLines, line)
	}
	return strings.Join(finalLines, "\n")
}

func renderRow(idx int, data TableViewData, width int) string {
	if idx < 0 || idx >= len(data.FilteredKeys) {
		return ""
	}

	kv := data.FilteredKeys[idx]
	selected := idx == data.Cursor
	cursor := getCursorIndicator(selected)

	if data.ShowValue {
		return renderKeyOnlyRow(idx, kv, cursor, width, selected)
	}

	return renderFullRow(idx, kv, cursor, data, width, selected)
}

func renderKeyOnlyRow(idx int, kv etcd.KeyValue, cursor string, width int, selected bool) string {
	numberWidth := 8
	keyDisplay := truncateString(kv.Key, width-6-numberWidth)
	rowNumber := style.RowNumber.Render(fmt.Sprintf("%4d ", idx+1))

	line := cursor + rowNumber + keyDisplay
	if lipgloss.Width(line) > width {
		line = utils.Truncate(line, width)
	}

	rowStyle := getRowStyle(selected).
		MaxWidth(width).
		MaxHeight(1).
		Inline(true)

	return rowStyle.Render(line) + "\n"
}

func renderFullRow(idx int, kv etcd.KeyValue, cursor string, data TableViewData, width int, selected bool) string {
	availableWidth := width - 3
	var kWidth, vWidth int

	if data.Filter.HasFilterText() {
		kWidth, vWidth = calculateFilteredColumnWidths(availableWidth)
	} else {
		kWidth, vWidth = calculateNormalColumnWidths(availableWidth)
	}

	keyContent := kv.Key
	valContent := kv.ValuePreview

	if !data.Filter.HasFilterText() {
		rowNumber := style.RowNumber.Render(fmt.Sprintf("%4d ", idx+1))
		keyContent = rowNumber + keyContent
	}

	if selected {
		return renderSelectedRow(cursor, keyContent, valContent, kWidth, vWidth, width)
	}

	keyDisplay := truncateString(keyContent, kWidth-2)
	valDisplay := truncateString(valContent, vWidth-2)

	keyCell := style.KeyColumn.Render(keyDisplay)
	valCell := style.ValueColumn.Render(valDisplay)

	keyCell = padToWidth(keyCell, kWidth)
	valCell = padToWidth(valCell, vWidth)

	rowContent := cursor + keyCell + ColumnGap + valCell

	if lipgloss.Width(rowContent) > width {
		if data.Filter.HasFilterText() {
			maxKeyWidth := (width - 8) * 80 / 100
			maxValWidth := (width - 8) * 20 / 100
			keyDisplay = truncateString(kv.Key, maxKeyWidth)
			valDisplay = truncateString(kv.ValuePreview, maxValWidth)
		} else {
			maxKeyWidth := (width - 16) * 75 / 100
			maxValWidth := (width - 16) * 25 / 100
			rowNumber := style.RowNumber.Render(fmt.Sprintf("%4d ", idx+1))
			keyDisplay = truncateString(rowNumber+kv.Key, maxKeyWidth)
			valDisplay = truncateString(kv.ValuePreview, maxValWidth)
		}
		keyCell = style.KeyColumn.Render(keyDisplay)
		valCell = style.ValueColumn.Render(valDisplay)
		keyCell = padToWidth(keyCell, kWidth)
		valCell = padToWidth(valCell, vWidth)
		rowContent = cursor + keyCell + ColumnGap + valCell
		if lipgloss.Width(rowContent) > width {
			rowContent = utils.Truncate(rowContent, width)
		}
	}

	rowStyle := lipgloss.NewStyle().
		MaxWidth(width).
		MaxHeight(1).
		Inline(true)

	return rowStyle.Render(rowContent) + "\n"
}

func calculateColumnWidths(available int) (int, int) {
	k := utils.Max(60, int(float64(available)*0.75))
	if available-k-4 < 20 {
		k = available - 24
	}
	v := available - k - 4
	return k, v
}

func calculateFilteredColumnWidths(width int) (keyColWidth, valueColWidth int) {
	keyColWidth = utils.Max(80, int(float64(width)*0.8))
	if width-keyColWidth-4 < 20 {
		keyColWidth = width - 24
	}
	valueColWidth = width - keyColWidth - 4
	return keyColWidth, valueColWidth
}

func calculateNormalColumnWidths(width int) (keyColWidth, valueColWidth int) {
	available := width - 16
	keyColWidth = int(float64(available) * 0.75)
	if available-keyColWidth < 20 {
		keyColWidth = available - 20
	}
	valueColWidth = available - keyColWidth
	return keyColWidth, valueColWidth
}

func renderSelectedRow(cursor, key, val string, kWidth, vWidth, width int) string {
	const gap = "  "

	keyDisplay := truncateString(key, kWidth)
	valDisplay := truncateString(val, vWidth)

	keyPadded := padToWidth(keyDisplay, kWidth)
	rowContent := cursor + keyPadded + gap + valDisplay

	if lipgloss.Width(rowContent) > width {
		rowContent = utils.Truncate(rowContent, width)
	}

	remaining := width - lipgloss.Width(rowContent)
	if remaining > 0 {
		rowContent += strings.Repeat(" ", remaining)
	}

	return style.SelectedRow.
		MaxHeight(1).
		Render(rowContent) + "\n"
}

func getCursorIndicator(isSelected bool) string {
	if isSelected {
		return "> "
	}
	return "  "
}

func padToWidth(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth > width {
		return utils.Truncate(s, width)
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

func getRowStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return style.SelectedRow
	}
	return style.Row
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func RenderValueView(data ValueViewData) string {
	if data.SelectedKey == "" {
		return ""
	}

	tableWidth := calculatePaneWidth(data.Width, data.SplitRatio, true)
	separatorWidth := lipgloss.Width(style.Separator.Render(constants.TableSeparator))
	valueWidth := data.Width - tableWidth - separatorWidth
	valueWidth = utils.Max(constants.MinPaneWidth, valueWidth)

	var b strings.Builder
	b.WriteString("\n")
	title := lipgloss.NewStyle().Bold(true).Render("Value") + ": " + data.SelectedKey
	if data.IsJSON {
		title += " " + style.Badge.Render("[JSON]")
	}
	b.WriteString(lipgloss.NewStyle().MaxWidth(valueWidth-2).MaxHeight(1).Render(title) + "\n")
	b.WriteString(strings.Repeat("─", valueWidth-2) + "\n")

	availableHeight := utils.Max(1, data.Height-2)
	lines := wrapOrSplitValue(data, valueWidth)

	renderValueContent(&b, lines, data.ValueViewport, availableHeight, valueWidth)

	return b.String()
}

func wrapOrSplitValue(data ValueViewData, width int) []string {
	if data.IsJSON {
		displayValue := data.FormattedValue
		if displayValue == "" {
			displayValue = data.SelectedValue
		}
		return strings.Split(displayValue, "\n")
	}
	displayValue := data.SelectedValue
	if displayValue == "" {
		return []string{}
	}
	return utils.WrapText(displayValue, width-4)
}

func renderValueContent(b *strings.Builder, lines []string, viewport, availableHeight, width int) {
	if len(lines) == 0 {
		b.WriteString("\n(empty value)\n")
		paddingNeeded := utils.Max(0, availableHeight-1)
		for i := 0; i < paddingNeeded; i++ {
			b.WriteString("\n")
		}
		return
	}

	needsFooter := len(lines) > availableHeight || viewport > 0

	maxVisibleLines := availableHeight
	if needsFooter {
		maxVisibleLines = availableHeight - 1
	}

	maxStart := utils.Max(0, len(lines)-maxVisibleLines)
	start := viewport

	if start < 0 {
		start = 0
	} else if start > maxStart && maxStart >= 0 && len(lines) > maxVisibleLines {
		start = maxStart
	}

	end := utils.Min(len(lines), start+maxVisibleLines)

	contentLinesWritten := 0
	for i := start; i < end && contentLinesWritten < maxVisibleLines; i++ {
		line := lines[i]
		if len(line) > width-4 {
			line = utils.Truncate(line, width-4)
		}
		b.WriteString(line)
		b.WriteString("\n")
		contentLinesWritten++
	}

	for contentLinesWritten < maxVisibleLines {
		b.WriteString("\n")
		contentLinesWritten++
	}

	if footer := buildValueFooter(viewport, end, len(lines), availableHeight); footer != "" {
		b.WriteString(style.KeyHelpDesc.Render(footer))
	}
}

func buildValueFooter(start, end, totalLines, maxVisibleLines int) string {
	var footer []string
	if end < totalLines {
		footer = append(footer, fmt.Sprintf("↓ %d more", totalLines-end))
	}
	if start > 0 {
		footer = append(footer, fmt.Sprintf("↑ %d above", start))
	}
	if totalLines > maxVisibleLines {
		footer = append(footer, "g/G to jump")
	}
	if len(footer) == 0 {
		return ""
	}
	return strings.Join(footer, " | ")
}

func RenderSplitView(tableContent, valueContent string, width int, splitRatio float64, draggingSplit bool) string {
	tableWidth := calculatePaneWidth(width, splitRatio, true)

	separator := constants.TableSeparator
	separatorStyle := style.Separator
	if draggingSplit {
		separator = " ▐ "
		separatorStyle = style.SeparatorDrag
	}

	separatorRendered := separatorStyle.Render(separator)
	separatorWidth := lipgloss.Width(separatorRendered)

	valueWidth := width - tableWidth - separatorWidth
	valueWidth = utils.Max(constants.MinPaneWidth, valueWidth)

	tableLines := strings.Split(tableContent, "\n")
	valueLines := strings.Split(valueContent, "\n")

	maxHeight := utils.Max(len(tableLines), len(valueLines))

	var result strings.Builder
	for i := 0; i < maxHeight; i++ {
		tableLine := getOrEmpty(tableLines, i)
		valueLine := getOrEmpty(valueLines, i)

		if i == 0 && strings.TrimSpace(tableLine) == "" && strings.TrimSpace(valueLine) == "" {
			result.WriteString("\n")
			continue
		}

		tableLine = padOrTruncateLine(tableLine, tableWidth)
		valueLine = padOrTruncateLine(valueLine, valueWidth)

		combinedLine := tableLine + separatorRendered + valueLine

		combinedWidth := lipgloss.Width(combinedLine)
		if combinedWidth > width {
			combinedLine = utils.TruncateANSI(combinedLine, width)
		}

		result.WriteString(combinedLine)
		result.WriteString("\n")
	}

	return result.String()
}

func getOrEmpty(lines []string, idx int) string {
	if idx < len(lines) {
		return lines[idx]
	}
	return ""
}

func padOrTruncateLine(line string, width int) string {
	lineWidth := lipgloss.Width(line)
	if lineWidth > width {
		return utils.TruncateANSI(line, width)
	} else if lineWidth < width {
		return line + strings.Repeat(" ", width-lineWidth)
	}
	return line
}
