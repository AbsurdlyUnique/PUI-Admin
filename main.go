// Julian: Entry point of the AbsurdPG-TUI application.
// Julian: This starts the TUI and allows the user to select or create configurations

package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Entry point of the AbsurdPG-TUI application.
// This starts the TUI and allows the user to select or create configurations.

func main() {
	// Initialize and start the Bubble Tea program
	p := tea.NewProgram(initialModel())

	// Run the TUI program and handle any errors
	if err := p.Start(); err != nil {
		log.Println("Error starting the TUI:", err)
		os.Exit(1)
	}
}

