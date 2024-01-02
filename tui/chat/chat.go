package chat

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hoani/hai/ai"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sashabaranov/go-openai"
)

type source int

const (
	sourceUser source = iota
	sourceAssistant
)

type message struct {
	source  source
	content string
}

type model struct {
	ready    bool
	input    textarea.Model
	spinner  spinner.Model
	viewport viewport.Model
	content  string
	chat     []message
	client   *ai.Chat
}

func New() tea.Model {
	ti := textarea.New()
	ti.ShowLineNumbers = false
	ti.CharLimit = 0
	ti.SetWidth(80)
	ti.SetHeight(3)
	ti.Placeholder = "How can I help today?"
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	return model{
		input:   ti,
		chat:    make([]message, 0),
		client:  ai.NewChat(),
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.input.SetWidth(msg.Width)

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-5)
			m.viewport.YPosition = 0
			m.viewport.HighPerformanceRendering = false // TODO: do we want this?
			m.viewport.SetContent("")
			m.viewport.KeyMap = ViewportKeymap()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 5
		}

		if m.viewport.Height < 0 {
			m.viewport.Height = 0 // TODO: might want to display some warning in this case?
		}

	// Handle non-input keys.
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.input.Focused() {
				m.input.Blur()
				userMessage := m.input.Value()
				m.content += wordwrap.String("> "+m.input.Value()+"\n", m.viewport.Width)
				m.viewport.SetContent(m.content)
				m.viewport.GotoBottom()
				m.input.Reset()
				m.input.Placeholder = ""

				return m, func() tea.Msg {
					result, err := m.client.Update(userMessage)
					if err != nil {
						return err
					}
					return result
				}
			}
		}

	case openai.ChatCompletionMessage:

		if msg.Role == openai.ChatMessageRoleAssistant {
			m.content += wordwrap.String(msg.Content+"\n", m.viewport.Width)
		} else {
			m.content += fmt.Sprintf("unexpected chat message from %s:", msg.Role)
		}
		m.input.Focus()

		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()

		return m, nil

	case error:
		m.content += msg.Error() + "\n"
		// Don't clear the input so that the user can try again.
		m.input.Focus()
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := ""
	s += m.viewport.View()
	s += "\n"

	s += m.input.View()

	s += "\n"

	if !m.input.Focused() {
		s += m.spinner.View()
	}

	// The footer
	s += " Press ctrl-C to quit.\n"

	// Send the UI for rendering
	return s
}

func ViewportKeymap() viewport.KeyMap {
	return viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("f/pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("shift+up"),
			key.WithHelp("shift+↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("shift+down"),
			key.WithHelp("shift+↓", "down"),
		),
	}
}
