package tui

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SitesModel struct {
	sites  []SiteItem
	cursor int
	width  int
	height int
	detail bool
}

type SiteItem struct {
	Name       string
	URL        string
	StatusCode int
	Latency    time.Duration
	Healthy    bool
	LastCheck  time.Time
	Labels     map[string]string
}

func NewSitesModel() SitesModel {
	return SitesModel{
		sites: []SiteItem{
			{Name: "myapp.com", URL: "https://myapp.com", Healthy: false, StatusCode: 0, Labels: map[string]string{"env": "prod"}},
			{Name: "api.myapp.com", URL: "https://api.myapp.com/health", Healthy: false, StatusCode: 0, Labels: map[string]string{"env": "prod", "type": "api"}},
			{Name: "docs.myapp.com", URL: "https://docs.myapp.com", Healthy: false, StatusCode: 0, Labels: map[string]string{"env": "prod", "type": "docs"}},
		},
	}
}

func (m SitesModel) Items() []SiteItem {
	return m.sites
}

func (m SitesModel) Init() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return sitesTickMsg(t)
	})
}

type sitesTickMsg time.Time

type siteCheckedMsg struct {
	index int
	item  SiteItem
}

func checkSites(sites []SiteItem) tea.Cmd {
	cmds := make([]tea.Cmd, len(sites))
	for i := range sites {
		cmds[i] = func(idx int, url string) tea.Cmd {
			return func() tea.Msg {
				start := time.Now()
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(url)
				item := SiteItem{URL: url}
				if err != nil {
					item.Healthy = false
					item.StatusCode = 0
				} else {
					item.StatusCode = resp.StatusCode
					item.Healthy = resp.StatusCode >= 200 && resp.StatusCode < 300
					resp.Body.Close()
				}
				item.Latency = time.Since(start)
				item.LastCheck = time.Now()
				return siteCheckedMsg{index: idx, item: item}
			}
		}(i, sites[i].URL)
	}
	return tea.Batch(cmds...)
}

func (m *SitesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sitesTickMsg:
		return m, checkSites(m.sites)
	case siteCheckedMsg:
		if msg.index < len(m.sites) {
			m.sites[msg.index].StatusCode = msg.item.StatusCode
			m.sites[msg.index].Healthy = msg.item.Healthy
			m.sites[msg.index].Latency = msg.item.Latency
			m.sites[msg.index].LastCheck = msg.item.LastCheck
		}
		return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
			return sitesTickMsg(t)
		})
	case tea.KeyMsg:
		if m.detail {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m *SitesModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.sites)-1 {
			m.cursor++
		}
	case "enter":
		m.detail = true
		return m, nil
	case "r":
		return m, checkSites(m.sites)
	}
	return m, nil
}

func (m *SitesModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.detail = false
	}
	return m, nil
}

func (m *SitesModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	if m.detail && m.cursor < len(m.sites) {
		return m.renderDetail()
	}

	var lines []string
	lines = append(lines, titleStyle.Render(" Sites "))
	lines = append(lines, "")
	lines = append(lines, headerStyle.Render(fmt.Sprintf("%-14s %-8s %-10s %s", "NAME", "STATUS", "LATENCY", "URL")))
	lines = append(lines, strings.Repeat("─", m.width-4))

	for i, site := range m.sites {
		var statusIcon string
		var statusColor lipgloss.Color
		if site.Healthy {
			statusIcon = "●"
			statusColor = lipgloss.Color("2")
		} else if site.StatusCode > 0 {
			statusIcon = "○"
			statusColor = lipgloss.Color("1")
		} else {
			statusIcon = "·"
			statusColor = lipgloss.Color("8")
		}

		var latency string
		if site.Latency > 0 {
			latency = site.Latency.Round(time.Millisecond).String()
		} else {
			latency = "-"
		}

		line := fmt.Sprintf("%-14s %s %-8s %-10s %s",
			truncate(site.Name, 14),
			lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
			fmt.Sprintf("%d", site.StatusCode),
			latency,
			truncate(site.URL, m.width-50),
		)

		if i == m.cursor {
			lines = append(lines, selectedStyle.Render("▶ "+line))
		} else {
			lines = append(lines, "   "+line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" j/k: navigate │ enter: details │ r: refresh │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m SitesModel) renderDetail() string {
	site := m.sites[m.cursor]
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	statusHealthy := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	statusUnhealthy := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf(" Site: %s ", site.Name)))
	lines = append(lines, "")

	var status string
	if site.Healthy {
		status = statusHealthy.Render("● HEALTHY")
	} else {
		status = statusUnhealthy.Render("○ UNHEALTHY")
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Status:"), status))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Code:"), valueStyle.Render(fmt.Sprintf("%d", site.StatusCode))))

	var latency string
	if site.Latency > 0 {
		latency = site.Latency.Round(time.Millisecond).String()
	} else {
		latency = "-"
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Latency:"), valueStyle.Render(latency)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("URL:"), valueStyle.Render(site.URL)))

	if len(site.Labels) > 0 {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Labels:"))
		for k, v := range site.Labels {
			lines = append(lines, fmt.Sprintf("  %s=%s", k, v))
		}
	}

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", m.width-4))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" o: open in browser │ r: refresh │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
