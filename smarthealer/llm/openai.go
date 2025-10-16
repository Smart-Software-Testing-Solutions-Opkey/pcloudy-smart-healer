package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
)

type openAILLM struct {
	key    string
	client openai.Client
}

func NewOpenAILLM(apiKey string) *openAILLM {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithEnvironmentProduction(),
	)

	return &openAILLM{
		key:    apiKey,
		client: client,
	}
}

func (llm *openAILLM) Completion(ctx context.Context, messages []Message, model shared.ChatModel, jsonResponse bool) (string, error) {
	var completionMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range messages {
		m, err := generateCompletionMessageParam(msg)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrLLMCompletionFailed, err)
		}

		completionMessages = append(completionMessages, m)
	}

	var r openai.ChatCompletionNewParamsResponseFormatUnion
	if jsonResponse {
		t := shared.NewResponseFormatJSONObjectParam()
		r = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &t,
		}
	} else {
		t := shared.NewResponseFormatTextParam()
		r = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &t,
		}
	}

	completion, err := llm.client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Messages:       completionMessages,
			Model:          model,
			ResponseFormat: r,
		},
	)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrLLMCompletionFailed, err)
	}

	res, err := getCompletionMessage(completion)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrLLMCompletionFailed, err)
	}

	return res, nil
}

func generateCompletionMessageParam(message Message) (openai.ChatCompletionMessageParamUnion, error) {

	switch message.Role {
	case SystemRole:
		return openai.SystemMessage(generateTextCompletionContent(message.Content)), nil
	case UserRole:

		return openai.UserMessage(generateChatCompletionContent(message.Content)), nil
	default:
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("invalid role specified in message")
	}
}

func generateTextCompletionContent(contents []MessageContent) []openai.ChatCompletionContentPartTextParam {
	var texts []openai.ChatCompletionContentPartTextParam

	for _, content := range contents {
		switch content.Type {
		case TextContent:
			text := openai.ChatCompletionContentPartTextParam{
				Text: content.Data,
			}
			texts = append(texts, text)
		case ImageContent:
			//! warn user of error
			continue
		default:
			//! warn user of error
			continue
		}
	}
	return texts
}

func generateChatCompletionContent(contents []MessageContent) []openai.ChatCompletionContentPartUnionParam {
	var params []openai.ChatCompletionContentPartUnionParam

	for _, content := range contents {
		switch content.Type {
		case TextContent:
			p := openai.TextContentPart(content.Data)
			params = append(params, p)
		case ImageContent:
			p := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL:    content.Data,
				Detail: content.Detail,
			})
			params = append(params, p)
		default:
			//! warn user of error
			continue
		}
	}
	return params
}

func getCompletionMessage(completion *openai.ChatCompletion) (string, error) {
	if len(completion.Choices) < 1 {
		return "", fmt.Errorf("failed to generate chat completion")
	}

	choice := completion.Choices[0]
	switch choice.FinishReason {
	case "length":
		return "", fmt.Errorf("max tokens for request reached: %s", choice.Message.Refusal)
	case "content_filter":
		return "", fmt.Errorf("content filtered by openai: %s", choice.Message.Refusal)
	}

	return choice.Message.Content, nil
}
