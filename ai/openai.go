package ai

import (
	"context"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Chat struct {
	*openai.Client
	req openai.ChatCompletionRequest
}

func NewChat() *Chat {
	return &Chat{
		Client: openai.NewClient(os.Getenv("OPENAI_KEY")),
		req: openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: make([]openai.ChatCompletionMessage, 0),
		},
	}
}

func (c *Chat) Update(userMessage string) []openai.ChatCompletionMessage {
	c.appendMessage(
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	)

	resp, err := c.CreateChatCompletion(context.Background(), c.req)
	if err != nil {
		// TODO: handle this better.
		return c.req.Messages
	}
	c.appendMessage(resp.Choices[0].Message)

	return c.req.Messages
}

func (c *Chat) appendMessage(msg openai.ChatCompletionMessage) {
	c.req.Messages = append(c.req.Messages, msg)
}
