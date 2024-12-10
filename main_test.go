package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveFirstLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Multi-line input",
			input:    "First line\nSecond line\nThird line",
			expected: "Second line\nThird line",
		},
		{
			name:     "Anthropic Code",
			input:    "```go\nFirst line\nSecond line\nThird line",
			expected: "First line\nSecond line\nThird line",
		},
		{
			name:     "Single-line input",
			input:    "Single line",
			expected: "",
		},
		{
			name:     "Empty string input",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := removeFirstLine(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRemoveLineAndAfter(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		marker   string
		expected string
	}{
		{
			name:     "marker in middle of text",
			input:    "line1\nline2\n```\nline3\nline4",
			marker:   "```",
			expected: "line1\nline2",
		},
		{
			name:     "marker not found",
			input:    "line1\nline2\nline3",
			marker:   "```",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "multi-line input with marker",
			input:    "func main() {\n    print('hello')\n```\n    extra code",
			marker:   "```",
			expected: "func main() {\n    print('hello')",
		},
		{
			name:     "single-line input with marker",
			input:    "single line with ```",
			marker:   "```",
			expected: "single line with ```",
		},
		{
			name:     "empty string input",
			input:    "",
			marker:   "```",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := removeLineAndAfter(tc.input, tc.marker)
			assert.Equal(t, tc.expected, result)
		})
	}
}
