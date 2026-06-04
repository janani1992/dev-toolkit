package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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
	width     int
	height    int
	regex     regexVaultModel
	git       gitAnalyticsModel
}

type regexVaultModel struct {
	inputs      [2]textinput.Model
	focusIndex  int
	width       int
	height      int
	compiled    *regexp.Regexp
	compileErr  string
	highlighted string
}

type gitRepoStat struct {
	Path         string
	TotalCommits int
	TopDays      []dayCount
}

type dayCount struct {
	Day   string
	Count int
}

type gitScanProgressMsg struct {
	RepoPath     string
	ReposScanned int
	RepoStat     gitRepoStat
}

type gitScanSuccessMsg struct {
	RootPath     string
	ReposScanned int
	TotalCommits int
	TopDays      []dayCount
	RepoStats    []gitRepoStat
	Duration     time.Duration
}

type gitScanErrorMsg struct {
	Err string
}

type gitAnalyticsModel struct {
	pathInput    textinput.Model
	width        int
	height       int
	scanning     bool
	reposScanned int
	lastRepo     string
	scanErr      string
	result       *gitScanSuccessMsg
	scanCh       <-chan tea.Msg
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

	gitPathInput := textinput.New()
	gitPathInput.Prompt = "Root Path: "
	gitPathInput.Placeholder = "/Users/username/Projects"
	gitPathInput.CharLimit = 1024

	gav := gitAnalyticsModel{
		pathInput: gitPathInput,
	}

	initialWidth, initialHeight := 80, 24
	rv.setSize(initialWidth-2, initialHeight-7)
	gav.setSize(initialWidth-2, initialHeight-7)

	return model{
		activeTab: tabRegexVault,
		isTyping:  true,
		width:     initialWidth,
		height:    initialHeight,
		regex:     rv,
		git:       gav,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "ctrl+x":
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		bodyWidth, bodyHeight := m.bodyDimensions()
		m.regex.setSize(bodyWidth, bodyHeight)
		m.git.setSize(bodyWidth, bodyHeight)
		return m, nil
	}

	if isGitScanMsg(msg) {
		var cmd tea.Cmd
		m.git, cmd = m.git.Update(msg)
		return m, cmd
	}

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

	if m.activeTab == tabGitAnalytics {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "tab", "shift+tab":
				goto GlobalKeys
			}
		}

		var cmd tea.Cmd
		m.git, cmd = m.git.Update(msg)
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
			m.syncTabFocus()
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			m.syncTabFocus()
		case "ctrl+c", "q":
			if !m.isTyping {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m *model) syncTabFocus() {
	switch m.activeTab {
	case tabRegexVault:
		m.regex.updateFocus()
		m.git.pathInput.Blur()
		m.isTyping = true
	case tabGitAnalytics:
		for i := range m.regex.inputs {
			m.regex.inputs[i].Blur()
		}
		m.git.pathInput.Focus()
		m.isTyping = true
	default:
		m.isTyping = false
	}
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
		body = m.git.View()
	default:
		body = "Unknown tab."
	}

	help := "\n\nTab: next tab | Shift+Tab: previous tab | Ctrl+X/Ctrl+C: quit"
	if m.activeTab == tabRegexVault {
		help = "\n\nRegex Vault Controls: up/down switch input focus"
	} else if m.activeTab == tabGitAnalytics {
		help = "\n\nGit Analytics Controls: Enter starts scan | Tab switches tabs"
	}

	composed := header + "\n\n" + body + help + "\n"
	return responsiveRender(composed, clampMin(m.width, 20), clampMin(m.height, 8))
}

func (r regexVaultModel) Update(msg tea.Msg) (regexVaultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			r.focusIndex = (r.focusIndex - 1 + len(r.inputs)) % len(r.inputs)
			r.updateFocus()
			return r, nil
		case "down":
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

func (r *regexVaultModel) setSize(width, height int) {
	r.width = clampMin(width, 20)
	r.height = clampMin(height, 8)

	patternPromptWidth := lipgloss.Width(r.inputs[0].Prompt)
	testPromptWidth := lipgloss.Width(r.inputs[1].Prompt)
	inputWidth := clampMin(r.width-maxInt(patternPromptWidth, testPromptWidth)-2, 10)
	r.inputs[0].Width = inputWidth
	r.inputs[1].Width = inputWidth
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

	content := strings.Join([]string{
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

	return responsiveRender(content, clampMin(r.width, 20), clampMin(r.height, 8))
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

func (g gitAnalyticsModel) Update(msg tea.Msg) (gitAnalyticsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if g.scanning {
				return g, nil
			}

			rootPath := strings.TrimSpace(g.pathInput.Value())
			if rootPath == "" {
				g.scanErr = "Please enter an absolute directory path."
				return g, nil
			}
			if !filepath.IsAbs(rootPath) {
				g.scanErr = "Path must be absolute."
				return g, nil
			}

			stat, err := os.Stat(rootPath)
			if err != nil || !stat.IsDir() {
				g.scanErr = "Path is not a readable directory."
				return g, nil
			}

			g.scanning = true
			g.reposScanned = 0
			g.lastRepo = ""
			g.scanErr = ""
			g.result = nil

			scanCh := make(chan tea.Msg, 32)
			g.scanCh = scanCh
			go runGitAnalyticsScan(rootPath, scanCh)

			return g, waitForGitScanMsg(scanCh)
		}

		updated, cmd := g.pathInput.Update(msg)
		g.pathInput = updated
		return g, cmd

	case gitScanProgressMsg:
		g.reposScanned = msg.ReposScanned
		g.lastRepo = msg.RepoPath
		return g, waitForGitScanMsg(g.scanCh)

	case gitScanSuccessMsg:
		g.scanning = false
		g.result = &msg
		g.reposScanned = msg.ReposScanned
		g.lastRepo = ""
		g.scanErr = ""
		g.scanCh = nil
		return g, nil

	case gitScanErrorMsg:
		g.scanning = false
		g.scanErr = msg.Err
		g.scanCh = nil
		return g, nil
	}

	return g, nil
}

func (g gitAnalyticsModel) View() string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	lines := []string{
		labelStyle.Render("Git Analytics"),
		g.pathInput.View(),
	}

	if g.scanning {
		status := fmt.Sprintf("Scanning: %d repos found...", g.reposScanned)
		if g.lastRepo != "" {
			status += " Last: " + g.lastRepo
		}
		lines = append(lines, "", metaStyle.Render(status))
	} else if g.result != nil {
		lines = append(lines,
			"",
			metaStyle.Render(fmt.Sprintf("Scan complete in %s", g.result.Duration.Round(10*time.Millisecond))),
			fmt.Sprintf("Root: %s", g.result.RootPath),
			fmt.Sprintf("Repositories: %d", g.result.ReposScanned),
			fmt.Sprintf("Total commits: %d", g.result.TotalCommits),
			formatTopDaysLine(g.result.TopDays),
			"",
			"Repositories:",
		)

		for _, repo := range g.result.RepoStats {
			lines = append(lines, fmt.Sprintf("- %s | commits: %d | %s", repo.Path, repo.TotalCommits, formatTopDaysInline(repo.TopDays)))
		}
	}

	if g.scanErr != "" {
		lines = append(lines, "", errorStyle.Render(g.scanErr))
	}

	if !g.scanning && g.result == nil && g.scanErr == "" {
		lines = append(lines, "", metaStyle.Render("Enter an absolute path and press Enter to scan local repositories."))
	}

	content := strings.Join(lines, "\n")
	return responsiveRender(content, clampMin(g.width, 20), clampMin(g.height, 8))
}

func (g *gitAnalyticsModel) setSize(width, height int) {
	g.width = clampMin(width, 20)
	g.height = clampMin(height, 8)
	inputWidth := clampMin(g.width-lipgloss.Width(g.pathInput.Prompt)-2, 10)
	g.pathInput.Width = inputWidth
}

func (m model) bodyDimensions() (int, int) {
	return clampMin(m.width-2, 20), clampMin(m.height-7, 8)
}

func clampMin(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func responsiveRender(text string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	return lipgloss.NewStyle().
		MaxWidth(width).
		MaxHeight(height).
		Render(text)
}

func isGitScanMsg(msg tea.Msg) bool {
	switch msg.(type) {
	case gitScanProgressMsg, gitScanSuccessMsg, gitScanErrorMsg:
		return true
	default:
		return false
	}
}

func waitForGitScanMsg(scanCh <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if scanCh == nil {
			return nil
		}
		msg, ok := <-scanCh
		if !ok {
			return nil
		}
		return msg
	}
}

func runGitAnalyticsScan(rootPath string, out chan<- tea.Msg) {
	defer close(out)

	start := time.Now()
	repoStats := make([]gitRepoStat, 0)
	aggregatedDayCounts := make(map[string]int)
	totalCommits := 0
	reposScanned := 0

	walkErr := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repoStat, dayCounts, statErr := analyzeGitRepository(repoPath)
			if statErr == nil {
				repoStats = append(repoStats, repoStat)
				totalCommits += repoStat.TotalCommits
				for day, count := range dayCounts {
					aggregatedDayCounts[day] += count
				}
			}

			reposScanned++
			out <- gitScanProgressMsg{
				RepoPath:     repoPath,
				ReposScanned: reposScanned,
				RepoStat:     repoStat,
			}

			return filepath.SkipDir
		}

		return nil
	})

	if walkErr != nil {
		out <- gitScanErrorMsg{Err: walkErr.Error()}
		return
	}

	out <- gitScanSuccessMsg{
		RootPath:     rootPath,
		ReposScanned: reposScanned,
		TotalCommits: totalCommits,
		TopDays:      topDaysFromMap(aggregatedDayCounts),
		RepoStats:    repoStats,
		Duration:     time.Since(start),
	}
}

func analyzeGitRepository(repoPath string) (gitRepoStat, map[string]int, error) {
	countCmd := exec.Command("git", "-C", repoPath, "rev-list", "--all", "--count")
	countOut, err := countCmd.Output()
	if err != nil {
		return gitRepoStat{}, nil, err
	}

	countStr := strings.TrimSpace(string(countOut))
	if countStr == "" {
		countStr = "0"
	}

	totalCommits, err := strconv.Atoi(countStr)
	if err != nil {
		return gitRepoStat{}, nil, err
	}

	logCmd := exec.Command("git", "-C", repoPath, "log", "--all", "--pretty=format:%ad", "--date=format:%A")
	logOut, err := logCmd.Output()
	if err != nil {
		return gitRepoStat{}, nil, err
	}

	dayCounts := make(map[string]int)
	for _, line := range strings.Split(strings.TrimSpace(string(logOut)), "\n") {
		day := strings.TrimSpace(line)
		if day == "" {
			continue
		}
		dayCounts[day]++
	}

	repoStat := gitRepoStat{
		Path:         repoPath,
		TotalCommits: totalCommits,
		TopDays:      topDaysFromMap(dayCounts),
	}

	return repoStat, dayCounts, nil
}

func topDaysFromMap(dayCounts map[string]int) []dayCount {
	orderedDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	result := make([]dayCount, 0, len(dayCounts))

	for day, count := range dayCounts {
		result = append(result, dayCount{Day: day, Count: count})
	}

	orderIndex := make(map[string]int, len(orderedDays))
	for i, day := range orderedDays {
		orderIndex[day] = i
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return orderIndex[result[i].Day] < orderIndex[result[j].Day]
		}
		return result[i].Count > result[j].Count
	})

	if len(result) > 3 {
		return result[:3]
	}

	return result
}

func formatTopDaysLine(days []dayCount) string {
	if len(days) == 0 {
		return "Top active days: n/a"
	}
	return "Top active days: " + formatTopDaysInline(days)
}

func formatTopDaysInline(days []dayCount) string {
	if len(days) == 0 {
		return "n/a"
	}
	parts := make([]string, 0, len(days))
	for _, day := range days {
		parts = append(parts, fmt.Sprintf("%s(%d)", day.Day, day.Count))
	}
	return strings.Join(parts, ", ")
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run app: %v\n", err)
		os.Exit(1)
	}
}
