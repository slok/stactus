package html

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/yuin/goldmark"
)

// RenderMarkdownToHTML will get a markdown string and render to HTML.
func RenderMarkdownToHTML(mdText string) (template.HTML, error) {
	var b bytes.Buffer
	err := goldmark.Convert([]byte(mdText), &b)
	if err != nil {
		return "", fmt.Errorf("could not convert markdown to HTML: %w", err)
	}

	return template.HTML(b.Bytes()), nil
}
