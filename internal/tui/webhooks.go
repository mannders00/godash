package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WebhooksModel struct {
	webhooks []WebhookItem
	cursor   int
	width    int
	height   int
	detail   bool
	response string
}

type WebhookItem struct {
	Name   string
	URL    string
	Method string
	Status string
	Labels map[string]string
}

func NewWebhooksModel() WebhooksModel {
	return WebhooksModel{
		webhooks: []WebhookItem{
			{Name: "deploy-prod", URL: "https://api.github.com/repos/user/repo/dispatches", Method: "POST", Status: "ready", Labels: map[string]string{"type": "ci"}},
			{Name: "notify-slack", URL: "https://hooks.slack.com/services/T00/B00/XXX", Method: "POST", Status: "ready", Labels: map[string]string{"type": "notify"}},
			{Name: "stripe-webhook", URL: "https://api.stripe.com/v1/webhook_endpoints", Method: "POST", Status: "ready", Labels: map[string]string{"type": "payment"}},
		},
	}
}

func (m WebhooksModel) Items() []WebhookItem {
	return m.webhooks
}

func (m WebhooksModel) Init() tea.Cmd {
	return nil
}

func (m WebhooksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.detail {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m *WebhooksModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.webhooks)-1 {
			m.cursor++
		}
	case "enter":
		m.detail = true
		return m, nil
	case "t":
		if m.cursor < len(m.webhooks) {
			return m, triggerWebhook(m.webhooks[m.cursor], m.cursor)
		}
	}
	return m, nil
}

func (m *WebhooksModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.detail = false
		m.response = ""
	case "t":
		if m.cursor < len(m.webhooks) {
			return m, triggerWebhook(m.webhooks[m.cursor], m.cursor)
		}
	}
	return m, nil
}

func (m WebhooksModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	if m.detail && m.cursor < len(m.webhooks) {
		return m.renderDetail()
	}

	var lines []string
	lines = append(lines, titleStyle.Render(" Webhooks "))
	lines = append(lines, "")
	lines = append(lines, headerStyle.Render(fmt.Sprintf("%-14s %-8s %-8s %s", "NAME", "METHOD", "STATUS", "URL")))
	lines = append(lines, strings.Repeat("─", m.width-4))

	for i, wh := range m.webhooks {
		statusColor := lipgloss.Color("2")
		status := "ready"
		if wh.Status == "running" {
			statusColor = lipgloss.Color("3")
			status = "running"
		}

		line := fmt.Sprintf("%-14s %-8s %-8s %s",
			truncate(wh.Name, 14),
			wh.Method,
			lipgloss.NewStyle().Foreground(statusColor).Render(status),
			truncate(wh.URL, m.width-40),
		)

		if i == m.cursor {
			lines = append(lines, selectedStyle.Render("▶ "+line))
		} else {
			lines = append(lines, "   "+line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" j/k: navigate │ enter: details │ t: trigger │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m WebhooksModel) renderDetail() string {
	wh := m.webhooks[m.cursor]
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	statusReady := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	statusRunning := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf(" Webhook: %s ", wh.Name)))
	lines = append(lines, "")

	var status string
	if wh.Status == "running" {
		status = statusRunning.Render("● RUNNING")
	} else {
		status = statusReady.Render("● READY")
	}
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Status:"), status))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Method:"), valueStyle.Render(wh.Method)))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("URL:"), valueStyle.Render(wh.URL)))

	if len(wh.Labels) > 0 {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Labels:"))
		for k, v := range wh.Labels {
			lines = append(lines, fmt.Sprintf("  %s=%s", k, v))
		}
	}

	if m.response != "" {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Response:"))
		for _, line := range strings.Split(m.response, "\n") {
			if line != "" {
				lines = append(lines, "  "+truncate(line, m.width-6))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", m.width-4))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(" t: trigger │ esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type webhookResultMsg struct {
	index    int
	response string
	err      error
}

func triggerWebhook(wh WebhookItem, index int) tea.Cmd {
	return func() tea.Msg {
		_ = wh
		_ = index
		return webhookResultMsg{index: index, response: "Webhook triggered (placeholder)", err: nil}
	}
}
