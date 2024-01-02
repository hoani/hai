package ai

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

type ChatStream interface {
	Recv() (response openai.ChatCompletionStreamResponse, err error)
	Close()
}

type ChatResponse string

type ChatDone struct{}

type ChatErr error

type Chat struct {
	*openai.Client
	req   openai.ChatCompletionRequest
	msgCh chan any
}

func NewChat() *Chat {
	return &Chat{
		Client: openai.NewClient(os.Getenv("OPENAI_KEY")),
		req: openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: make([]openai.ChatCompletionMessage, 0),
			Stream:   true,
		},
		msgCh: make(chan any),
	}
}

func (c *Chat) Send(userMessage string) error {
	c.appendMessage(
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	)

	stream, err := c.CreateChatCompletionStream(context.Background(), c.req)
	if err != nil {
		return errors.New("failed to retrieve chat response")
	}

	go c.handleChatStream(stream)

	return nil
}

func (c *Chat) handleChatStream(stream ChatStream) {
	defer stream.Close()
	msg := ""
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			c.msgCh <- ChatDone{}
			c.appendMessage(openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg,
			})
			return
		}
		if err != nil {
			c.msgCh <- ChatErr(err)
			return
		}
		if len(resp.Choices) == 0 {
			continue
		}
		c.msgCh <- ChatResponse(resp.Choices[0].Delta.Content)
		msg += resp.Choices[0].Delta.Content
	}
}

func (c *Chat) Recv() any {
	msg := <-c.msgCh
	return msg
}

func (c *Chat) appendMessage(msg openai.ChatCompletionMessage) {
	c.req.Messages = append(c.req.Messages, msg)
}
