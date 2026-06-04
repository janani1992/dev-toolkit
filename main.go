package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	regex     regexVaultModel
}

type regexVaultModel struct {
	inputs      [2]textinput.Model
	focusIndex  int
	compiled    *regexp.Regexp
	compileErr  string
	highlighted string
}

func initialModel() model {
	patternInput := textinput.New()
	patternInput.Prompt = "Regex Pattern: "
	patternInput.Placeholder = `e.g. \b\w+@\w+\.\w+\b`
	patternInput.Focus()
	patternInput.CharLimit = 256

	testInput := textinput.New()
	testInput.Prompt = "Test String:   "
	testInput.Placeholder = "Type text to evaluate against the regex"
	testInput.CharLimit = 512

	rv := regexVaultModel{
		inputs:     [2]textinput.Model{patternInput, testInput},
		focusIndex: 0,
	}
	rv.recompute()

	return model{
		activeTab: tabRegexVault,
		isTyping:  true,
		regex:     rv,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.activeTab == tabRegexVault {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "tab", "shift+tab":
				// Let global tab navigation handle these keys.
				goto GlobalKeys
			}
		}

		var cmd tea.Cmd
		m.regex, cmd = m.regex.Update(msg)
		m.isTyping = true

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "q", "ctrl+c":
				return m, nil
			}
		}

		return m, cmd
	}

	m.isTyping = false

	GlobalKeys:
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.activeTab = (m.activeTab + 1) % tabCount
			if m.activeTab == tabRegexVault {
				m.isTyping = true
			}
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			if m.activeTab == tabRegexVault {
				m.isTyping = true
			}
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
		body = m.regex.View()
	case tabGitAnalytics:
		body = "Git Analytics is active."
	default:
		body = "Unknown tab."
	}

	help := "\n\nTab: next tab | Shift+Tab: previous tab | q/ctrl+c: quit"
	if m.activeTab == tabRegexVault {
		help = "\n\nRegex Vault Controls: up/down or k/j switch input focus"
	}

	return header + "\n\n" + body + help + "\n"
}

func (r regexVaultModel) Update(msg tea.Msg) (regexVaultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			r.focusIndex = (r.focusIndex - 1 + len(r.inputs)) % len(r.inputs)
			r.updateFocus()
			return r, nil
		case "down", "j":
			r.focusIndex = (r.focusIndex + 1) % len(r.inputs)
			r.updateFocus()
			return r, nil
		}
	}

	updated, cmd := r.inputs[r.focusIndex].Update(msg)
	r.inputs[r.focusIndex] = updated
	r.recompute()
	return r, cmd
}

func (r *regexVaultModel) updateFocus() {
	for i := range r.inputs {
		if i == r.focusIndex {
			r.inputs[i].Focus()
		} else {
			r.inputs[i].Blur()
		}
	}
}

func (r *regexVaultModel) recompute() {
	pattern := r.inputs[0].Value()
	testStr := r.inputs[1].Value()

	r.compileErr = ""
	r.compiled = nil
	r.highlighted = testStr

	if pattern == "" {
		return
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		r.compileErr = err.Error()
		return
	}

	r.compiled = compiled
	r.highlighted = highlightMatches(compiled, testStr)
}

func (r regexVaultModel) View() string {
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	patternTitle := inactiveStyle.Render("Pattern")
	testTitle := inactiveStyle.Render("Test String")
	if r.focusIndex == 0 {
		patternTitle = activeStyle.Render("Pattern")
	}
	if r.focusIndex == 1 {
		testTitle = activeStyle.Render("Test String")
	}

	status := inactiveStyle.Render("Regex compiles successfully.")
	if r.compileErr != "" {
		status = errorStyle.Render("Invalid regex: " + r.compileErr)
	}

	return strings.Join([]string{
		patternTitle,
		r.inputs[0].View(),
		"",
		testTitle,
		r.inputs[1].View(),
		"",
		"Matches:",
		r.highlighted,
		"",
		status,
	}, "\n")
}

func highlightMatches(re *regexp.Regexp, s string) string {
	if s == "" {
		return ""
	}

	matches := re.FindAllStringIndex(s, -1)
	if len(matches) == 0 {
		return s
	}

	highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62"))

	var b strings.Builder
	last := 0
	for _, m := range matches {
		start, end := m[0], m[1]
		if start < last {
			continue
		}
		if start > len(s) || end > len(s) || start > end {
			continue
		}
		if start == end {
			continue
		}

		b.WriteString(s[last:start])
		b.WriteString(highlightStyle.Render(s[start:end]))
		last = end
	}
	b.WriteString(s[last:])

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run app: %v\n", err)
		os.Exit(1)
	}
}
