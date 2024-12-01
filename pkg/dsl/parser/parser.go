package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/sebasnallar/valu-cli/pkg/dsl/types"
	"github.com/sebasnallar/valu-cli/pkg/validation"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	envPrefix string
	validator *validation.Validator
}

type Option func(*Parser)

func WithEnvPrefix(prefix string) Option {
	return func(p *Parser) {
		p.envPrefix = prefix
	}
}

func New(options ...Option) *Parser {
	validator, err := validation.NewValidator()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize validator: %v", err))
	}

	p := &Parser{
		envPrefix: "VALU_",
		validator: validator,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *Parser) Parse(filename string) (*types.Scope, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	fmt.Printf("Raw file content:\n%s\n", string(data))

	interpolated := p.interpolateEnvVars(string(data))

	var scope types.Scope
	if err := yaml.Unmarshal([]byte(interpolated), &scope); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	jsonData, _ := json.MarshalIndent(scope, "", "  ")
	fmt.Printf("\nParsed scope structure:\n%s\n", string(jsonData))

	if err := p.validator.ValidateScope(&scope); err != nil {
		return nil, fmt.Errorf("validating scope: %w", err)
	}

	return &scope, nil
}

func (p *Parser) interpolateEnvVars(content string) string {
	re := regexp.MustCompile(`\${([^:}]+)(:([^}]+))?}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		envVar := parts[1]
		defaultVal := parts[3]

		if val, exists := os.LookupEnv(p.envPrefix + envVar); exists {
			return val
		}

		if val, exists := os.LookupEnv(envVar); exists {
			return val
		}

		if defaultVal != "" {
			return defaultVal
		}

		return match
	})
}
