package tui

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MachinesModel struct {
	machines []MachineItem
	cursor   int
	width    int
	height   int
	detail   bool
}

type MachineItem struct {
	Name    string
	Host    string
	User    string
	Port    int
	KeyPath string
	Status  string
	Latency time.Duration
	Uptime  string
	Labels  map[string]string
}

func NewMachinesModel() MachinesModel {
	return MachinesModel{
		machines: []MachineItem{
			{Name: "prod-doji", Host: "147.182.123.45", User: "root", Port: 22, Status: "unknown", Labels: map[string]string{"env": "prod", "provider": "do"}},
			{Name: "staging", Host: "192.168.1.101", User: "ubuntu", Port: 22, Status: "unknown", Labels: map[string]string{"env": "staging"}},
			{Name: "hetzner-1", Host: "95.217.64.200", User: "root", Port: 22, Status: "unknown", Labels: map[string]string{"env": "prod", "provider": "hetzner"}},
		},
	}
}

func (m MachinesModel) Items() []MachineItem {
	return m.machines
}

func (m MachinesModel) Init() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func checkMachines(machines []MachineItem) tea.Cmd {
	return func() tea.Msg {
		for i := range machines {
			start := time.Now()
			addr := fmt.Sprintf("%s:%d", machines[i].Host, machines[i].Port)
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
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
}

type machinesCheckedMsg []MachineItem

func (m *MachinesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, checkMachines(m.machines)
	case machinesCheckedMsg:
		m.machines = msg
		return m, tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyMsg:
		if m.detail {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m *MachinesModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.machines)-1 {
			m.cursor++
		}
	case "enter":
		m.detail = true
		return m, nil
	case "s":
		if m.cursor < len(m.machines) {
			return m, m.SSH(m.machines[m.cursor])
		}
	}
	return m, nil
}

func (m *MachinesModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.detail = false
	}
	return m, nil
}

func (m *MachinesModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	if m.detail && m.cursor < len(m.machines) {
		return m.renderDetail()
	}

	var lines []string
	lines = append(lines, titleStyle.Render(" Machines "))
	lines = append(lines, "")
	lines = append(lines, headerStyle.Render(fmt.Sprintf("%-14s %-10s %-15s %8s", "NAME", "STATUS", "HOST", "LATENCY")))
	lines = append(lines, strings.Repeat("─", m.width-4))

	for i, machine := range m.machines {
		var statusIcon string
		var statusColor lipgloss.Color
		switch machine.Status {
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

		latency := "-"
		if machine.Latency > 0 {
			latency = machine.Latency.Round(time.Millisecond).String()
		}

		line := fmt.Sprintf("%-14s %s %-10s %-15s %8s",
			truncate(machine.Name, 14),
			lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
			machine.Status,
			machine.Host,
			latency,
		)

		if i == m.cursor {
			lines = append(lines, selectedStyle.Render("▶ "+line))
		} else {
			lines = append(lines, "   "+line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" j/k: navigate │ enter: details │ s: SSH │ r: refresh │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m MachinesModel) renderDetail() string {
	machine := m.machines[m.cursor]
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	statusOnline := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	statusOffline := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf(" Machine: %s ", machine.Name)))
	lines = append(lines, "")

	var status string
	if machine.Status == "online" {
		status = statusOnline.Render("● ONLINE")
	} else {
		status = statusOffline.Render("○ OFFLINE")
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Status:"), status))

	latency := "-"
	if machine.Latency > 0 {
		latency = machine.Latency.Round(time.Millisecond).String()
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Latency:"), valueStyle.Render(latency)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Host:"), valueStyle.Render(machine.Host)))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Port:"), valueStyle.Render(fmt.Sprintf("%d", machine.Port))))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("User:"), valueStyle.Render(machine.User)))
	if machine.KeyPath != "" {
		lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Key:"), valueStyle.Render(machine.KeyPath)))
	}

	if len(machine.Labels) > 0 {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Labels:"))
		for k, v := range machine.Labels {
			lines = append(lines, fmt.Sprintf("  %s=%s", k, v))
		}
	}

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", m.width-4))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" s: SSH connect │ r: refresh │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m MachinesModel) SSH(machine MachineItem) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		_ = ctx
		return nil
	}
}
