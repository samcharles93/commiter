package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

const defaultMarkdownWidth = 80

// MarkdownRenderer renders markdown content with ANSI styling and
// falls back to plain text on renderer failures.
type MarkdownRenderer struct {
	width   int
	factory rendererFactory
}

type termRenderer interface {
	Render(string) (string, error)
}

type rendererFactory func(width int) (termRenderer, error)

// NewMarkdownRenderer creates a markdown renderer with default width.
func NewMarkdownRenderer() MarkdownRenderer {
	return MarkdownRenderer{
		width:   defaultMarkdownWidth,
		factory: defaultRendererFactory,
	}
}

// SetWidth updates wrapping width.
func (r *MarkdownRenderer) SetWidth(width int) {
	if width <= 0 {
		width = defaultMarkdownWidth
	}
	r.width = width
}

// Render renders markdown; if rendering fails, the raw input is returned.
func (r MarkdownRenderer) Render(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}

	width := r.width
	if width <= 0 {
		width = defaultMarkdownWidth
	}

	factory := r.factory
	if factory == nil {
		factory = defaultRendererFactory
	}

	renderer, err := factory(width)
	if err != nil {
		return content
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		return content
	}

	return strings.TrimRight(rendered, "\n")
}

func defaultRendererFactory(width int) (termRenderer, error) {
	renderer, err := glamour.NewTermRenderer(glamour.WithWordWrap(width))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize markdown renderer: %w", err)
	}
	return renderer, nil
}
