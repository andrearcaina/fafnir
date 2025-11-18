package tui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	switch m.currentView {
	case viewForm:
		return m.renderForm()
	case viewLoading:
		return m.renderLoading()
	case viewLogs:
		return m.renderLogsView()
	default:
		return ""
	}
}

func (m Model) renderForm() string {
	var sb strings.Builder

	// title
	sb.WriteString(TitleStyle().Render("🔍 logctl - Fafnir Elasticsearch Log Viewer"))
	sb.WriteString("\n\n")

	// service input/list
	sb.WriteString(LabelStyle().Render("Service:"))
	sb.WriteString("\n")

	if m.svcListFocus {
		sb.WriteString(m.renderServiceList())
	} else {
		style := BlurredStyle()
		if m.focusedInput == inputService {
			style = FocusedStyle()
		}
		sb.WriteString(style.Render(m.inputs[inputService].View()))
		sb.WriteString(" ")
		sb.WriteString(HelpStyle().Render("(press 'x' to select from list)"))
	}
	sb.WriteString("\n\n")

	// time range input
	sb.WriteString(LabelStyle().Render("Time Range:"))
	sb.WriteString("\n")
	style := BlurredStyle()
	if m.focusedInput == inputTimeRange {
		style = FocusedStyle()
	}
	sb.WriteString(style.Render(m.inputs[inputTimeRange].View()))
	sb.WriteString("\n\n")

	// search input
	sb.WriteString(LabelStyle().Render("Search:"))
	sb.WriteString("\n")
	style = BlurredStyle()
	if m.focusedInput == inputSearch {
		style = FocusedStyle()
	}
	sb.WriteString(style.Render(m.inputs[inputSearch].View()))
	sb.WriteString("\n\n")

	// limit input
	sb.WriteString(LabelStyle().Render("Limit:"))
	sb.WriteString("\n")
	style = BlurredStyle()
	if m.focusedInput == inputLimit {
		style = FocusedStyle()
	}
	sb.WriteString(style.Render(m.inputs[inputLimit].View()))
	sb.WriteString("\n\n")

	// follow mode toggle
	sb.WriteString(LabelStyle().Render("Follow Mode:"))
	sb.WriteString(" ")
	if m.followMode {
		sb.WriteString(SelectedItemStyle().Render("✓ ENABLED"))
	} else {
		sb.WriteString("✗ DISABLED")
	}
	sb.WriteString(" ")
	sb.WriteString(HelpStyle().Render("(press 'f' to toggle)"))
	sb.WriteString("\n\n")

	// error display
	if m.err != nil {
		sb.WriteString(ErrorStyle().Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
	}

	// help text
	sb.WriteString(HelpStyle().Render("tab/shift+tab: navigate • enter: submit query • esc: quit"))

	return sb.String()
}

func (m Model) renderServiceList() string {
	var items []string
	for i, svc := range m.services {
		if i == m.selectedSvc {
			items = append(items, SelectedItemStyle().Render("▸ "+svc))
		} else {
			items = append(items, "  "+svc)
		}
	}
	content := strings.Join(items, "\n")
	return ListStyle().Render(content)
}

func (m Model) renderLoading() string {
	var sb strings.Builder

	sb.WriteString(TitleStyle().Render("🔍 logctl - Fafnir Elasticsearch Log Viewer"))
	sb.WriteString("\n\n")

	sb.WriteString(LoadingStyle().Render("⏳ Querying Elasticsearch..."))
	sb.WriteString("\n\n")

	sb.WriteString(HelpStyle().Render("Please wait while we fetch your logs"))

	return sb.String()
}

func (m Model) renderLogsView() string {
	var sb strings.Builder

	// header
	header := TitleStyle().Render("🔍 logctl - Fafnir Elasticsearch Log Viewer")
	sb.WriteString(header)
	sb.WriteString("\n")

	// status bar
	status := fmt.Sprintf("Showing %d logs", len(m.logs))
	if m.following {
		status += " • FOLLOWING (updating every 2s)"
	}
	sb.WriteString(StatusStyle().Render(status))
	sb.WriteString("\n\n")

	// viewport with logs
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	// help text
	help := "↑/↓: scroll • pgup/pgdown: page • enter: back to form • esc: quit"
	sb.WriteString(HelpStyle().Render(help))

	return sb.String()
}
