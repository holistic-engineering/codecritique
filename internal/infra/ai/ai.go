package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type Provider string

const (
	ProviderOllama    Provider = "Ollama"
	ProviderOpenAI    Provider = "OpenAI"
	ProviderAnthropic Provider = "Anthropic"
	ProviderGroq      Provider = "Groq"
)

type Client struct {
	provider    Provider
	ollamaURL   string
	ollamaModel string
}

func New(provider Provider) *Client {
	return &Client{
		provider:    provider,
		ollamaURL:   "http://localhost:11434/api/generate",
		ollamaModel: "llama3.1",
	}
}

func (c *Client) Review(ctx context.Context, pr *model.PullRequest) (*model.Review, error) {
	switch c.provider {
	case ProviderOllama:
		return c.reviewWithOllama(ctx, pr)
	case ProviderOpenAI, ProviderAnthropic, ProviderGroq:
		return nil, fmt.Errorf("AI provider %s not implemented yet", c.provider)
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", c.provider)
	}
}

func (c *Client) reviewWithOllama(ctx context.Context, pr *model.PullRequest) (*model.Review, error) {
	prompt := c.generatePrompt(pr)

	requestBody, err := json.Marshal(map[string]string{
		"model":  c.ollamaModel,
		"prompt": prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.ollamaURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned non-OK status: %s", resp.Status)
	}

	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var result map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
		}
		if response, ok := result["response"].(string); ok {
			fullResponse.WriteString(response)
		}
		if done, ok := result["done"].(bool); ok && done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Ollama response: %w", err)
	}

	review := &model.Review{
		PullRequest: pr,
		Suggestions: []*model.Suggestion{},
	}

	// Parse the full Ollama response and create suggestions
	// This is a simplified example; you may need to adjust based on your specific requirements
	suggestion := &model.Suggestion{
		File:    pr.Files[0], // Assuming the suggestion is for the first file
		Line:    1,           // Placeholder line number
		Message: fullResponse.String(),
	}
	review.Suggestions = append(review.Suggestions, suggestion)

	return review, nil
}

func (c *Client) generatePrompt(pr *model.PullRequest) string {
	prompt := fmt.Sprintf("Review the following pull request:\n\nTitle: %s\nDescription: %s\n\n", pr.Title, pr.Description)
	for _, file := range pr.Files {
		prompt += fmt.Sprintf("File: %s\nContent:\n%s\n\n", file.Name, file.Content)
	}
	prompt += "Please provide a code review for this pull request. Focus on code quality, potential issues, and suggestions for improvement."
	return prompt
}
