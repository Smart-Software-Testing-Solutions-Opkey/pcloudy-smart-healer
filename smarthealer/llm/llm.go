package llm

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v2/shared"
)

type MessageContentType int

const (
	TextContent MessageContentType = iota
	ImageContent
)

type MessageContent struct {
	Type   MessageContentType
	Data   string
	Detail string
}

type Role int

const (
	SystemRole Role = iota
	UserRole
)

type Message struct {
	Role    Role
	Content []MessageContent
}

var ErrLLMCompletionFailed = errors.New("failed to generate completion")

type LLM interface {
	Completion(ctx context.Context, messages []Message, model shared.ChatModel, jsonResponse bool) (string, error)
}
