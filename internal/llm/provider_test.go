package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateMessage(t *testing.T) {
	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Bearer token, got %s", r.Header.Get("Authorization"))
		}

		resp := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: "feat: add unit tests",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	provider := NewGenericProvider("test-key", "test-model", ts.URL)
	ctx := context.Background()
	msg, err := provider.GenerateMessage(ctx, "test diff", nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "feat: add unit tests"
	if msg != expected {
		t.Errorf("Expected message %q, got %q", expected, msg)
	}
}
