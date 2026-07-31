package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
)

// TemplateEngine renders HTML and text email templates with partials, loops, and conditional blocks.
type TemplateEngine struct {
	mu       sync.RWMutex
	partials map[string]string // partialName -> template string
}

// NewTemplateEngine creates a TemplateEngine instance.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		partials: make(map[string]string),
	}
}

// RegisterPartial registers a reusable template partial snippet (e.g. "header", "footer").
func (te *TemplateEngine) RegisterPartial(name, content string) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.partials[name] = content
}

// RenderTemplate parses and executes a template string with data context.
func (te *TemplateEngine) RenderTemplate(templateStr string, data interface{}) (string, error) {
	te.mu.RLock()
	partials := make(map[string]string, len(te.partials))
	for k, v := range te.partials {
		partials[k] = v
	}
	te.mu.RUnlock()

	// Pre-process custom partial tag syntax: {{partial "header"}} -> {{template "header" .}}
	processedTplStr := processPartialTags(templateStr)

	tmpl := template.New("email")
	// Register all partial templates
	for pName, pContent := range partials {
		_, err := tmpl.New(pName).Parse(pContent)
		if err != nil {
			return "", fmt.Errorf("failed to parse partial '%s': %w", pName, err)
		}
	}

	tmpl, err := tmpl.Parse(processedTplStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

func processPartialTags(input string) string {
	// Simple replacement for {{partial "name"}} to Go template {{template "name" .}}
	out := input
	for {
		startIdx := strings.Index(out, "{{partial ")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(out[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		fullMatch := out[startIdx : startIdx+endIdx+2]
		rawArg := strings.TrimSpace(fullMatch[10 : len(fullMatch)-2]) // e.g. "header"
		replacement := fmt.Sprintf("{{template %s .}}", rawArg)
		out = strings.Replace(out, fullMatch, replacement, 1)
	}
	return out
}
