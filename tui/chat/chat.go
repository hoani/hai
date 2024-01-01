package chat

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoani/hai/ai"
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
	userMessage string
	chat        []message
	client      *ai.Chat
	ready       bool
}

func New() tea.Model {
	return model{
		chat:   make([]message, 0),
		client: ai.NewChat(),
		ready:  true,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			// TODO: Send the message.
			userMessage := m.userMessage
			m.userMessage = ""
			m.ready = false
			return m, func() tea.Msg {
				return m.client.Update(userMessage)
			}

		default:
			if m.ready {
				m.userMessage += string(msg.Runes)
			}
		}

	case []openai.ChatCompletionMessage:
		m.chat = make([]message, 0, len(msg))
		for _, cmsg := range msg {
			if cmsg.Role == openai.ChatMessageRoleAssistant {
				m.chat = append(m.chat,
					message{
						source:  sourceAssistant,
						content: cmsg.Content,
					},
				)
			} else if cmsg.Role == openai.ChatMessageRoleUser {
				m.chat = append(m.chat,
					message{
						source:  sourceUser,
						content: cmsg.Content,
					},
				)
			}
		}
		m.ready = true
		return m, nil
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "How can I help you?\n\n"
	for _, m := range m.chat {
		if m.source == sourceUser {
			s += "> "
		}
		s += m.content + "\n"
	}
	if m.ready {
		s += "> " + m.userMessage
	}

	// The footer
	s += "\nPress ctrl-C to quit.\n"

	// Send the UI for rendering
	return s
}
