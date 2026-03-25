package tui

import (
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SitesModel struct {
	sites  []SiteItem
	cursor int
	width  int
	height int
}

type SiteItem struct {
	Name       string
	URL        string
	StatusCode int
	Latency    time.Duration
	Healthy    bool
	LastCheck  time.Time
}

func NewSitesModel() SitesModel {
	return SitesModel{
		sites: []SiteItem{
			{Name: "myapp.com", URL: "https://myapp.com", Healthy: false, StatusCode: 0},
			{Name: "api.myapp.com", URL: "https://api.myapp.com/health", Healthy: false, StatusCode: 0},
		},
	}
}

func (m SitesModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
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

func (m SitesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m, tea.Tick(time.Second*30, func(t time.Time) tea.Msg {
			return sitesTickMsg(t)
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sites)-1 {
				m.cursor++
			}
		case "r":
			return m, checkSites(m.sites)
		}
	}
	return m, nil
}

func (m SitesModel) View() string {
	var lines []string
	lines = append(lines, StyleTitle.Render("Sites"))
	lines = append(lines, "")
	lines = append(lines, StyleHeader.Render(fmt.Sprintf("%-20s %-35s %-10s %s", "NAME", "URL", "STATUS", "LATENCY")))
	lines = append(lines, StyleBorder.Render(string(make([]byte, m.width-2))))

	for i, site := range m.sites {
		status := fmt.Sprintf("%d", site.StatusCode)
		if !site.Healthy {
			status = StyleError.Render(status)
		} else {
			status = StyleSuccess.Render(status)
		}
		line := fmt.Sprintf("%-20s %-35s %-10s %s",
			site.Name,
			site.URL,
			status,
			site.Latency.Round(time.Millisecond),
		)
		if i == m.cursor {
			lines = append(lines, StyleSelected.Render("> "+line))
		} else {
			prefix := "·"
			if site.Healthy {
				prefix = StyleSuccess.Render("●")
			} else {
				prefix = StyleError.Render("○")
			}
			lines = append(lines, prefix+" "+line[len(prefix)+1:])
		}
	}

	lines = append(lines, "")
	lines = append(lines, StyleMuted.Render("r: refresh | j/k: navigate | o: open in browser | esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
