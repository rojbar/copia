package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rojbar/copia/internal/parser"
)

// ErrEmptyOutputDir is returned when the output directory is empty
var ErrEmptyOutputDir = fmt.Errorf("output directory cannot be empty")

// Option is a functional option for configuring the writer
type Option func(*PackageWriter)

// PackageWriter is responsible for writing parsed templates to the output directory
type PackageWriter struct {
	outputDir    string
	dryRun       bool
	verbose      bool
	varExtension string
}

// WithVerbose sets the verbosity of the writer
func WithVerbose(verbose bool) Option {
	return func(p *PackageWriter) {
		p.verbose = verbose
	}
}

// WithDryRun sets the dry run mode of the writer
func WithDryRun(dryRun bool) Option {
	return func(p *PackageWriter) {
		p.dryRun = dryRun
	}
}

// WithOutputDir sets the output directory of the writer
func WithOutputDir(outputDir string) Option {
	return func(p *PackageWriter) {
		p.outputDir = outputDir
	}
}

// WithVarExtension sets the variable file extension used in templates
func WithVarExtension(ext string) Option {
	return func(p *PackageWriter) {
		p.varExtension = ext
	}
}

// New creates a new PackageWriter with the given options
func New(opts ...Option) *PackageWriter {
	p := &PackageWriter{
		outputDir:    "./output",
		dryRun:       false,
		verbose:      false,
		varExtension: ".tmpl",
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Write processes templates and writes them to the output directory
func (w *PackageWriter) Write(templatePackage *parser.TemplatePackage) error {
	if w.outputDir == "" {
		return ErrEmptyOutputDir
	}

	if !w.dryRun {
		if err := os.MkdirAll(w.outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	for _, tmpl := range templatePackage.Templates {
		outputPath := filepath.Join(w.outputDir, tmpl.Path)

		outputPath = strings.TrimSuffix(outputPath, w.varExtension)

		if w.verbose {
			fmt.Printf("Processing: %s -> %s\n", tmpl.Path, outputPath)
		}

		t, err := template.New(tmpl.Path).Parse(tmpl.Content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", tmpl.Path, err)
		}

		if w.dryRun {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", outputPath, err)
		}

		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}

		if err := t.Execute(outFile, templatePackage.Variables); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to execute template %s: %w", tmpl.Path, err)
		}

		if err := outFile.Close(); err != nil {
			return fmt.Errorf("failed to close output file %s: %w", outputPath, err)
		}

		if w.verbose {
			fmt.Printf("Written: %s\n", outputPath)
		}
	}

	return nil
}
