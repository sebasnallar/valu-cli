package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sebasnallar/valu-cli/pkg/dsl/parser"
	"github.com/sebasnallar/valu-cli/pkg/template"
	"github.com/sebasnallar/valu-cli/pkg/validation"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "valu",
		Short: "Valu - Infrastructure as Code made simple",
		Long: `A modern tool for defining and managing infrastructure configurations 
               with support for templates and advanced validation.`,
	}

	var validateCmd = &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a configuration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			validator, err := validation.NewValidator()
			if err != nil {
				return fmt.Errorf("initializing validator: %w", err)
			}

			p := parser.New()
			scope, err := p.Parse(args[0])
			if err != nil {
				return fmt.Errorf("parsing configuration: %w", err)
			}

			if err := validator.ValidateScope(scope); err != nil {
				return fmt.Errorf("validating configuration: %w", err)
			}

			fmt.Printf("✅ Configuration is valid! Scope: %s\n", scope.Name)
			return nil
		},
	}

	var templateCmd = &cobra.Command{
		Use:   "template",
		Short: "Work with templates",
	}

	var listTemplatesCmd = &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available templates:")
			fmt.Println("- web-service")
			fmt.Println("- database")
			return nil
		},
	}

	var renderTemplateCmd = &cobra.Command{
		Use:   "render [template] [output]",
		Short: "Render a template to a file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			varsFile, _ := cmd.Flags().GetString("vars")

			vars := make(map[string]interface{})
			if varsFile != "" {
				data, err := os.ReadFile(varsFile)
				if err != nil {
					return fmt.Errorf("reading variables file: %w", err)
				}
				if err := yaml.Unmarshal(data, &vars); err != nil {
					return fmt.Errorf("parsing variables: %w", err)
				}
			}

			engine := template.NewTemplateEngine()

			if err := engine.LoadTemplateFromFile("main", args[0]); err != nil {
				return fmt.Errorf("loading template: %w", err)
			}

			templateData := template.TemplateData{
				Name:        "my-app",
				Version:     "1.0.0",
				Environment: "development",
				Variables:   vars,
			}

			rendered, err := engine.RenderTemplate("main", templateData)
			if err != nil {
				return fmt.Errorf("rendering template: %w", err)
			}

			if err := os.WriteFile(args[1], rendered, 0644); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			fmt.Printf("✅ Template rendered successfully to %s\n", args[1])
			return nil
		},
	}

	var schemaCmd = &cobra.Command{
		Use:   "schema",
		Short: "Work with resource schemas",
	}

	var showSchemaCmd = &cobra.Command{
		Use:   "show [type]",
		Short: "Show schema for a resource type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			validator, err := validation.NewValidator()
			if err != nil {
				return fmt.Errorf("initializing validator: %w", err)
			}

			schema, err := validator.GetSchema(args[0])
			if err != nil {
				return fmt.Errorf("getting schema: %w", err)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(schema); err != nil {
				return fmt.Errorf("encoding schema: %w", err)
			}

			return nil
		},
	}

	var listSchemasCmd = &cobra.Command{
		Use:   "list",
		Short: "List available schemas",
		RunE: func(cmd *cobra.Command, args []string) error {
			validator, err := validation.NewValidator()
			if err != nil {
				return fmt.Errorf("initializing validator: %w", err)
			}

			schemas := validator.GetSupportedTypes()
			fmt.Println("Available schemas:")
			for _, s := range schemas {
				fmt.Printf("- %s\n", s)
			}
			return nil
		},
	}

	renderTemplateCmd.Flags().String("vars", "", "Variables file in YAML format")

	templateCmd.AddCommand(listTemplatesCmd)
	templateCmd.AddCommand(renderTemplateCmd)

	schemaCmd.AddCommand(showSchemaCmd)
	schemaCmd.AddCommand(listSchemasCmd)

	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(schemaCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
