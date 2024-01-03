package tuiinit

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoani/hai/config"
)

type model struct {
	openaiKey    textinput.Model
	saveErr      error
	saveComplete bool
}

type saveError error

type saveComplete struct{}

type done struct{}

func New() tea.Model {
	ti := textinput.New()
	ti.Placeholder = "<key>"
	ti.Prompt = "OpenAI key: "
	ti.CharLimit = 0
	ti.Width = 80
	ti.Focus()

	return &model{
		openaiKey: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.openaiKey.Focused() {
				m.openaiKey.Blur()

				return m, func() tea.Msg {
					err := config.Save(config.WithOpenAIKey(m.openaiKey.Value()))
					if err != nil {
						return saveError(err)
					}
					return saveComplete{}
				}
			}
		}
	case saveError:
		m.saveErr = msg
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return done{} })

	case saveComplete:
		m.saveComplete = true
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return done{} })

	case done:
		return m, tea.Quit
	}

	m.openaiKey, cmd = m.openaiKey.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := ""
	s += m.openaiKey.View()
	s += "\n\n"
	if m.openaiKey.Focused() {
		s += "Press enter to continue \n"
	} else if m.saveErr != nil {
		s += fmt.Sprintf("Error saving file: %s \n", m.saveErr)
	} else if m.saveComplete {
		s += "Initialization complete \n"
	} else {
		s += "Saving... \n"
	}

	return s
}
