package tui

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WebhooksModel struct {
	webhooks []WebhookItem
	cursor   int
	width    int
	height   int
	response string
	running  int
}

type WebhookItem struct {
	Name   string
	URL    string
	Method string
	Status string
}

func NewWebhooksModel() WebhooksModel {
	return WebhooksModel{
		webhooks: []WebhookItem{
			{Name: "deploy-hook", URL: "https://api.example.com/deploy", Method: "POST", Status: "ready"},
			{Name: "notify-slack", URL: "https://hooks.slack.com/services/...", Method: "POST", Status: "ready"},
		},
	}
}

func (m WebhooksModel) Init() tea.Cmd {
	return nil
}

type webhookResultMsg struct {
	index    int
	response string
	err      error
}

func triggerWebhook(wh WebhookItem, index int) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 10 * time.Second}
		var resp *http.Response
		var err error

		switch wh.Method {
		case "GET":
			resp, err = client.Get(wh.URL)
		case "POST":
			resp, err = client.Post(wh.URL, "application/json", bytes.NewBuffer([]byte{}))
		default:
			resp, err = client.Get(wh.URL)
		}

		if err != nil {
			return webhookResultMsg{index: index, response: "", err: err}
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		return webhookResultMsg{index: index, response: string(body), err: nil}
	}
}

func (m WebhooksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case webhookResultMsg:
		if msg.index < len(m.webhooks) {
			m.webhooks[msg.index].Status = "ready"
		}
		if msg.err != nil {
			m.response = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.response = msg.response
		}
		return m, nil

	case tea.KeyMsg:
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
			if m.cursor < len(m.webhooks) {
				m.webhooks[m.cursor].Status = "running"
				m.response = ""
				return m, triggerWebhook(m.webhooks[m.cursor], m.cursor)
			}
		}
	}
	return m, nil
}

func (m WebhooksModel) View() string {
	var lines []string
	lines = append(lines, StyleTitle.Render("Webhooks"))
	lines = append(lines, "")
	lines = append(lines, StyleHeader.Render(fmt.Sprintf("%-20s %-10s %-10s %s", "NAME", "METHOD", "STATUS", "URL")))
	lines = append(lines, StyleBorder.Render(string(make([]byte, m.width-2))))

	for i, wh := range m.webhooks {
		status := wh.Status
		switch status {
		case "running":
			status = StyleWarning.Render("running")
		case "ready":
			status = StyleSuccess.Render("ready")
		default:
			status = StyleMuted.Render(status)
		}

		line := fmt.Sprintf("%-20s %-10s %-10s %s",
			wh.Name,
			wh.Method,
			status,
			truncate(wh.URL, m.width-50),
		)
		if i == m.cursor {
			lines = append(lines, StyleSelected.Render("> "+line))
		} else {
			lines = append(lines, "  "+line)
		}
	}

	if m.response != "" {
		lines = append(lines, "")
		lines = append(lines, StyleHeader.Render("Response:"))
		lines = append(lines, StyleMuted.Render(truncate(m.response, m.width-4)))
	}

	lines = append(lines, "")
	lines = append(lines, StyleMuted.Render("enter: trigger | j/k: navigate | esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen && maxLen > 0 {
		return s[:maxLen-3] + "..."
	}
	return s
}
