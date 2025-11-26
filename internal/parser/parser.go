package parser

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrMaxDepthExceeded is returned when the maximum directory depth is exceeded during parsing
	ErrMaxDepthExceeded = fmt.Errorf("maximum directory depth exceeded")
)

// PackageParser is responsible for parsing template packages from a given path
type PackageParser struct {
	maxDepth       int
	ignorePatterns []string
	varsfileName   string
}

// Option is a functional option for configuring the parser
type Option func(*PackageParser)

// WithMaxDepth sets the maximum directory depth to traverse, -1 to set unlimited
func WithMaxDepth(depth int) Option {
	return func(p *PackageParser) {
		p.maxDepth = depth
	}
}

// WithIgnorePatterns sets patterns to ignore during parsing
func WithIgnorePatterns(patterns ...string) Option {
	return func(p *PackageParser) {
		p.ignorePatterns = patterns
	}
}

// WithVarsFileName sets the variables file name to be used during parsing
func WithVarsFileName(fileName string) Option {
	return func(p *PackageParser) {
		p.varsfileName = fileName
	}
}

// New creates a new PackageParser with the given options
func New(opts ...Option) *PackageParser {
	p := &PackageParser{
		maxDepth:       100,
		ignorePatterns: []string{},
		varsfileName:   "package_vars.json",
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Template represents a template file with its variables
type Template struct {
	Path    string // Relative path from template root
	Content string // Template content
}

// TemplatePackage represents a parsed template package structure
type TemplatePackage struct {
	Templates []Template
	Path      string
	Variables map[string]any
}

// Parse reads a template package and returns all templates with their variables
func (p *PackageParser) Parse(packagePath string) (*TemplatePackage, error) {
	absPath, err := filepath.Abs(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve package path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access package path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("package path must be a directory: %s", absPath)
	}

	vars, err := p.fetchVars(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch variables: %w", err)
	}

	templatePackage := &TemplatePackage{
		Path:      absPath,
		Templates: make([]Template, 0),
		Variables: vars,
	}

	err = filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if p.isVariablesFile(relPath, d) {
			return nil
		}

		if p.reachedMaxDepth(p.getCurrentDepth(path, d)) {
			return ErrMaxDepthExceeded
		}

		if p.shouldIgnore(d.Name(), relPath) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}

		templatePackage.Templates = append(templatePackage.Templates, Template{
			Path:    relPath,
			Content: string(content),
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk template directory: %w", err)
	}

	return templatePackage, nil
}

func (p *PackageParser) fetchVars(path string) (map[string]any, error) {
	vars := make(map[string]any)

	varsData, err := os.ReadFile(filepath.Join(path, p.varsfileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read vars file %w", err)

	}

	err = json.Unmarshal(varsData, &vars)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vars file %w", err)
	}

	return vars, nil
}

func (p *PackageParser) getCurrentDepth(path string, d fs.DirEntry) int {
	depth := 0
	if path != "." {
		depth = strings.Count(filepath.ToSlash(path), "/")
		if !d.IsDir() {
			depth++ // Files are one level deeper than their directory
		}
	}

	return depth
}

// shouldIgnore checks if a file should be ignored based on patterns
func (p *PackageParser) shouldIgnore(filename, relPath string) bool {
	for _, pattern := range p.ignorePatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}

		// Match against relative path for patterns like "*/test/*"
		if matched, _ := filepath.Match(pattern, filepath.ToSlash(relPath)); matched {
			return true
		}
	}
	return false
}

func (p *PackageParser) reachedMaxDepth(currentDepth int) bool {
	return p.maxDepth >= 0 && currentDepth > p.maxDepth
}

func (p *PackageParser) isVariablesFile(path string, d fs.DirEntry) bool {
	return d.Name() == p.varsfileName && path == p.varsfileName
}
