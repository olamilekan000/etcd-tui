package header

import (
	"strings"

	"github.com/olamilekan000/etcd-tui/internal/tui/style"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	logo, logoColor, endpoint, version, keyHelp string
	compact                                     bool
	width                                       int
}

func New(logo string, logoColor string, endpoint, version, keyHelp string) Model {
	return Model{logo: logo, logoColor: logoColor, endpoint: endpoint, version: version, keyHelp: keyHelp}
}

func (m Model) View() string {
	logoStyle := style.Logo
	if m.logoColor != "" {
		logoStyle = logoStyle.Foreground(lipgloss.Color(m.logoColor))
	}
	clusterUrl := style.ClusterUrl.Render(m.endpoint)
	if m.compact {
		versionStyle := style.Regular.Padding(0, 2, 0, 0)
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			logoStyle.Padding(0).Margin(0).Render("ETCD TUI"),
			style.KeyHelp.Render(m.keyHelp),
			versionStyle.Render(m.version),
			clusterUrl,
		) + "\n"
	}
	logo := logoStyle.Render(m.logo)
	left := style.Header.Render(lipgloss.JoinVertical(lipgloss.Center, logo, m.version, clusterUrl))

	keyHelpLines := strings.Split(m.keyHelp, "\n")
	var keyHelpRows []string
	for _, line := range keyHelpLines {
		if line != "" {
			keyHelpRows = append(keyHelpRows, style.KeyHelp.Render(line))
		}
	}
	keyHelpVertical := lipgloss.JoinVertical(lipgloss.Top, keyHelpRows...)

	leftWidth := lipgloss.Width(left)
	keyHelpWidth := lipgloss.Width(keyHelpVertical)
	availableWidth := m.width
	if availableWidth == 0 {
		availableWidth = 120
	}
	spacing := availableWidth - leftWidth - keyHelpWidth
	if spacing < 1 {
		spacing = 1
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		strings.Repeat(" ", spacing),
		keyHelpVertical,
	) + "\n"
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m *Model) SetKeyHelp(keyHelp string) {
	m.keyHelp = keyHelp
}

func (m *Model) ToggleCompact() {
	m.compact = !m.compact
}

func (m *Model) SetEndpoint(endpoint string) {
	m.endpoint = endpoint
}

func (m *Model) SetWidth(width int) {
	m.width = width
}
