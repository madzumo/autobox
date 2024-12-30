package main

import (
	"fmt"

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

`
	lipHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
	lipManifestStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("56"))
)

func main() {
	fmt.Println("hello world")
	ShowMenuList(lipHeaderStyle.Render(headerMenu) + "\n" + lipManifestStyle.Render(getManifest()))
}

func getManifest() string {

	return "ok"
}
