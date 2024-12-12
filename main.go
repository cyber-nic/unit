package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfg = struct {
	AIProvider   string
	AISecretPath string
	Color        bool
	Debug        bool
	Write        bool
}{}

const (
	providerOpenAI    = "openai"
	providerAnthropic = "anthropic"
)

// colorizer
var au *aurora.Aurora

type AIClient interface {
	Name() string
	GetSuggestions(ctx context.Context, systemPrompt, usrPrompt string) ([]Suggestion, error)
	CreateTest(ctx context.Context, systemPrompt, usrPrompt string) (string, error)
}

var aiClient AIClient

var errNotCached = errors.New("not cached")

type Suggestion struct {
	Title   string   `json:"title"`
	Reasons []string `json:"reasons"`
}

type Result struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func init() {
	viper.SetConfigName(".unit")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	// optionally read config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// disappointing: vipe not failing if invalid file content eg. json in .yaml
			panic(fmt.Errorf("fatal error config file: %s \n", err))
		}
	}

	// handle config flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] path\n", os.Args[0])
		flag.PrintDefaults()
	}

	provider := "anthropic"
	if viper.IsSet("provider") {
		provider = viper.GetString("provider")
	}
	flag.String("provider", provider, "AI Provider")

	secretPath := ".provider_api_key"
	if viper.IsSet("secret_path") {
		secretPath = viper.GetString("secret_path")
	}

	flag.String("secret_path", secretPath, "AI Secret Path")
	flag.Bool("color", true, "Toggle color")
	flag.Bool("debug", false, "Toggle debug")
	flag.Bool("write", false, "Toggle write file")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	cfg.AIProvider = viper.GetString("provider")
	cfg.AISecretPath = viper.GetString("secret_path")
	cfg.Color = viper.GetBool("color")
	cfg.Debug = viper.GetBool("debug")
	cfg.Write = viper.GetBool("write")

	au = aurora.New(aurora.WithColors(cfg.Color))

	var aiKey string

	// Recommended: check file
	if key, err := os.ReadFile(cfg.AISecretPath); err == nil {
		aiKey = string(key)
	}

	// Optional: check environment variable
	if aiKey == "" {
		if key, exists := os.LookupEnv("AI_API_KEY"); exists {
			aiKey = key
		}
	}

	// Required: exit if no API key is found
	if aiKey == "" {
		log.Fatalf("api key is required")
	}

	if cfg.AIProvider == providerOpenAI {
		aiClient = NewOpenAIClient(strings.Trim(aiKey, "\n"))
	} else if cfg.AIProvider == providerAnthropic {
		aiClient = NewAnthropicClient(strings.Trim(aiKey, "\n"))
	} else {
		log.Fatalf("invalid provider: %s", cfg.AIProvider)
	}
}

func main() {
	logger := zap.Must(zap.NewProduction())
	if cfg.Debug {
		logger = zap.Must(zap.NewDevelopment())
	}
	defer logger.Sync()

	// check for fs path
	var path string
	for i := 1; i < len(os.Args); i++ {
		// validate argument is fs path
		if _, err := os.Stat(os.Args[i]); err == nil {
			path = os.Args[i]
			break
		}
	}

	// fail if no path is found
	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read file: %w", err))
	}

	// Compute the SHA256 hash of the file
	h := sha256.New()
	if _, err := io.Copy(h, bytes.NewReader(content)); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to compute hash: %v", err))
	}
	logger.Debug("Input",
		zap.String("file", path),
		zap.String("hash", fmt.Sprintf("%x", h.Sum(nil))),
		zap.String("provider", cfg.AIProvider),
		zap.Int("length", len(content)),
	)

	var suggestions []Suggestion

	suggestions, err = readCachedSuggestions(h)
	if err != nil {
		if err == errNotCached {
			suggestions, err = getSuggestedUnitTests(ctx, content)
			if err != nil {
				logger.Fatal(fmt.Sprintf("Failed to get suggestions: %v", err))
			}

			// Write the suggestions to a cache file
			go func() {
				if err := writeCachedSuggestions(h, suggestions); err != nil {
					logger.Fatal(fmt.Sprintf("Failed to write cached suggestions: %v", err))
				}
			}()
		} else {
			logger.Fatal(fmt.Sprintf("Failed to read cached suggestions: %v", err))
		}
	}

	// user selects a unit test
	suggestion, err := selectSuggestedUnitTest(suggestions)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to select unit test: %v", err))
	}

	unitTest, err := createUnitTest(ctx, content, suggestion)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create unit test: %v", err))
	}

	if cfg.Write {
		out := fmt.Sprintf("%s/%s_test.go", filepath.Dir(path), "unit")
		// create dir
		if _, err := os.Stat(out); os.IsNotExist(err) {
			err := os.WriteFile(out, []byte(unitTest), 0644)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to write file: %v", err))
				fmt.Println(unitTest)
				return
			}
		}

		// Open the file in append mode, create if it doesn't exist
		file, err := os.OpenFile(out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			return
		}
		defer file.Close()

		// Write the content to the file
		if _, err := file.WriteString(unitTest); err != nil {
			fmt.Printf("Failed to write to file: %v\n", err)
			return
		}

		fmt.Printf("Test case written to file: %s\n", out)
		return
	}

	fmt.Println(unitTest)
}

// getTempDir returns the temporary directory
func getTempDir() string {
	dir := "/tmp/unit"

	// create dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	return dir
}

// writeCachedSuggestions writes the test cases to the cache
func writeCachedSuggestions(h hash.Hash, suggestions []Suggestion) error {
	// Write the suggestions to a cache file
	cachedTestFile := fmt.Sprintf("%s/%x", getTempDir(), h.Sum(nil))
	content, err := json.Marshal(suggestions)
	if err != nil {
		return fmt.Errorf("failed to marshal suggestions: %w", err)
	}

	err = os.WriteFile(cachedTestFile, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// readTestsFromCache reads the test cases from the cache
func readCachedSuggestions(h hash.Hash) ([]Suggestion, error) {
	// Check if the tmp file exists
	cachedTestFile := fmt.Sprintf("%s/%x", getTempDir(), h.Sum(nil))
	_, err := os.Stat(cachedTestFile)

	// no file found
	if errors.Is(err, os.ErrNotExist) {

		return nil, errNotCached
	}

	// read the file content
	content, err := os.ReadFile(cachedTestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var suggestions []Suggestion
	err = json.Unmarshal(content, &suggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return suggestions, nil
}

func getSuggestedUnitTests(ctx context.Context, content []byte) ([]Suggestion, error) {
	var suggestions = []Suggestion{}

	sysPrompt := "You are a seasoned engineer that generates unit test suggestions for the provided code."
	usrPrompt := fmt.Sprintf("Analyze the following code and list possible unit tests that should be generated for great coverage. Do not explain. Document all functions clearly. \n\n### Code\n\n%s", content)

	suggestions, err := aiClient.GetSuggestions(ctx, sysPrompt, usrPrompt)
	if err != nil {
		return suggestions, fmt.Errorf("failed to call AI: %w", err)
	}

	return suggestions, nil
}

func selectSuggestedUnitTest(suggestions []Suggestion) (Suggestion, error) {
	index := 0

	// Step 5: Print the suggestions with index
	fmt.Printf("\nSuggested Unit Tests:\n")
	for i, s := range suggestions {
		fmt.Printf("\n%d. %s\n", i+1, s.Title)
		for _, reason := range s.Reasons {
			fmt.Print(au.Gray(0, fmt.Sprintf("   - %s\n", reason)).String())
		}
	}

	// Prompt the user
	fmt.Printf("\nSelect unit test: ")

	// Create a buffered reader
	reader := bufio.NewReader(os.Stdin)

	// Read input until the first newline
	input, err := reader.ReadString('\n')
	if err != nil {
		return Suggestion{}, fmt.Errorf("failed to read input: %w", err)
	}

	// Trim any trailing whitespace, including newline
	input = strings.TrimSpace(input)

	// Convert input to an integer
	index, err = strconv.Atoi(input)
	if err != nil {
		return Suggestion{}, fmt.Errorf("invalid input: %w", err)
	}

	// Validate index
	if index <= 0 || index > len(suggestions) {
		return Suggestion{}, fmt.Errorf("invalid index: %d", index)
	}

	return suggestions[index-1], nil
}

// createUnitTest generates a unit test for the given code snippet
func createUnitTest(ctx context.Context, content []byte, testCase Suggestion) (string, error) {

	usrPrompt := fmt.Sprintf("Write a unit test for the following code:\n### CODE\n%s\n\n### UNIT TEST\n: %v", content, testCase.Title)
	for _, reason := range testCase.Reasons {
		usrPrompt += fmt.Sprintf("\n- %s", reason)
	}

	sysPrompt := `
	"You are a seasoned engineer who writes amazing unit tests.
	
	First write your unit test.
	
	Second review you code:
	- Validate that it is efficient and effective.
	- Validateyour imports -- they must be real.
	- Validate your functions -- they must be real.
	- Validate your code -- it must be correct.

	Third, you take time to optimize your test.

	Finally return your unit test. Do not explain.
`
	// perform api call
	test, err := aiClient.CreateTest(ctx, sysPrompt, usrPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	return test, nil
}
