package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCountDiffFiles(t *testing.T) {
	single := `diff --git a/a.go b/a.go
index 111..222 100644
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old
+new`

	multi := `diff --git a/a.go b/a.go
index 111..222 100644
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old
+new
diff --git a/b.go b/b.go
index 333..444 100644
--- a/b.go
+++ b/b.go
@@ -1 +1 @@
-x
+y`

	if got := countDiffFiles(single); got != 1 {
		t.Fatalf("expected single diff file count 1, got %d", got)
	}
	if got := countDiffFiles(multi); got != 2 {
		t.Fatalf("expected multi diff file count 2, got %d", got)
	}
}

func TestBuildCommitPromptMultiFileIncludesDetailGuidance(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
index 111..222 100644
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old
+new
diff --git a/b.go b/b.go
index 333..444 100644
--- a/b.go
+++ b/b.go
@@ -1 +1 @@
-x
+y`

	prompt := buildCommitPrompt(diff)
	if !strings.Contains(prompt, "spans 2 files") {
		t.Fatalf("expected multi-file guidance in prompt, got: %q", prompt)
	}
	if !strings.Contains(prompt, "1-3 bullets") {
		t.Fatalf("expected bullet guidance in prompt, got: %q", prompt)
	}
}

func TestBuildCommitPromptSingleFileStaysSimple(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
index 111..222 100644
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old
+new`

	prompt := buildCommitPrompt(diff)
	if strings.Contains(prompt, "1-3 bullets") {
		t.Fatalf("expected single-file prompt without multi-file guidance, got: %q", prompt)
	}
}
