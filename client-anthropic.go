package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

func (c *AnthropicClient) Name() string {
	return providerAnthropic
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
		MaxTokens: 8092,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	fmt.Printf(resp.Content[0].GetText())

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
		MaxTokens: 8092,
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	// trim prefix "```go"
	code := removeFirstLine(resp.Content[0].GetText())

	// remove suffix "```"
	code = removeLineAndAfter(code, "```")

	return code, nil
}

func removeFirstLine(input string) string {
	// Split the string into two parts at the first newline
	parts := strings.SplitN(input, "\n", 2)

	// If there's only one part, return an empty string
	if len(parts) < 2 {
		return ""
	}

	// Return the second part (everything after the first newline)
	return parts[1]
}

func removeLineAndAfter(input, marker string) string {
	// Split the string into lines
	lines := strings.Split(input, "\n")

	// Find the line containing the marker and truncate
	for i, line := range lines {
		if line == marker {
			// Return everything before the marker line
			return strings.Join(lines[:i], "\n")
		}
	}

	// If marker is not found, return the original string
	return input
}
