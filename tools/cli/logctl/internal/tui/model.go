package tui

import (
	"fmt"
	"strings"
	"time"

	"fafnir/tools/logctl/internal/elastic"
	"fafnir/tools/logctl/internal/types"
	"fafnir/tools/logctl/internal/utils"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewForm view = iota
	viewLogs
	viewLoading
)

type Model struct {
	// view state
	currentView view
	width       int
	height      int

	// form inputs
	inputs       []textinput.Model
	focusedInput int

	// services list
	services     []string
	selectedSvc  int
	svcListFocus bool

	// follow mode toggle (if toggled it will start seeing the logs after the initial query)
	followMode bool

	// logs viewport
	viewport viewport.Model
	logs     []types.LogEntry

	// elasticsearch client
	client *elastic.Client

	// query state
	querying bool
	err      error

	// follow mode state
	following   bool
	lastQuery   time.Time
	queryTicker *time.Ticker
}

type queryCompleteMsg struct {
	logs []types.LogEntry
	err  error
}

type tickMsg time.Time

const (
	inputService = iota
	inputTimeRange
	inputSearch
	inputLimit
)

func initialModel() Model {
	// initialize text inputs
	inputs := make([]textinput.Model, 4)

	inputs[inputService] = textinput.New()
	inputs[inputService].Placeholder = "e.g., auth, user, stock, api-gateway"
	inputs[inputService].Focus()
	inputs[inputService].CharLimit = 50
	inputs[inputService].Width = 50

	inputs[inputTimeRange] = textinput.New()
	inputs[inputTimeRange].Placeholder = "e.g., 1h, 30m, 5m (default: 5m)"
	inputs[inputTimeRange].CharLimit = 20
	inputs[inputTimeRange].Width = 50

	inputs[inputSearch] = textinput.New()
	inputs[inputSearch].Placeholder = "keyword to search in logs"
	inputs[inputSearch].CharLimit = 100
	inputs[inputSearch].Width = 50

	inputs[inputLimit] = textinput.New()
	inputs[inputLimit].Placeholder = "100"
	inputs[inputLimit].CharLimit = 5
	inputs[inputLimit].Width = 50

	// initialize viewport
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		PaddingLeft(2).
		PaddingRight(2)

	// create Elasticsearch client
	client, _ := elastic.NewClient()

	return Model{
		currentView:  viewForm,
		inputs:       inputs,
		focusedInput: 0,
		services:     utils.ValidServicesList(),
		selectedSvc:  0,
		svcListFocus: false,

		followMode: false,
		viewport:   vp,
		logs:       []types.LogEntry{},
		client:     client,
		querying:   false,
		following:  false,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc": // when user presses ctrl+c or esc, quit the program
			if m.following && m.queryTicker != nil {
				m.queryTicker.Stop()
			}
			return m, tea.Quit

		case "tab": // when user presses tab, move focus to next input
			if m.currentView == viewForm {
				// cycle through inputs
				m.focusedInput++
				if m.focusedInput >= len(m.inputs) {
					m.focusedInput = 0
				}
				return m, m.updateInputFocus()
			}

		case "shift+tab": // when user presses shift+tab, move focus to previous input
			if m.currentView == viewForm {
				m.focusedInput--
				if m.focusedInput < 0 {
					m.focusedInput = len(m.inputs) - 1
				}
				return m, m.updateInputFocus()
			}

		case "enter": // when user presses enter, either submit query or go back to form
			if m.currentView == viewForm && !m.querying {
				if m.svcListFocus {
					m.inputs[inputService].SetValue(m.services[m.selectedSvc])
					m.svcListFocus = false
					return m, nil
				}

				// Submit query
				return m, m.submitQuery()
			} else if m.currentView == viewLogs {
				// Go back to form
				m.currentView = viewForm
				if m.following && m.queryTicker != nil {
					m.queryTicker.Stop()
					m.following = false
				}
				return m, nil
			}

		case "x": // when user presses x, toggle service list focus
			if m.currentView == viewForm && m.focusedInput == inputService {
				m.svcListFocus = !m.svcListFocus
				return m, nil
			}

		case "f": // when user presses f, toggle follow mode
			if m.currentView == viewForm {
				m.followMode = !m.followMode
				return m, nil
			}

		case "up", "k": // when user presses up or k, scroll up the viewport (for logs view)
			if m.currentView == viewLogs {
				m.viewport.SetYOffset(m.viewport.YOffset - 1)
			} else if m.svcListFocus {
				if m.selectedSvc > 0 {
					m.selectedSvc--
				}
			}

		case "down", "j": // when user presses down or j, scroll down the viewport (for logs view)
			if m.currentView == viewLogs {
				m.viewport.SetYOffset(m.viewport.YOffset + 1)
			} else if m.svcListFocus {
				if m.selectedSvc < len(m.services)-1 {
					m.selectedSvc++
				}
			}

		case "pgup": // when user presses pgup, scroll up the viewport (for logs view)
			if m.currentView == viewLogs {
				m.viewport.PageUp()
			}

		case "pgdown": // when user presses pgdown, scroll down the viewport (for logs view)
			if m.currentView == viewLogs {
				m.viewport.PageDown()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8

	case queryCompleteMsg:
		m.querying = false
		if msg.err != nil {
			m.err = msg.err
			m.currentView = viewForm
		} else {
			// for follow mode, append new logs; for regular queries, replace logs
			if m.following {
				m.logs = append(m.logs, msg.logs...)
			} else {
				m.logs = msg.logs
			}
			m.viewport.SetContent(m.renderLogs())
			m.viewport.GotoBottom() // always scroll to bottom as requested
			m.currentView = viewLogs
		}

	case tickMsg:
		if m.following {
			return m, tea.Batch(m.submitQuery(), m.waitForTick())
		}
	}

	// update active input
	if m.currentView == viewForm && !m.svcListFocus && !m.querying {
		cmd := m.updateInput(msg)
		commands = append(commands, cmd)
	}

	// update viewport
	if m.currentView == viewLogs {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		commands = append(commands, cmd)
	}

	return m, tea.Batch(commands...)
}

func (m Model) updateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return cmd
}

func (m Model) updateInputFocus() tea.Cmd {
	commands := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusedInput {
			commands[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(commands...)
}

func (m Model) submitQuery() tea.Cmd {
	m.querying = true
	m.currentView = viewLoading

	service := m.inputs[inputService].Value()
	if service == "" {
		service = m.services[m.selectedSvc]
	}

	timeRange := m.inputs[inputTimeRange].Value()
	since := time.Now().Add(-7 * 24 * time.Hour) // default to last 7 days
	if timeRange != "" {
		if parsedTime, err := utils.ParseTimeFlag(timeRange); err == nil {
			since = parsedTime
		}
	}

	search := m.inputs[inputSearch].Value()

	limit := 100
	if limitStr := m.inputs[inputLimit].Value(); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	opts := &types.QueryOptions{
		Service: service,
		Since:   since,
		Until:   time.Now().Add(24 * time.Hour), // handle timezone differences
		Search:  search,
		Limit:   limit,
	}

	// handle follow mode vs regular queries
	if m.followMode && !m.following {
		// starting follow mode - clear logs and start ticker
		m.logs = []types.LogEntry{}
		m.following = true
		m.queryTicker = time.NewTicker(2 * time.Second)
	} else if !m.followMode {
		// regular query - stop following if active and clear logs
		if m.following && m.queryTicker != nil {
			m.queryTicker.Stop()
			m.following = false
		}
		m.logs = []types.LogEntry{}
	}

	return func() tea.Msg {
		logs, err := m.client.QueryLogs(opts)
		return queryCompleteMsg{logs: logs, err: err}
	}
}

func (m Model) waitForTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) renderLogs() string {
	var sb strings.Builder

	for _, log := range m.logs {
		timestamp := log.Timestamp.Format("2006-01-02 15:04:05")

		sb.WriteString(fmt.Sprintf("[%s] %s [%s]\n",
			timestamp,
			log.Kubernetes.Container.Name,
			log.Message,
		))

		if log.RequestID != "" {
			sb.WriteString(fmt.Sprintf("  RequestID: %s\n", log.RequestID))
		}
		if log.Error != "" {
			sb.WriteString(fmt.Sprintf("  Error: %s\n", log.Error))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
