package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed init_templates/generic_schema.yaml
var genericSchemaTmpl string

//go:embed init_templates/generic_example.md
var genericExampleTmpl string

//go:embed init_templates/aspice_schema.yaml
var aspiceSchemaTmpl string

//go:embed init_templates/aspice_example.md
var aspiceExampleTmpl string

type initConfig struct {
	Dir      string
	Preset   string
	ID       string
	Title    string
	Level    string
	IDPrefix string
	Force    bool
}

func newInitCmd() *cobra.Command {
	var cfg initConfig

	cmd := &cobra.Command{
		Use:   "init <directory>",
		Short: "Scaffold a new requirements directory with schema.yaml and example",
		Long: `Scaffold a new requirements directory.

Creates a schema.yaml file and an example .md file with sample requirements
to help users get started quickly.

Presets:
  generic   Generic requirements template (default)
  aspice    Automotive SPICE-oriented template with ASIL and safety attributes`,
		Example: `  reqmd init my-project/
  reqmd init my-project/ --preset aspice
  reqmd init my-project/ --preset aspice --id-prefix REQ`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Dir = args[0]
			return runInit(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.Preset, "preset", "generic", "Template preset (generic or aspice)")
	cmd.Flags().StringVar(&cfg.ID, "id", "", "Schema $id (default: directory basename)")
	cmd.Flags().StringVar(&cfg.Title, "title", "", "Schema title (default: '<dir> Requirements')")
	cmd.Flags().StringVar(&cfg.Level, "level", "requirements", "x-reqmd level")
	cmd.Flags().StringVar(&cfg.IDPrefix, "id-prefix", "", "x-reqmd id-prefix")
	cmd.Flags().BoolVar(&cfg.Force, "force", false, "Overwrite existing files")
	return cmd
}

func runInit(cfg initConfig) error {
	// Validate preset
	var schemaTmpl, exampleTmpl string
	switch cfg.Preset {
	case "generic":
		schemaTmpl = genericSchemaTmpl
		exampleTmpl = genericExampleTmpl
	case "aspice":
		schemaTmpl = aspiceSchemaTmpl
		exampleTmpl = aspiceExampleTmpl
	default:
		return fmt.Errorf("unknown preset %q; valid presets: generic, aspice", cfg.Preset)
	}

	// Create directory if needed
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", cfg.Dir, err)
	}

	// Check for existing files
	schemaPath := filepath.Join(cfg.Dir, "schema.yaml")
	examplePath := filepath.Join(cfg.Dir, "requirements.md")

	if !cfg.Force {
		for _, p := range []string{schemaPath, examplePath} {
			if _, err := os.Stat(p); err == nil {
				return fmt.Errorf("%s already exists; use --force to overwrite", p)
			}
		}
	}

	// Fill template data
	dirName := filepath.Base(cfg.Dir)
	id := cfg.ID
	if id == "" {
		id = dirName
	}
	title := cfg.Title
	if title == "" {
		title = dirName + " Requirements"
	}
	level := cfg.Level
	data := struct {
		ID       string
		Title    string
		Level    string
		IDPrefix string
	}{
		ID:       id,
		Title:    title,
		Level:    level,
		IDPrefix: cfg.IDPrefix,
	}

	// Write schema.yaml
	if err := writeTemplate(schemaPath, "schema", schemaTmpl, data); err != nil {
		return fmt.Errorf("writing schema: %w", err)
	}

	// Write example .md
	if err := writeTemplate(examplePath, "example", exampleTmpl, data); err != nil {
		return fmt.Errorf("writing example: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Initialized %s requirements in %s/\n", cfg.Preset, cfg.Dir)
	fmt.Fprintf(os.Stderr, "  %s\n", schemaPath)
	fmt.Fprintf(os.Stderr, "  %s\n", examplePath)
	fmt.Fprintf(os.Stderr, "Run: reqmd check %s/\n", cfg.Dir)
	return nil
}

func writeTemplate(path, name, tmplStr string, data any) error {
	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", name, err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}
