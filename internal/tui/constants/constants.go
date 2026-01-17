package constants

import "strings"

var LogoString = strings.Join([]string{
	"░█▀▀░▀█▀░█▀▀░█▀▄░░░░░▀█▀░█░█░▀█▀",
	"░█▀▀░░█░░█░░░█░█░▄▄▄░░█░░█░█░░█░",
	"░▀▀▀░░▀░░▀▀▀░▀▀░░░░░░░▀░░▀▀▀░▀▀▀",
}, "\n")

const (
	Version        = "v0.0.1"
	MinPaneWidth   = 20
	MinSplitRatio  = 0.2
	MaxSplitRatio  = 0.8
	SplitAdjustInc = 0.05
)

const (
	TableSeparator = " │ "
)

const (
	FocusTable = "table"
	FocusValue = "value"
)
