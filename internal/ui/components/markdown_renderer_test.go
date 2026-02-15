package components

import (
	"errors"
	"testing"
)

func TestMarkdownRendererRender(t *testing.T) {
	t.Parallel()

	renderer := NewMarkdownRenderer()
	out := renderer.Render("# Title\n\n- item")
	if out == "" {
		t.Fatal("expected rendered markdown output, got empty string")
	}
}

func TestMarkdownRendererFallbackOnFactoryError(t *testing.T) {
	t.Parallel()

	renderer := MarkdownRenderer{
		width: 80,
		factory: func(int) (termRenderer, error) {
			return nil, errors.New("factory failed")
		},
	}

	input := "plain text"
	if got := renderer.Render(input); got != input {
		t.Fatalf("expected raw fallback output %q, got %q", input, got)
	}
}

func TestMarkdownRendererFallbackOnRenderError(t *testing.T) {
	t.Parallel()

	renderer := MarkdownRenderer{
		width: 80,
		factory: func(int) (termRenderer, error) {
			return failingTermRenderer{}, nil
		},
	}

	input := "plain text"
	if got := renderer.Render(input); got != input {
		t.Fatalf("expected raw fallback output %q, got %q", input, got)
	}
}

type failingTermRenderer struct{}

func (failingTermRenderer) Render(string) (string, error) {
	return "", errors.New("render failed")
}
