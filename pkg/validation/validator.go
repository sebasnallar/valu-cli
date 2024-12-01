package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"sort"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schemas/*.json
var schemasFS embed.FS

type Validator struct {
	rootSchema *gojsonschema.Schema
	schemas    map[string]*gojsonschema.Schema
	rawSchemas map[string]map[string]interface{}
}

func NewValidator() (*Validator, error) {
	v := &Validator{
		schemas:    make(map[string]*gojsonschema.Schema),
		rawSchemas: make(map[string]map[string]interface{}),
	}

	if err := v.loadSchemas(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Validator) loadSchemas() error {
	rootSchemaBytes, err := schemasFS.ReadFile("schemas/scope.json")
	if err != nil {
		return fmt.Errorf("reading root schema: %w", err)
	}

	sl := gojsonschema.NewSchemaLoader()
	sl.Validate = true

	entries, err := schemasFS.ReadDir("schemas")
	if err != nil {
		return fmt.Errorf("reading schemas directory: %w", err)
	}

	for _, entry := range entries {
		if path.Ext(entry.Name()) != ".json" {
			continue
		}

		schemaBytes, err := schemasFS.ReadFile(path.Join("schemas", entry.Name()))
		if err != nil {
			return fmt.Errorf("reading schema %s: %w", entry.Name(), err)
		}

		var rawSchema map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &rawSchema); err != nil {
			return fmt.Errorf("parsing schema %s: %w", entry.Name(), err)
		}

		resourceType := path.Base(entry.Name()[:len(entry.Name())-5])
		v.rawSchemas[resourceType] = rawSchema

		if err := sl.AddSchema("schemas/"+entry.Name(), gojsonschema.NewBytesLoader(schemaBytes)); err != nil {
			return fmt.Errorf("adding schema %s to loader: %w", entry.Name(), err)
		}
	}

	rootSchema, err := sl.Compile(gojsonschema.NewBytesLoader(rootSchemaBytes))
	if err != nil {
		return fmt.Errorf("compiling root schema: %w", err)
	}

	v.rootSchema = rootSchema
	return nil
}

func (v *Validator) ValidateScope(data interface{}) error {
	if v.rootSchema == nil {
		return fmt.Errorf("root schema not loaded")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}

	result, err := v.rootSchema.Validate(gojsonschema.NewBytesLoader(jsonData))
	if err != nil {
		return fmt.Errorf("validating against schema: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, err := range result.Errors() {
			errors = append(errors, fmt.Sprintf("- %s: %s", err.Field(), err.Description()))
		}
		return fmt.Errorf("validation errors:\n%s", formatErrors(errors))
	}

	return nil
}

func (v *Validator) GetSupportedTypes() []string {
	types := make([]string, 0, len(v.rawSchemas))
	for schemaType := range v.rawSchemas {
		if schemaType != "scope" {
			types = append(types, schemaType)
		}
	}
	sort.Strings(types)
	return types
}

func (v *Validator) GetSchema(resourceType string) (map[string]interface{}, error) {
	if schema, exists := v.rawSchemas[resourceType]; exists {
		return schema, nil
	}
	return nil, fmt.Errorf("schema not found for type: %s", resourceType)
}

func formatErrors(errors []string) string {
	var result string
	for _, err := range errors {
		result += err + "\n"
	}
	return result
}
