package tui

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/config"
	"github.com/spf13/viper"
)

// OrganizationListModel represents the model for organization selection
type OrganizationListModel struct {
	Organizations []cerebras.Organization
	Cursor        int
	Selected      map[int]struct{}
}

// NewOrganizationListModel creates a new organization list model
func NewOrganizationListModel(orgs []cerebras.Organization) OrganizationListModel {
	return OrganizationListModel{
		Organizations: orgs,
		Selected:      make(map[int]struct{}),
	}
}

// Init initializes the model
func (m OrganizationListModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update handles messages in the model
func (m OrganizationListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Organizations)-1 {
				m.Cursor++
			}
		case "enter", " ":
			// Select the organization
			org := m.Organizations[m.Cursor]

			// Save the organization ID to configuration
			viper.Set("org-id", org.ID)
			if err := viper.WriteConfig(); err != nil {
				// If config file doesn't exist, create it
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					configDir := config.GetConfigDir()
					configPath := filepath.Join(configDir, "settings.yaml")
					if err := viper.WriteConfigAs(configPath); err != nil {
						fmt.Printf("Error saving configuration: %v\n", err)
						return m, tea.Quit
					}
				} else {
					fmt.Printf("Error saving configuration: %v\n", err)
					return m, tea.Quit
				}
			}

			fmt.Printf("Organization %s selected and saved to configuration.\n", org.Name)
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the model
func (m OrganizationListModel) View() string {
	s := "Select an organization:\n\n"

	for i, org := range m.Organizations {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s (ID: %s)\n", cursor, org.Name, org.ID)
	}

	s += "\nPress 'enter' or 'space' to select an organization, 'q' to quit.\n"

	return s
}
