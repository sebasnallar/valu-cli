package template

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

type TemplateEngine struct {
	templates map[string]*template.Template
}

type TemplateData struct {
	Name        string
	Version     string
	Environment string
	Variables   map[string]interface{}
}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
	}
}

func (te *TemplateEngine) RegisterTemplate(name, content string) error {
	tmpl := template.New(name).Funcs(sprig.TxtFuncMap())

	parsedTmpl, err := tmpl.Parse(content)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", name, err)
	}

	te.templates[name] = parsedTmpl
	return nil
}

func (te *TemplateEngine) RenderTemplate(name string, data TemplateData) ([]byte, error) {
	tmpl, ok := te.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering template: %w", err)
	}

	var tmp interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &tmp); err != nil {
		return nil, fmt.Errorf("rendered template is not valid YAML: %w", err)
	}

	return buf.Bytes(), nil
}

func (te *TemplateEngine) LoadTemplateFromFile(name, filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading template file: %w", err)
	}

	return te.RegisterTemplate(name, string(content))
}
