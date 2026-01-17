package keymap

import (
	"strings"

	"github.com/olamilekan000/etcd-tui/internal/tui/style"
)

func GenerateKeyHelp(showValue bool) string {
	getShortHelp := func(shortcuts []string) string {
		var output string
		for _, sc := range shortcuts {
			parts := strings.Fields(sc)
			if len(parts) >= 2 {
				key := parts[0]
				desc := strings.Join(parts[1:], " ")
				output += style.KeyHelpKey.Render(key) + " " + style.KeyHelpDesc.Render(desc) + "  "
			}
		}
		return strings.TrimSpace(output)
	}

	firstRow := []string{"q/ctrl+c exit", "r refresh", "/ filter", "c copy"}
	secondRow := []string{"↑/k up", "↓/j down", "g top", "G bottom"}
	thirdRow := []string{"tab focus", "enter view", "esc back"}

	var rows []string
	rows = append(rows, getShortHelp(firstRow))
	rows = append(rows, getShortHelp(secondRow))
	rows = append(rows, getShortHelp(thirdRow))

	if showValue {
		fourthRow := []string{"←/h shrink", "→/l expand"}
		rows = append(rows, getShortHelp(fourthRow))
	}

	return strings.Join(rows, "\n")
}
