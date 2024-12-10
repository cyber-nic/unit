package main

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(key string) AIClient {
	return &OpenAIClient{
		client: openai.NewClient(key),
	}
}

func (c *OpenAIClient) Name() string {
	return providerOpenAI
}

// GetSuggestions returns suggestions
func (c *OpenAIClient) GetSuggestions(ctx context.Context, systemPrompt, usrPrompt string) ([]Suggestion, error) {
	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, fmt.Errorf("GenerateSchemaForType error: %w", err)
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: usrPrompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "unit_test_suggestions",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	var res Result
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res.Suggestions, nil
}

// CreateTest creates a test
func (c *OpenAIClient) CreateTest(ctx context.Context, systemPrompt, usrPrompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: usrPrompt,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}
