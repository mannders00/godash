package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type View int

const (
	ViewDashboard View = iota
	ViewMachineDetail
	ViewSiteDetail
	ViewWebhookDetail
	ViewScriptDetail
	ViewLog
)

type Model struct {
	currentView View
	focusPanel  int
	width       int
	height      int
	machines    MachinesModel
	sites       SitesModel
	webhooks    WebhooksModel
	scripts     ScriptsModel
	logs        LogModel
	dashboard   DashboardModel
	err         error
	lastRefresh time.Time
}

func NewModel() Model {
	m := Model{
		currentView: ViewDashboard,
		focusPanel:  0,
		machines:    NewMachinesModel(),
		sites:       NewSitesModel(),
		webhooks:    NewWebhooksModel(),
		scripts:     NewScriptsModel(),
		logs:        NewLogModel(),
		dashboard:   NewDashboardModel(),
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.machines.Init(),
		m.sites.Init(),
		m.webhooks.Init(),
		m.scripts.Init(),
		m.dashboard.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == ViewDashboard {
				return m, tea.Quit
			}
			m.currentView = ViewDashboard
			return m, nil
		case "tab":
			if m.currentView == ViewDashboard {
				m.focusPanel = (m.focusPanel + 1) % 4
				return m, nil
			}
		case "shift+tab":
			if m.currentView == ViewDashboard {
				m.focusPanel = (m.focusPanel - 1 + 4) % 4
				return m, nil
			}
		case "enter":
			if m.currentView == ViewDashboard {
				switch m.focusPanel {
				case 0:
					m.currentView = ViewMachineDetail
					return m, nil
				case 1:
					m.currentView = ViewSiteDetail
					return m, nil
				case 2:
					m.currentView = ViewWebhookDetail
					return m, nil
				case 3:
					m.currentView = ViewScriptDetail
					return m, nil
				}
			}
		case "esc":
			m.currentView = ViewDashboard
			return m, nil
		case "r":
			m.lastRefresh = time.Now()
			cmds = append(cmds,
				func() tea.Msg { return refreshMachinesMsg{} },
				func() tea.Msg { return refreshSitesMsg{} },
			)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.machines.width = msg.Width
		m.machines.height = msg.Height
		m.sites.width = msg.Width
		m.sites.height = msg.Height
		m.webhooks.width = msg.Width
		m.webhooks.height = msg.Height
		m.scripts.width = msg.Width
		m.scripts.height = msg.Height
		m.logs.width = msg.Width
		m.logs.height = msg.Height
		m.dashboard.SetSize(msg.Width, msg.Height)
	}

	switch m.currentView {
	case ViewDashboard:
		newD, cmd := m.dashboard.Update(msg)
		m.dashboard = *newD.(*DashboardModel)
		m.focusPanel = m.dashboard.Focused()
		cmds = append(cmds, cmd)
	case ViewMachineDetail:
		newM, cmd := m.machines.Update(msg)
		m.machines = *newM.(*MachinesModel)
		cmds = append(cmds, cmd)
	case ViewSiteDetail:
		newS, cmd := m.sites.Update(msg)
		m.sites = *newS.(*SitesModel)
		cmds = append(cmds, cmd)
	case ViewWebhookDetail:
		newW, cmd := m.webhooks.Update(msg)
		m.webhooks = *newW.(*WebhooksModel)
		cmds = append(cmds, cmd)
	case ViewScriptDetail:
		newSc, cmd := m.scripts.Update(msg)
		m.scripts = *newSc.(*ScriptsModel)
		cmds = append(cmds, cmd)
	case ViewLog:
		newL, cmd := m.logs.Update(msg)
		m.logs = *newL.(*LogModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.currentView {
	case ViewMachineDetail:
		return m.machines.View()
	case ViewSiteDetail:
		return m.sites.View()
	case ViewWebhookDetail:
		return m.webhooks.View()
	case ViewScriptDetail:
		return m.scripts.View()
	case ViewLog:
		return m.logs.View()
	default:
		m.dashboard.SetMachines(m.machines.Items())
		m.dashboard.SetSites(m.sites.Items())
		m.dashboard.SetWebhooks(m.webhooks.Items())
		m.dashboard.SetScripts(m.scripts.Items())
		return renderDashboard(m.dashboard, m.focusPanel, m.width, m.height, m.lastRefresh)
	}
}

type refreshMachinesMsg struct{}
type refreshSitesMsg struct{}

type tickMsg time.Time

func renderDashboard(m DashboardModel, focus int, width, height int, lastRefresh time.Time) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	focusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))

	panelWidth := (width - 4) / 2
	panelHeight := (height - 8) / 2

	topLeft := renderMachinesPanel(m.machines, panelWidth, panelHeight, focus == 0)
	topRight := renderSitesPanel(m.sites, panelWidth, panelHeight, focus == 1)
	bottomLeft := renderWebhooksPanel(m.webhooks, panelWidth, panelHeight, focus == 2)
	bottomRight := renderScriptsPanel(m.scripts, panelWidth, panelHeight, focus == 3)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, " ", topRight)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, bottomLeft, " ", bottomRight)
	content := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

	header := titleStyle.Render("  godash  ") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("indie hacker dashboard")

	helpBar := "│ " +
		focusStyle.Render("Enter") + " select │ " +
		focusStyle.Render("Tab") + "/" + focusStyle.Render("Shift+Tab") + " switch panel │ " +
		focusStyle.Render("r") + " refresh │ " +
		focusStyle.Render("q") + " quit"

	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(fmt.Sprintf(" %s │ Last refresh: %s ",
			time.Now().Format("15:04:05"),
			lastRefresh.Format("15:04:05")))

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		content,
		"",
		helpBar,
		statusBar,
	)
}

func renderMachinesPanel(machines []MachineItem, width, height int, focused bool) string {
	borderColor := lipgloss.Color("8")
	if focused {
		borderColor = lipgloss.Color("4")
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)

	title := " Machines "
	if focused {
		title = " Machines [M] "
		title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render(title)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(title)
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7")).
		Render(fmt.Sprintf("%-12s %-8s %s", "NAME", "STATUS", "HOST"))
	lines = append(lines, header)

	for _, m := range machines {
		var statusIcon string
		var statusColor lipgloss.Color
		switch m.Status {
		case "online":
			statusIcon = "●"
			statusColor = lipgloss.Color("2")
		case "offline":
			statusIcon = "○"
			statusColor = lipgloss.Color("1")
		default:
			statusIcon = "·"
			statusColor = lipgloss.Color("8")
		}

		latency := ""
		if m.Latency > 0 {
			latency = m.Latency.Round(time.Millisecond).String()
		}

		name := truncate(m.Name, 12)
		host := truncate(m.Host, width-28)
		line := fmt.Sprintf("%-12s %s %-8s %s %s",
			name,
			lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
			m.Status,
			host,
			lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render(latency),
		)
		lines = append(lines, line)
	}

	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderSitesPanel(sites []SiteItem, width, height int, focused bool) string {
	borderColor := lipgloss.Color("8")
	if focused {
		borderColor = lipgloss.Color("4")
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)

	title := " Sites "
	if focused {
		title = " Sites [S] "
		title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render(title)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(title)
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7")).
		Render(fmt.Sprintf("%-12s %-8s %s", "NAME", "STATUS", "LATENCY"))
	lines = append(lines, header)

	for _, s := range sites {
		var statusIcon string
		var statusColor lipgloss.Color
		if s.Healthy {
			statusIcon = "●"
			statusColor = lipgloss.Color("2")
		} else if s.StatusCode > 0 {
			statusIcon = "○"
			statusColor = lipgloss.Color("1")
		} else {
			statusIcon = "·"
			statusColor = lipgloss.Color("8")
		}

		var latency string
		if s.Latency > 0 {
			latency = s.Latency.Round(time.Millisecond).String()
		}

		name := truncate(s.Name, 12)
		line := fmt.Sprintf("%-12s %s %-8s %s",
			name,
			lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
			fmt.Sprintf("%d", s.StatusCode),
			lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render(latency),
		)
		lines = append(lines, line)
	}

	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderWebhooksPanel(webhooks []WebhookItem, width, height int, focused bool) string {
	borderColor := lipgloss.Color("8")
	if focused {
		borderColor = lipgloss.Color("4")
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)

	title := " Webhooks "
	if focused {
		title = " Webhooks [W] "
		title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render(title)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(title)
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7")).
		Render(fmt.Sprintf("%-12s %-8s %s", "NAME", "METHOD", "STATUS"))
	lines = append(lines, header)

	for _, w := range webhooks {
		statusColor := lipgloss.Color("2")
		status := "ready"
		if w.Status == "running" {
			statusColor = lipgloss.Color("3")
			status = "running"
		}

		name := truncate(w.Name, 12)
		line := fmt.Sprintf("%-12s %-8s %s",
			name,
			w.Method,
			lipgloss.NewStyle().Foreground(statusColor).Render(status),
		)
		lines = append(lines, line)
	}

	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderScriptsPanel(scripts []ScriptItem, width, height int, focused bool) string {
	borderColor := lipgloss.Color("8")
	if focused {
		borderColor = lipgloss.Color("4")
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)

	title := " Scripts "
	if focused {
		title = " Scripts [X] "
		title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render(title)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(title)
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7")).
		Render(fmt.Sprintf("%-12s %-7s %s", "NAME", "TYPE", "STATUS"))
	lines = append(lines, header)

	for _, s := range scripts {
		statusColor := lipgloss.Color("8")
		status := "ready"
		switch s.Status {
		case "running":
			statusColor = lipgloss.Color("3")
			status = "running"
		case "done":
			statusColor = lipgloss.Color("2")
			status = "done"
		case "error":
			statusColor = lipgloss.Color("1")
			status = "error"
		}

		scriptType := "local"
		if s.Remote {
			scriptType = "remote"
		}

		name := truncate(s.Name, 12)
		line := fmt.Sprintf("%-12s %-7s %s",
			name,
			scriptType,
			lipgloss.NewStyle().Foreground(statusColor).Render(status),
		)
		lines = append(lines, line)
	}

	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen && maxLen > 0 {
		return s[:maxLen-1] + "…"
	}
	return s
}
