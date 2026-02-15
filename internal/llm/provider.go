package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/samcharles93/commiter/internal/config"
)

const (
	maxSummaryDiffChars = 12000
)

// Provider is the interface for LLM providers
type Provider interface {
	GenerateMessage(ctx context.Context, diff string, history []Message, template *config.CommitTemplate) (string, error)
	SummarizeChanges(ctx context.Context, diff string) (string, error)
}

// GenericProvider implements Provider for generic LLM APIs
type GenericProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewGenericProvider creates a new GenericProvider
func NewGenericProvider(apiKey, model, baseURL string) *GenericProvider {
	return &GenericProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *GenericProvider) complete(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    s.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		if urlErr, ok := errors.AsType[*url.Error](err); ok && urlErr.Timeout() {
			return "", fmt.Errorf("request to LLM timed out: %w", err)
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// GenerateMessage generates a commit message from a diff
func (s *GenericProvider) GenerateMessage(ctx context.Context, diff string, history []Message, template *config.CommitTemplate) (string, error) {
	systemPrompt, _ := os.ReadFile("SYSTEM.md")
	memoryPrompt, _ := os.ReadFile("MEMORY.md")

	fullSystemPrompt := string(systemPrompt)
	if len(memoryPrompt) > 0 {
		fullSystemPrompt += "\n\nUser Preferences:\n" + string(memoryPrompt)
	}

	// Add template instructions if provided
	if template != nil && template.Prompt != "" {
		fullSystemPrompt += "\n\nCommit Message Template Instructions:\n" + template.Prompt
		if template.Format != "" {
			fullSystemPrompt += "\n\nUse this format: " + template.Format
		}
	}

	messages := []Message{
		{Role: "system", Content: fullSystemPrompt},
	}

	if len(history) == 0 {
		messages = append(messages, Message{
			Role:    "user",
			Content: fmt.Sprintf("Analyze this git diff and provide a commit message:\n\n%s", diff),
		})
	} else {
		messages = append(messages, history...)
	}

	return s.complete(ctx, messages)
}

// SummarizeChanges generates a summary of the changes
func (s *GenericProvider) SummarizeChanges(ctx context.Context, diff string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "You summarize git diffs for developers. Return 2-4 short bullet points. " +
				"Focus on behavior and intent, and include file names where helpful.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Summarize these staged changes:\n\n%s", truncateDiffForSummary(diff)),
		},
	}

	return s.complete(ctx, messages)
}

func truncateDiffForSummary(diff string) string {
	if len(diff) <= maxSummaryDiffChars {
		return diff
	}
	return diff[:maxSummaryDiffChars] + "\n...[diff truncated for summary]"
}
