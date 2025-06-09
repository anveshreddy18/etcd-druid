package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Error running TUI:", err)
	}
}
