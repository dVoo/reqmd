package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

//go:embed init_templates/results_schema.yaml
var resultsSchemaTmpl string

//go:embed init_templates/results_example.md
var resultsExampleTmpl string

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
		Long: `Scaffold a new requirements directory with schema.yaml and an example .md file.

Built-in presets:
  generic   Generic requirements template (default)
  aspice    Automotive SPICE-oriented template with ASIL and safety attributes
  results   Manual verification results template (review/inspection/analysis)

Custom presets:
  --preset <dir>   Use a directory containing schema.yaml.tmpl + example.md.tmpl
                   (or schema.yaml + any .md file). Templates support Go template
                   syntax with .ID, .Title, .Level, and .IDPrefix variables.`,
		Example: `  reqmd init my-project/
  reqmd init my-project/ --preset aspice
  reqmd init my-project/ --preset aspice --id-prefix REQ
  reqmd init my-results/ --preset results --id-prefix VR
  reqmd init my-project/ --preset ./my-preset/ --id-prefix MY`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Dir = args[0]
			return runInit(cfg)
		},
	}
	cmd.Flags().StringVar(&cfg.Preset, "preset", "generic", "Template preset (generic, aspice, results) or path to a custom preset directory")
	cmd.Flags().StringVar(&cfg.ID, "id", "", "Schema $id (default: directory basename)")
	cmd.Flags().StringVar(&cfg.Title, "title", "", "Schema title (default: '<dir> Requirements')")
	cmd.Flags().StringVar(&cfg.Level, "level", "requirements", "x-reqmd level")
	cmd.Flags().StringVar(&cfg.IDPrefix, "id-prefix", "", "x-reqmd id-prefix")
	cmd.Flags().BoolVar(&cfg.Force, "force", false, "Overwrite existing files")
	return cmd
}

func runInit(cfg initConfig) error {
	// Resolve preset: named built-in or custom directory path.
	var schemaTmpl, exampleTmpl string
	var exampleName string

	if isDirPath(cfg.Preset) {
		// Custom preset: load templates from a directory.
		customDir, err := filepath.Abs(cfg.Preset)
		if err != nil {
			return fmt.Errorf("resolving preset path: %w", err)
		}
		schemaTmpl, exampleTmpl, exampleName, err = loadCustomPreset(customDir)
		if err != nil {
			return err
		}
	} else {
		switch cfg.Preset {
		case "generic":
			schemaTmpl = genericSchemaTmpl
			exampleTmpl = genericExampleTmpl
			exampleName = "requirements.md"
		case "aspice":
			schemaTmpl = aspiceSchemaTmpl
			exampleTmpl = aspiceExampleTmpl
			exampleName = "requirements.md"
		case "results":
			schemaTmpl = resultsSchemaTmpl
			exampleTmpl = resultsExampleTmpl
			exampleName = "results.md"
		default:
			return fmt.Errorf("unknown preset %q; valid presets: generic, aspice, results, or a directory path", cfg.Preset)
		}
	}

	// Create directory if needed
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", cfg.Dir, err)
	}

	// Check for existing files
	schemaPath := filepath.Join(cfg.Dir, "schema.yaml")
	examplePath := filepath.Join(cfg.Dir, exampleName)

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
	// Results preset defaults: VR prefix, verify-results level, different title.
	idPrefix := cfg.IDPrefix
	level := cfg.Level
	title := cfg.Title
	if cfg.Preset == "results" {
		if idPrefix == "" {
			idPrefix = "VR"
		}
		level = "verify-results"
		if title == "" {
			title = dirName + " Verification Results"
		}
	} else if title == "" {
		title = dirName + " Requirements"
	}
	data := struct {
		ID       string
		Title    string
		Level    string
		IDPrefix string
	}{
		ID:       id,
		Title:    title,
		Level:    level,
		IDPrefix: idPrefix,
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

// isDirPath reports whether the preset string looks like a filesystem path
// (contains a path separator or ends in a known extension), as opposed to a
// named preset like "generic", "aspice", or "results".
func isDirPath(preset string) bool {
	if strings.ContainsAny(preset, "/\\.") || preset == "." || preset == ".." {
		return true
	}
	if _, err := os.Stat(preset); err == nil {
		return true
	}
	return false
}

// loadCustomPreset reads schema and example templates from a directory.
// It looks for files in this order of preference:
//   - schema.yaml.tmpl + example.md.tmpl  (Go template files)
//   - schema.yaml       + example.md      (plain files, used verbatim)
//
// The example filename (without .tmpl) is used as the output filename.
// If multiple .md files exist, the first one found is used as the example.
func loadCustomPreset(dir string) (string, string, string, error) {
	// Try .tmpl variants first
	schemaPath := filepath.Join(dir, "schema.yaml.tmpl")
	exampleTmplPath := filepath.Join(dir, "example.md.tmpl")
	exampleName := "requirements.md"

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		// Fall back to plain schema.yaml
		schemaPath = filepath.Join(dir, "schema.yaml")
		schemaBytes, err = os.ReadFile(schemaPath)
		if err != nil {
			return "", "", "", fmt.Errorf("custom preset: no schema.yaml.tmpl or schema.yaml in %s", dir)
		}
	}

	exampleBytes, err := os.ReadFile(exampleTmplPath)
	if err != nil {
		// Fall back: find the first .md file in the directory
		entries, err2 := os.ReadDir(dir)
		if err2 != nil {
			return "", "", "", fmt.Errorf("reading custom preset dir: %w", err2)
		}
		found := false
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			exampleTmplPath = filepath.Join(dir, e.Name())
			exampleName = e.Name()
			exampleBytes, err = os.ReadFile(exampleTmplPath)
			if err != nil {
				return "", "", "", fmt.Errorf("reading example file %s: %w", exampleTmplPath, err)
			}
			found = true
			break
		}
		if !found {
			return "", "", "", fmt.Errorf("custom preset: no example.md.tmpl or .md file in %s", dir)
		}
	} else {
		exampleName = "requirements.md"
	}

	return string(schemaBytes), string(exampleBytes), exampleName, nil
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
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("rendering template to %s: %w", path, err)
	}
	return nil
}
