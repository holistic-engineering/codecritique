package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type Provider string

const (
	ProviderAnthropic Provider = "Anthropic"
	ProviderGroq      Provider = "Groq"
	ProviderOllama    Provider = "Ollama"
	ProviderOpenAI    Provider = "OpenAI"
)

type Client struct {
	provider    Provider
	ollamaURL   string
	ollamaModel string
	groqAPIKey  string
	groqModel   string
}

func New(cfg *config.AIConfig) *Client {
	return &Client{
		provider:    Provider(cfg.Provider),
		ollamaURL:   cfg.OllamaURL,
		ollamaModel: cfg.OllamaModel,
		groqAPIKey:  cfg.GroqAPIKey,
		groqModel:   cfg.GroqModel,
	}
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

	review := &model.Review{
		PullRequest: pr,
		Suggestions: []*model.Suggestion{},
	}

	// Parse the Groq response and create suggestions
	// This is a simplified example; you may need to adjust based on your specific requirements
	suggestion := &model.Suggestion{
		File:    pr.Files[0], // Assuming the suggestion is for the first file
		Line:    1,           // Placeholder line number
		Message: content,
	}
	review.Suggestions = append(review.Suggestions, suggestion)

	return review, nil
}

const reviewPromptTemplate = `You are PR-Reviewer, an AI language model designed to review pull requests.

Your goal is to review the code changes in the provided pull request and offer feedback and suggestions for improvement.
Be informative, constructive, and give examples. Try to be as specific as possible.

Please provide your review in JSON format with the following structure:
{
  "review": {
    "summary": "A brief summary of the PR",
    "overall_impression": "Your overall impression of the changes",
    "code_quality": {
      "strengths": ["List of strengths in the code"],
      "areas_for_improvement": ["List of areas that could be improved"]
    },
    "potential_issues": ["List of potential issues or bugs"],
    "suggestions": ["List of suggestions for improvement"],
    "security_concerns": "Any security concerns, or 'None identified' if none",
    "testing": "Comments on test coverage and suggestions for additional tests",
    "estimated_effort_to_review": "Estimated effort to review on a scale of 1-5",
    "code_feedback": [
      {
        "file": "Filename",
        "line": "Line number (if applicable)",
        "suggestion": "Specific suggestion for this file/line"
      }
    ]
  }
}

Here are some guidelines for your review:
- Focus on code quality, potential issues, and suggestions for improvement.
- Comment on code readability, maintainability, and adherence to best practices.
- Identify any potential bugs or edge cases that may not be handled.
- Suggest optimizations or alternative approaches where appropriate.
- Consider the overall architecture and design of the changes.
- Assess whether the code changes match the PR description and solve the intended problem.
- Evaluate test coverage and suggest additional test scenarios if needed.

PR Information:
Title: {{.Title}}
Description: {{.Description}}

Files changed:
{{range .Files}}- {{.Name}}
{{end}}

File contents:
{{range .Files}}
File: {{.Name}}
Content:
{{.Content}}

{{end}}
Please review the provided pull request and provide your feedback in the JSON format specified above.`

func (c *Client) generatePrompt(pr *model.PullRequest) (string, error) {
	tmpl, err := template.New("reviewPrompt").Parse(reviewPromptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pr)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
