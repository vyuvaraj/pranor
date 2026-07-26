package templates

import (
	"strings"
	"testing"
)

func TestTemplateEngine_RenderWithPartialsAndLoops(t *testing.T) {
	engine := NewTemplateEngine()
	engine.RegisterPartial("header", "<header><h1>Welcome {{.AppName}}</h1></header>")

	tpl := `{{partial "header"}}
{{if .IsAdmin}}
<p>Admin Dashboard</p>
{{end}}
<ul>
{{range .Items}}
  <li>Item: {{.}}</li>
{{end}}
</ul>`

	data := map[string]interface{}{
		"AppName": "ServMail",
		"IsAdmin": true,
		"Items":   []string{"Email 1", "Email 2"},
	}

	output, err := engine.RenderTemplate(tpl, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(output, "<h1>Welcome ServMail</h1>") {
		t.Errorf("header partial output missing: %s", output)
	}
	if !strings.Contains(output, "<p>Admin Dashboard</p>") {
		t.Errorf("conditional block output missing: %s", output)
	}
	if !strings.Contains(output, "<li>Item: Email 1</li>") || !strings.Contains(output, "<li>Item: Email 2</li>") {
		t.Errorf("loop block output missing: %s", output)
	}
}
