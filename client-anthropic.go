package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type AnthropicClient struct {
	client *anthropic.Client
}

func NewAnthropicClient(key string) AIClient {
	return &AnthropicClient{
		client: anthropic.NewClient(key),
	}
}

// GetSuggestions returns suggestions
func (c *AnthropicClient) GetSuggestions(ctx context.Context, systemPrompt, usrPrompt string) ([]Suggestion, error) {
	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, fmt.Errorf("GenerateSchemaForType error: %w", err)
	}

	// Convert the schema object to a JSON string
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize schema: %w", err)
	}
	schemaStr := string(schemaBytes)

	resp, err := c.client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Dot5HaikuLatest,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(systemPrompt),
			anthropic.NewAssistantTextMessage(schemaStr),
			anthropic.NewUserTextMessage(usrPrompt),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	var res Result
	err = json.Unmarshal([]byte(resp.Content[0].GetText()), &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res.Suggestions, nil
}

// CreateTest creates a test
func (c *AnthropicClient) CreateTest(ctx context.Context, systemPrompt, usrPrompt string) (string, error) {

	resp, err := c.client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Dot5HaikuLatest,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(systemPrompt),
			anthropic.NewUserTextMessage(usrPrompt),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	return resp.Content[0].GetText(), nil
}
