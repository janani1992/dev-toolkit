package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type tabView int

const (
	tabRegexVault tabView = iota
	tabGitAnalytics
	tabCount
)

type model struct {
	activeTab tabView
	isTyping  bool
}

func initialModel() model {
	return model{activeTab: tabRegexVault}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.activeTab = (m.activeTab + 1) % tabCount
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
		case "ctrl+c", "q":
			if !m.isTyping {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	tabs := []string{"Regex Vault", "Git Analytics"}
	parts := make([]string, 0, len(tabs))

	for i, t := range tabs {
		if tabView(i) == m.activeTab {
			parts = append(parts, "["+t+"]")
		} else {
			parts = append(parts, " "+t+" ")
		}
	}

	header := "Developer Toolkit\n" + strings.Join(parts, " | ")

	body := ""
	switch m.activeTab {
	case tabRegexVault:
		body = "Regex Vault is active."
	case tabGitAnalytics:
		body = "Git Analytics is active."
	default:
		body = "Unknown tab."
	}

	help := "\n\nTab: next tab | Shift+Tab: previous tab | q/ctrl+c: quit"

	return header + "\n\n" + body + help + "\n"
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run app: %v\n", err)
		os.Exit(1)
	}
}
