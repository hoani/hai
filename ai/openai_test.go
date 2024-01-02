package ai

import (
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

func TestHandleChatStream(t *testing.T) {
	c := NewChat()

	ts := newTestChatStream(
		withChatStreamDelta("Hello W"),
		withChatStreamDelta("orld!"),
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.handleChatStream(ts)
	}()

	result := c.Recv()
	require.Equal(t, ChatResponse("Hello W"), result)

	result = c.Recv()
	require.Equal(t, ChatResponse("orld!"), result)

	result = c.Recv()
	require.Equal(t, ChatDone{}, result)

	wg.Wait() // Wait for handleChatStream to finish.

	require.True(t, ts.closed)

	require.Equal(t, "Hello World!", c.req.Messages[0].Content)
}

func TestHandleChatStreamWithError(t *testing.T) {
	c := NewChat()

	ts := newTestChatStream(
		withChatStreamDelta("Hello W"),
		withChatStreamError(errors.New("failed")),
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.handleChatStream(ts)
	}()

	result := c.Recv()
	require.Equal(t, ChatResponse("Hello W"), result)

	result = c.Recv()
	require.Equal(t, ChatErr(errors.New("failed")), result)

	wg.Wait() // Wait for handleChatStream to finish.

	require.True(t, ts.closed)
}

// Helpers.

type testChatStream struct {
	closed    bool
	responses []any
}

type testChatStreamOption func(*testChatStream)

func newTestChatStream(opts ...testChatStreamOption) *testChatStream {
	s := &testChatStream{
		responses: make([]any, 0),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

func withChatStreamDelta(delta string) testChatStreamOption {
	return func(s *testChatStream) {
		s.responses = append(
			s.responses,
			openai.ChatCompletionStreamResponse{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Role:    openai.ChatMessageRoleAssistant,
							Content: delta,
						},
					},
				},
			},
		)
	}
}

func withChatStreamError(err error) testChatStreamOption {
	return func(s *testChatStream) {
		s.responses = append(s.responses, err)
	}
}

func (s *testChatStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if len(s.responses) == 0 || s.closed {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	resp := s.responses[0]
	s.responses = s.responses[1:]
	if err, ok := resp.(error); ok {
		return openai.ChatCompletionStreamResponse{}, err
	}
	return resp.(openai.ChatCompletionStreamResponse), nil
}

func (s *testChatStream) Close() {
	s.closed = true
}
