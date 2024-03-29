package tuichat

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hoani/hai/ai"
	"github.com/hoani/hai/config"
	"github.com/muesli/reflow/wordwrap"
)

const (
	inputHeight       = 3
	chatTopHeight     = 0
	chatBottomHeight  = 1
	inputTopHeight    = 1
	inputBottomHeight = 1
	nonChatHeight     = inputHeight + chatTopHeight + chatBottomHeight + inputTopHeight + inputBottomHeight
)

type model struct {
	ready    bool
	input    textarea.Model
	spinner  spinner.Model
	viewport viewport.Model
	response string
	content  string
	client   *ai.Chat
}

func New() (tea.Model, error) {
	ti := textarea.New()
	ti.ShowLineNumbers = false
	ti.CharLimit = 0
	ti.SetWidth(80)
	ti.SetHeight(inputHeight)
	ti.Placeholder = "How can I help today?"
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return model{
		input:   ti,
		client:  ai.NewChat(cfg.AI.OpenAI.Key),
		spinner: s,
	}, nil
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
			m.viewport = viewport.New(msg.Width, msg.Height-nonChatHeight)
			m.viewport.YPosition = 0
			m.viewport.HighPerformanceRendering = false // TODO: do we want this?
			m.viewport.SetContent("")
			m.viewport.KeyMap = ViewportKeymap()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - nonChatHeight
			m.viewport.GotoBottom()
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
				m.content += "\n" + wordwrap.String("> "+m.input.Value(), m.viewport.Width)
				m.viewport.SetContent(m.content)
				m.viewport.GotoBottom()
				m.input.Reset()
				m.input.Placeholder = ""
				m.response = ""

				return m, func() tea.Msg {
					if err := m.client.Send(userMessage); err != nil {
						return err
					}
					return m.client.Recv()
				}
			}
		}

	case ai.ChatResponse:
		m.response += string(msg)
		response := wordwrap.String(m.response, m.viewport.Width)
		response = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#ff22ff")).Render(response)
		m.viewport.SetContent(m.content + response)
		m.viewport.GotoBottom()
		return m, func() tea.Msg { return m.client.Recv() }

	case ai.ChatDone:
		response := wordwrap.String(m.response, m.viewport.Width)
		response = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd77ff")).Render(response)
		m.content += "\n" + response
		m.viewport.SetContent(m.content)
		m.viewport.GotoBottom()
		m.input.Focus()
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

	s := m.viewport.View() + "\n"

	chatBorder := "[ Up/Dn (shift-↑/↓) | ½Pg Up/Dn (ctrl-u/d) | Pg Up/Dn (pgup/pgdn) ] "

	if borderLen := m.viewport.Width - len([]rune(chatBorder)) - 2; borderLen > 0 {
		chatBorder = strings.Repeat(" ", borderLen) + chatBorder
	}

	chatBorder = lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render(chatBorder)

	if !m.input.Focused() {
		chatBorder = m.spinner.View() + chatBorder
	} else {
		chatBorder = "  " + chatBorder
	}

	s += chatBorder + "\n"

	s += strings.Repeat("=", m.viewport.Width) + "\n"

	s += m.input.View() + "\n"

	footer := "[ Send (enter) | Navigate (↑/↓/→/←) | Quit (ctrl-C) ]="
	if borderLen := m.viewport.Width - len([]rune(footer)); borderLen > 0 {
		footer = strings.Repeat("=", borderLen) + footer
	}
	s += footer + "\n"

	// Send the UI for rendering
	return s
}

func ViewportKeymap() viewport.KeyMap {
	return viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
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
