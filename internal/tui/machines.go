package tui

import (
	"context"
	"fmt"
	"net"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MachinesModel struct {
	machines []MachineItem
	cursor   int
	width    int
	height   int
	checking bool
}

type MachineItem struct {
	Name    string
	Host    string
	Port    int
	Status  string
	Latency time.Duration
}

type tickMsg time.Time

func NewMachinesModel() MachinesModel {
	return MachinesModel{
		machines: []MachineItem{
			{Name: "prod-server", Host: "192.168.1.100", Port: 22, Status: "unknown"},
			{Name: "staging", Host: "192.168.1.101", Port: 22, Status: "unknown"},
		},
	}
}

func (m MachinesModel) Init() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func checkMachines(machines []MachineItem) tea.Msg {
	for i := range machines {
		start := time.Now()
		timeout := time.Second * 2
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", machines[i].Host, machines[i].Port), timeout)
		if err != nil {
			machines[i].Status = "offline"
			machines[i].Latency = 0
		} else {
			conn.Close()
			machines[i].Status = "online"
			machines[i].Latency = time.Since(start)
		}
	}
	return machinesCheckedMsg(machines)
}

type machinesCheckedMsg []MachineItem

func (m MachinesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, func() tea.Msg {
			return checkMachines(m.machines)
		}
	case machinesCheckedMsg:
		m.machines = msg
		m.checking = false
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.machines)-1 {
				m.cursor++
			}
		case "r":
			m.checking = true
			return m, func() tea.Msg {
				return checkMachines(m.machines)
			}
		}
	}
	return m, nil
}

func (m MachinesModel) View() string {
	var lines []string
	lines = append(lines, StyleTitle.Render("Machines"))
	lines = append(lines, "")
	lines = append(lines, StyleHeader.Render(fmt.Sprintf("%-20s %-20s %-10s %s", "NAME", "HOST", "STATUS", "LATENCY")))
	lines = append(lines, StyleBorder.Render(string(make([]byte, m.width-2))))

	for i, machine := range m.machines {
		line := fmt.Sprintf("%-20s %-20s %-10s %s",
			machine.Name,
			fmt.Sprintf("%s:%d", machine.Host, machine.Port),
			machine.Status,
			machine.Latency,
		)
		if i == m.cursor {
			lines = append(lines, StyleSelected.Render("> "+line))
		} else {
			var styled string
			switch machine.Status {
			case "online":
				styled = StyleSuccess.Render("● ") + line[2:]
			case "offline":
				styled = StyleError.Render("○ ") + line[2:]
			default:
				styled = StyleMuted.Render("· ") + line[2:]
			}
			lines = append(lines, styled)
		}
	}

	lines = append(lines, "")
	lines = append(lines, StyleMuted.Render("r: refresh | j/k: navigate | enter: connect | esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m MachinesModel) SSH(machine MachineItem) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		_ = ctx
		return nil
	}
}
