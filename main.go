package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var openaiKey = ""

func init() {
	// check OPENAI_API_KEY environment variable

	if key, exists := os.LookupEnv("OPENAI_API_KEY"); exists {
		openaiKey = key
		return
	}

	if key, err := os.ReadFile("./secrets/openai_api_key"); err != nil {
		openaiKey = string(key)
	}

	if openaiKey == "" {
		log.Fatalf("OPENAI_API_KEY environment variable is not set")
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <file-path>", os.Args[0])
	}

	filePath := os.Args[1]
	fmt.Println("Reading file:", filePath)
	err := listUnitTests(filePath)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func listUnitTests(filePath string) error {
	// Step 1: Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Step 2: Set up OpenAI client
	client := openai.NewClient(openaiKey)

	ctx := context.Background()

	type Result struct {
		Suggestions []struct {
			Title   string   `json:"title"`
			Reasons []string `json:"reasons"`
		} `json:"suggestions"`
	}

	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}

	// Step 3: Make the OpenAI API call
	prompt := fmt.Sprintf("Analyze the following Go code and list possible unit tests that could be generated:\n\n%s", content)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant that generates unit test suggestions.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "unit_test_list",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	var res Result
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &res)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Step 5: Print the suggestions with index
	fmt.Println("Suggested Unit Tests:")
	for i, suggestion := range res.Suggestions {
		fmt.Printf("%d. %s\n", i+1, suggestion.Title)
		for _, reason := range suggestion.Reasons {
			fmt.Printf("   - %s\n", reason)
		}
	}

	// Step 6: Wait for user input
	fmt.Println("\nPress Enter to exit...")
	_, _ = fmt.Scanln()

	return nil
}
