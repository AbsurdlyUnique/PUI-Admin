package main

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI states
const (
	stateConfigWizard = iota
	stateConnected
	stateError
	stateQuerying
)

// Model contains the app state
type model struct {
	state         int                 // Current state of the TUI
	configInputs  []textinput.Model   // Input fields for configuration
	focusedInput  int                 // Which input is currently focused
	content       string              // Content shown after connecting
	errorMessage  string              // Error message in case of connection failure
	width         int                 // Terminal width
	height        int                 // Terminal height
	selectedTab   int                 // Index of the selected tab
	db            *sqlx.DB            // Database connection
	tables        []string            // List of tables
	tableRowCount map[string]int      // Row count for each table
}

// Initialize the TUI model
func initialModel() model {
	inputs := make([]textinput.Model, 5)

	// Initialize text inputs
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Username"

	// Password input field with * masking
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Password"
	inputs[1].EchoMode = textinput.EchoPassword // Mask input with *
	inputs[1].EchoCharacter = '*'               // Set the mask character to *

	// Host input field (no default value)
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Host"

	// Port input field
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Port (default 5432)"

	// Database Name input field
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Database Name"

	// Focus the first input
	inputs[0].Focus()

	return model{
		state:        stateConfigWizard,
		configInputs: inputs,
		focusedInput: 0,
		selectedTab:  0,
		tableRowCount: make(map[string]int),
	}
}

// Style definitions using Lipgloss
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true).
			MarginBottom(1)

	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			MarginBottom(1)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true).
			Margin(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true).
			MarginBottom(1)

	appBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#B4BEFE")).
			Padding(1, 2).
			Width(50).
			Align(lipgloss.Center)

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B4BEFE")).
			Bold(true).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true).
			Padding(0, 1)
)

// Init initializes the app
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and state transitions
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch m.state {

		case stateConfigWizard:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit

			case "tab", "down":
				m.focusedInput = (m.focusedInput + 1) % len(m.configInputs)
				for i := range m.configInputs {
					if i == m.focusedInput {
						m.configInputs[i].Focus()
					} else {
						m.configInputs[i].Blur()
					}
				}

			case "shift+tab", "up":
				m.focusedInput = (m.focusedInput - 1 + len(m.configInputs)) % len(m.configInputs)
				for i := range m.configInputs {
					if i == m.focusedInput {
						m.configInputs[i].Focus()
					} else {
						m.configInputs[i].Blur()
					}
				}

			case "enter":
				config := DBConfig{
					User:     m.configInputs[0].Value(),
					Password: m.configInputs[1].Value(),
					Host:     m.configInputs[2].Value(),
					Port:     m.configInputs[3].Value(),
					DBName:   m.configInputs[4].Value(),
				}

				db, err := connectToDatabase(config)
				if err != nil {
					m.errorMessage = err.Error()
					m.state = stateError
				} else {
					m.db = db
					// Fetch tables and row counts
					m.state = stateQuerying
					go m.fetchTableData()
				}
			}

		case stateError:
			switch msg.String() {
			case "enter":
				m.state = stateConfigWizard
			case "ctrl+c":
				return m, tea.Quit
			}

		case stateConnected:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "left":
				m.selectedTab = (m.selectedTab - 1 + 4) % 4 // Assuming 4 tabs
			case "right":
				m.selectedTab = (m.selectedTab + 1) % 4 // Assuming 4 tabs
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	for i := range m.configInputs {
		var cmd tea.Cmd
		m.configInputs[i], cmd = m.configInputs[i].Update(msg)
		if cmd != nil {
			return m, cmd
		}
	}

	return m, nil
}

func (m *model) fetchTableData() {

	// Fetch table list
	tables, err := fetchTables(m.db)
	if err != nil {
		m.errorMessage = fmt.Sprintf("Failed to load tables: %v", err)
		m.state = stateError
		return
	}


	// Store the tables
	m.tables = tables

	// Fetch row count for each table
	for _, table := range tables {
		rowCount, err := fetchRowCount(m.db, table)
		if err != nil {
			m.errorMessage = fmt.Sprintf("Failed to fetch row count for table %s: %v", table, err)
			m.state = stateError
			return
		}
		m.tableRowCount[table] = rowCount
	}

	// Update state to connected after data is fetched
	m.state = stateConnected
}

// View renders the UI based on the current state
func (m model) View() string {
	switch m.state {

	case stateConfigWizard:
		return m.renderConfigWizard()

	case stateConnected:
		return m.renderDatabasePanel()

	case stateError:
		return m.renderErrorUI()

	case stateQuerying:
		return m.renderLoadingUI()

	default:
		return "Invalid state"
	}
}

// renderConfigWizard renders the configuration wizard UI
func (m model) renderConfigWizard() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Configure PostgreSQL") + "\n\n")

	// Render the input fields
	for _, input := range m.configInputs {
		b.WriteString(textStyle.Render(input.View()) + "\n")
	}

	// Add a connect button
	b.WriteString("\n")
	b.WriteString(buttonStyle.Render("Connect (Press Enter)"))

	// Return the full content within the application border
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		appBorderStyle.Render(b.String()))
}

// renderLoadingUI shows a loading message while fetching data
func (m model) renderLoadingUI() string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		appBorderStyle.Render(textStyle.Render("Loading database information...")))
}

// renderErrorUI shows an error message if the connection failed
func (m model) renderErrorUI() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("Error: ") + m.errorMessage + "\n")
	b.WriteString(buttonStyle.Render("Press Enter to retry"))

	// Return the full content within the application border
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		appBorderStyle.Render(b.String()))
}

// renderDatabasePanel renders the UI after a successful connection
func (m model) renderDatabasePanel() string {
	var b strings.Builder

	// Simulated tabs at the top
	tabs := []string{
		"Dashboard",
		"Queries",
		"Logs",
		"Settings",
	}

	for i, tab := range tabs {
		if i == m.selectedTab {
			b.WriteString(activeTabStyle.Render(tab) + " ")
		} else {
			b.WriteString(tabStyle.Render(tab) + " ")
		}
	}

	b.WriteString("\n\n")

	// Main content (centered) - List of tables and row counts
	b.WriteString(textStyle.Render("Tables:\n"))
	for _, table := range m.tables {
		rowCount := m.tableRowCount[table]
		b.WriteString(fmt.Sprintf("%s: %d rows\n", table, rowCount))
	}

	// Footer
	b.WriteString("\n" + buttonStyle.Render("Press q or Ctrl+C to quit"))

	// Return the full content within the application border
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		appBorderStyle.Render(b.String()))
}

