package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LogModel struct {
	logs   []LogEntry
	cursor int
	width  int
	height int
	follow bool
}

type LogEntry struct {
	Time    time.Time
	Level   string
	Source  string
	Message string
}

func NewLogModel() LogModel {
	return LogModel{
		follow: true,
		logs: []LogEntry{
			{Time: time.Now().Add(-5 * time.Minute), Level: "INFO", Source: "machines", Message: "Machine prod-server is online (latency: 23ms)"},
			{Time: time.Now().Add(-4 * time.Minute), Level: "INFO", Source: "sites", Message: "Site myapp.com health check: 200 OK (45ms)"},
			{Time: time.Now().Add(-3 * time.Minute), Level: "WARN", Source: "sites", Message: "Site api.myapp.com slow response: 1.2s"},
			{Time: time.Now().Add(-2 * time.Minute), Level: "INFO", Source: "webhooks", Message: "Webhook deploy triggered by user"},
			{Time: time.Now().Add(-1 * time.Minute), Level: "INFO", Source: "scripts", Message: "Script backup-db completed successfully"},
			{Time: time.Now().Add(-30 * time.Second), Level: "INFO", Source: "machines", Message: "Machine staging responded (latency: 89ms)"},
		},
	}
}

func (m LogModel) Init() tea.Cmd {
	return nil
}

func (m *LogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.follow = false
			}
		case "down", "j":
			if m.cursor < len(m.logs)-1 {
				m.cursor++
			}
		case "end":
			m.cursor = len(m.logs) - 1
			m.follow = true
		case "home":
			m.cursor = 0
			m.follow = false
		}
	}
	return m, nil
}

func (m *LogModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	levelInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	levelWarn := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	levelError := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	sourceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))

	var lines []string
	lines = append(lines, titleStyle.Render(" Activity Log "))
	lines = append(lines, "")

	for i, log := range m.logs {
		var levelStyled string
		switch log.Level {
		case "INFO":
			levelStyled = levelInfo.Render("INFO ")
		case "WARN":
			levelStyled = levelWarn.Render("WARN ")
		case "ERROR":
			levelStyled = levelError.Render("ERROR")
		}

		line := fmt.Sprintf("%s %s [%s] %s",
			timeStyle.Render(log.Time.Format("15:04:05")),
			levelStyled,
			sourceStyle.Render(log.Source),
			log.Message,
		)

		if i == m.cursor {
			lines = append(lines, selectedStyle.Render(" "+line))
		} else {
			lines = append(lines, " "+line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, timeStyle.Render(" j/k: navigate | home/end: jump | esc: back"))

	return strings.Join(lines, "\n")
}

func (m *LogModel) AddLog(level, source, message string) {
	m.logs = append(m.logs, LogEntry{
		Time:    time.Now(),
		Level:   level,
		Source:  source,
		Message: message,
	})
	if m.follow {
		m.cursor = len(m.logs) - 1
	}
}
