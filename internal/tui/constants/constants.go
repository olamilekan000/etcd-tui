package constants

import "strings"

var LogoString = strings.Join([]string{
	"░█▀▀░▀█▀░█▀▀░█▀▄░░░░░▀█▀░█░█░▀█▀",
	"░█▀▀░░█░░█░░░█░█░▄▄▄░░█░░█░█░░█░",
	"░▀▀▀░░▀░░▀▀▀░▀▀░░░░░░░▀░░▀▀▀░▀▀▀",
}, "\n")

const (
	Version        = "v0.0.4"
	MinPaneWidth   = 20
	MinSplitRatio  = 0.2
	MaxSplitRatio  = 0.8
	SplitAdjustInc = 0.05
)

const (
	HeaderPadding  = 2 // Padding for headers, titles, and separator lines
	ContentPadding = 4 // Padding for content wrapping and truncation
)

const (
	TableSeparator = " │ "
)

const (
	FocusTable = "table"
	FocusValue = "value"
)

const (
	KeyEnter = "enter"
	KeyTab   = "tab"
	KeyEsc   = "esc"
	KeyUp    = "up"
	KeyCtrlC = "ctrl+c"
	KeyDown  = "down"
	KeyLeft  = "left"
	KeyRight = "right"
	KeyQ     = "q"
	KeyR     = "r"
	KeyRCaps = "R"
	KeyC     = "c"
	KeyY     = "y"
	KeyG     = "g"
	KeyGCaps = "G"
	KeyH     = "h"
	KeyJ     = "j"
	KeyK     = "k"
	KeyL     = "l"
	KeySlash = "/"
)
