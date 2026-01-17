package style

import "github.com/charmbracelet/lipgloss"

const (
	black     = lipgloss.Color("#000000")
	blue      = lipgloss.Color("6")
	grey      = lipgloss.Color("#737373")
	red       = lipgloss.Color("#FF5353")
	yellow    = lipgloss.Color("#DBBD70")
	amberGold = lipgloss.Color("#D79921")
	warmGrey  = lipgloss.Color("#665C54")
	dullGreen = lipgloss.Color("#98971A")
)

var (
	Regular       = lipgloss.NewStyle()
	Bold          = Regular.Bold(true)
	Logo          = Regular.Padding(0, 0).Foreground(yellow)
	ClusterUrl    = Bold
	Header        = Regular.Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	Endpoint      = Regular.Foreground(grey)
	Status        = Regular.Foreground(lipgloss.Color("#04B575")).Bold(true)
	Error         = Regular.Foreground(red).Bold(true)
	KeyHelp       = Regular.Padding(0, 1)
	KeyHelpKey    = Regular.Foreground(blue).Bold(true)
	KeyHelpDesc   = Regular.Foreground(grey)
	FilterPrefix  = Regular.Padding(0, 3).Border(lipgloss.NormalBorder(), true)
	FilterEditing = Regular.Foreground(black).Background(blue)
	FilterApplied = Regular.Foreground(black).Background(lipgloss.Color("#00A095"))
	FilterLabel   = Regular.Foreground(lipgloss.Color("#FAFAFA")).Bold(true)
	FilterInput   = Regular.Foreground(grey).Border(lipgloss.RoundedBorder()).BorderForeground(grey).Padding(0, 1)
	TableHeader   = Regular.Foreground(amberGold).Bold(true).Underline(true)
	SelectedRow   = Regular.Foreground(lipgloss.Color("#FFFFFF")).Background(blue).Bold(true)
	Row           = Regular.Foreground(lipgloss.Color("#FAFAFA"))
	KeyColumn     = Regular.Foreground(yellow)
	ValueColumn   = Regular.Foreground(dullGreen)
	RowNumber     = Regular.Foreground(warmGrey)
	Separator     = Regular.Foreground(warmGrey)
	SeparatorDrag = Regular.Foreground(lipgloss.Color("#00D9FF")).Background(lipgloss.Color("#333333"))
	Badge         = Regular.Foreground(lipgloss.Color("#04B575")).Bold(true)
	Focused       = Regular.Foreground(lipgloss.Color("#00D9FF")).Bold(true)
)
