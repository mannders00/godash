package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScriptsModel struct {
	scripts []ScriptItem
	cursor  int
	width   int
	height  int
	detail  bool
	running bool
	output  string
}

type ScriptItem struct {
	Name    string
	Command string
	Args    []string
	Remote  bool
	Host    string
	Running bool
	Status  string
	Labels  map[string]string
}

func NewScriptsModel() ScriptsModel {
	return ScriptsModel{
		scripts: []ScriptItem{
			{Name: "backup-db", Command: "./scripts/backup.sh", Remote: false, Status: "ready", Labels: map[string]string{"type": "backup"}},
			{Name: "deploy-prod", Command: "./scripts/deploy.sh", Args: []string{"prod"}, Remote: false, Status: "ready", Labels: map[string]string{"type": "deploy"}},
			{Name: "logs-tail", Command: "tail", Args: []string{"-f", "/var/log/app.log"}, Remote: false, Status: "ready", Labels: map[string]string{"type": "logs"}},
			{Name: "restart-api", Command: "systemctl", Args: []string{"restart", "api"}, Remote: true, Host: "prod", Status: "ready", Labels: map[string]string{"type": "ops"}},
		},
	}
}

func (m ScriptsModel) Items() []ScriptItem {
	return m.scripts
}

func (m ScriptsModel) Init() tea.Cmd {
	return nil
}

type scriptDoneMsg struct {
	index  int
	err    error
	output string
}

func runScriptCmd(script ScriptItem, index int) tea.Cmd {
	return func() tea.Msg {
		parts := strings.Fields(script.Command)
		if len(parts) == 0 {
			return scriptDoneMsg{index: index, err: fmt.Errorf("empty command")}
		}

		cmd := exec.Command(parts[0], append(parts[1:], script.Args...)...)
		output, err := cmd.CombinedOutput()
		return scriptDoneMsg{index: index, err: err, output: string(output)}
	}
}

func (m *ScriptsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scriptDoneMsg:
		if msg.index < len(m.scripts) {
			m.scripts[msg.index].Running = false
			if msg.err != nil {
				m.scripts[msg.index].Status = "error"
			} else {
				m.scripts[msg.index].Status = "done"
			}
			m.output = msg.output
		}
		m.running = false
		return m, nil

	case tea.KeyMsg:
		if m.detail {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m *ScriptsModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.scripts)-1 {
			m.cursor++
		}
	case "enter":
		m.detail = true
		return m, nil
	case "r":
		if m.cursor < len(m.scripts) && !m.running {
			m.scripts[m.cursor].Running = true
			m.scripts[m.cursor].Status = "running"
			m.running = true
			return m, runScriptCmd(m.scripts[m.cursor], m.cursor)
		}
	}
	return m, nil
}

func (m *ScriptsModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.detail = false
		m.output = ""
	case "r":
		if m.cursor < len(m.scripts) && !m.running {
			m.scripts[m.cursor].Running = true
			m.scripts[m.cursor].Status = "running"
			m.running = true
			return m, runScriptCmd(m.scripts[m.cursor], m.cursor)
		}
	case "e":
		if m.cursor < len(m.scripts) {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}
			cmd := exec.Command(editor, m.scripts[m.cursor].Command)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
	return m, nil
}

func (m *ScriptsModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	if m.detail && m.cursor < len(m.scripts) {
		return m.renderDetail()
	}

	var lines []string
	lines = append(lines, titleStyle.Render(" Scripts "))
	lines = append(lines, "")
	lines = append(lines, headerStyle.Render(fmt.Sprintf("%-14s %-7s %-8s %s", "NAME", "TYPE", "STATUS", "COMMAND")))
	lines = append(lines, strings.Repeat("─", m.width-4))

	for i, s := range m.scripts {
		scriptType := "local"
		if s.Remote {
			scriptType = "remote"
		}

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

		cmd := s.Command
		if len(s.Args) > 0 {
			cmd = cmd + " " + strings.Join(s.Args, " ")
		}

		line := fmt.Sprintf("%-14s %-7s %-8s %s",
			truncate(s.Name, 14),
			scriptType,
			lipgloss.NewStyle().Foreground(statusColor).Render(status),
			truncate(cmd, m.width-40),
		)

		if i == m.cursor {
			lines = append(lines, selectedStyle.Render("▶ "+line))
		} else {
			lines = append(lines, "   "+line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" j/k: navigate │ enter: details │ r: run │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m ScriptsModel) renderDetail() string {
	s := m.scripts[m.cursor]
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	statusReady := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	statusRunning := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	statusError := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf(" Script: %s ", s.Name)))
	lines = append(lines, "")

	var status string
	switch s.Status {
	case "running":
		status = statusRunning.Render("● RUNNING")
	case "error":
		status = statusError.Render("● ERROR")
	default:
		status = statusReady.Render("● READY")
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Status:"), status))

	scriptType := "local"
	if s.Remote {
		scriptType = "remote (host: " + s.Host + ")"
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Type:"), valueStyle.Render(scriptType)))

	fullCmd := s.Command
	if len(s.Args) > 0 {
		fullCmd = fullCmd + " " + strings.Join(s.Args, " ")
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Command:"), valueStyle.Render(fullCmd)))

	if len(s.Labels) > 0 {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Labels:"))
		for k, v := range s.Labels {
			lines = append(lines, fmt.Sprintf("  %s=%s", k, v))
		}
	}

	if m.output != "" {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Output:"))
		for _, line := range strings.Split(m.output, "\n") {
			if line != "" {
				lines = append(lines, "  "+truncate(line, m.width-6))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", m.width-4))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" r: run │ e: edit │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
