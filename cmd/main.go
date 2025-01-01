package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	headerMenu = `

+========================================================+
|    _              _               ____     ___   __  __|
|   / \     _   _  | |_    ___     | __ )   / _ \  \ \/ /|
|  / _ \   | | | | | __|  / _ \    |  _ \  | | | |  \  / |
| / ___ \  | |_| | | |_  | (_) |   | |_) | | |_| |  /  \ |
|/_/   \_\  \__,_|  \__|  \___/    |____/   \___/  /_/\_\|
+========================================================+
												by madzumo
`
	lipHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
	lipManifestStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
)

func main() {
	// fmt.Println("hello world")
	ShowMenuList(lipHeaderStyle.Render(headerMenu) + lipManifestStyle.Render(getManifest()))
}

func getManifest() string {

	return "\nManifest:"
}
