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
	scripts  []ScriptItem
	cursor   int
	width    int
	height   int
	running  bool
	output   string
	selected int
}

type ScriptItem struct {
	Name    string
	Command string
	Args    []string
	Remote  bool
	Host    string
	Running bool
	Output  string
	Status  string
}

func NewScriptsModel() ScriptsModel {
	return ScriptsModel{
		scripts: []ScriptItem{
			{Name: "backup-db", Command: "./scripts/backup.sh", Remote: false, Status: "ready"},
			{Name: "deploy-prod", Command: "ssh", Args: []string{"prod", "deploy.sh"}, Remote: true, Host: "prod", Status: "ready"},
			{Name: "logs-tail", Command: "tail", Args: []string{"-f", "/var/log/app.log"}, Remote: false, Status: "ready"},
		},
	}
}

func (m ScriptsModel) Init() tea.Cmd {
	return nil
}

type scriptOutputMsg struct {
	index  int
	output string
	err    error
}

type scriptDoneMsg struct {
	index  int
	err    error
	output string
}

func runScript(script ScriptItem, index int) tea.Cmd {
	return func() tea.Msg {
		parts := strings.Fields(script.Command)
		if len(parts) == 0 {
			return scriptDoneMsg{index: index, err: fmt.Errorf("empty command")}
		}

		cmd := exec.Command(parts[0], append(parts[1:], script.Args...)...)
		cmd.Stdout = nil
		cmd.Stderr = nil

		output, err := cmd.CombinedOutput()
		return scriptDoneMsg{index: index, err: err, output: string(output)}
	}
}

func (m ScriptsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scriptDoneMsg:
		if msg.index < len(m.scripts) {
			m.scripts[msg.index].Running = false
			m.scripts[msg.index].Status = "done"
			if msg.err != nil {
				m.scripts[msg.index].Status = "error"
			}
			m.output = msg.output
		}
		m.running = false
		return m, nil

	case tea.KeyMsg:
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
			if m.cursor < len(m.scripts) && !m.running {
				m.scripts[m.cursor].Running = true
				m.scripts[m.cursor].Status = "running"
				m.running = true
				m.output = ""
				return m, runScript(m.scripts[m.cursor], m.cursor)
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
	}
	return m, nil
}

func (m ScriptsModel) View() string {
	var lines []string
	lines = append(lines, StyleTitle.Render("Scripts"))
	lines = append(lines, "")
	lines = append(lines, StyleHeader.Render(fmt.Sprintf("%-20s %-10s %-10s %s", "NAME", "TYPE", "STATUS", "COMMAND")))
	lines = append(lines, StyleBorder.Render(string(make([]byte, m.width-2))))

	for i, s := range m.scripts {
		scriptType := "local"
		if s.Remote {
			scriptType = "remote"
		}

		status := s.Status
		switch status {
		case "running":
			status = StyleWarning.Render("running")
		case "done":
			status = StyleSuccess.Render("done")
		case "error":
			status = StyleError.Render("error")
		default:
			status = StyleMuted.Render("ready")
		}

		line := fmt.Sprintf("%-20s %-10s %-10s %s",
			s.Name,
			scriptType,
			status,
			truncate(s.Command+" "+strings.Join(s.Args, " "), m.width-50),
		)
		if i == m.cursor {
			lines = append(lines, StyleSelected.Render("> "+line))
		} else {
			lines = append(lines, "  "+line)
		}
	}

	if m.output != "" {
		lines = append(lines, "")
		lines = append(lines, StyleHeader.Render("Output:"))
		for _, line := range strings.Split(m.output, "\n") {
			if line != "" {
				lines = append(lines, StyleMuted.Render(truncate(line, m.width-4)))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, StyleMuted.Render("enter: run | e: edit | j/k: navigate | esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
