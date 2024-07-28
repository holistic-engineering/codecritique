package ai

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

//go:embed prompts/reviewer.prompt
var promptFS embed.FS

type Provider string

const (
	ProviderAnthropic Provider = "Anthropic"
	ProviderGroq      Provider = "Groq"
	ProviderOllama    Provider = "Ollama"
	ProviderOpenAI    Provider = "OpenAI"
)

type Client struct {
	provider         Provider
	ollamaURL        string
	ollamaModel      string
	groqAPIKey       string
	groqModel        string
	reviewerTemplate *template.Template
}

func New(cfg *config.AIConfig) (*Client, error) {
	reviewerPrompt, err := promptFS.ReadFile("prompts/reviewer.prompt")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt template: %w", err)
	}

	tmpl, err := template.New("reviewPrompt").Parse(string(reviewerPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt template: %w", err)
	}

	return &Client{
		provider:         Provider(cfg.Provider),
		ollamaURL:        cfg.OllamaURL,
		ollamaModel:      cfg.OllamaModel,
		groqAPIKey:       cfg.GroqAPIKey,
		groqModel:        cfg.GroqModel,
		reviewerTemplate: tmpl,
	}, nil
}

func (c *Client) Review(ctx context.Context, pr *model.PullRequest) (*model.Review, error) {
	switch c.provider {
	case ProviderOllama:
		return c.reviewWithOllama(ctx, pr)
	case ProviderGroq:
		return c.reviewWithGroq(ctx, pr)
	case ProviderOpenAI, ProviderAnthropic:
		return nil, fmt.Errorf("AI provider %s not implemented yet", c.provider)
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", c.provider)
	}
}

func (c *Client) reviewWithOllama(ctx context.Context, pr *model.PullRequest) (*model.Review, error) {
	prompt, err := c.generatePrompt(pr)
	if err != nil {
		return nil, fmt.Errorf("could not generate prompt: %w", err)
	}

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

	return c.parseResponse(fullResponse.String(), pr)
}

func (c *Client) reviewWithGroq(ctx context.Context, pr *model.PullRequest) (*model.Review, error) {
	prompt, err := c.generatePrompt(pr)
	if err != nil {
		return nil, fmt.Errorf("could not generate prompt: %w", err)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": c.groqModel,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a code review assistant."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  1024,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.groqAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Groq: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groq returned non-OK status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Groq response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("invalid response format from Groq")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid message format in Groq response")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid content in Groq response")
	}

	return c.parseResponse(content, pr)
}

func (c *Client) generatePrompt(pr *model.PullRequest) (string, error) {
	var buf bytes.Buffer
	err := c.reviewerTemplate.Execute(&buf, pr)
	if err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return buf.String(), nil
}

func (c *Client) parseResponse(response string, pr *model.PullRequest) (*model.Review, error) {
	var reviewData struct {
		Review model.Review `json:"review"`
	}

	err := json.Unmarshal([]byte(response), &reviewData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI response: %w", err)
	}

	reviewData.Review.PullRequest = pr
	return &reviewData.Review, nil
}
