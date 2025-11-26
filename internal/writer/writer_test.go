package writer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rojbar/copia/internal/parser"
	"github.com/rojbar/copia/internal/writer"
)

func TestPackageWriter_Write(t *testing.T) {
	tests := []struct {
		name             string
		setupFunc        func(t *testing.T) (*parser.TemplatePackage, string)
		options          []writer.Option
		wantFiles        map[string]string // path -> content
		wantErr          bool
		errContains      string
		checkDirNotExist bool
	}{
		{
			name: "simple_template_write",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "main.go",
							Content: `package main`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			wantFiles: map[string]string{
				"main.go": "package main",
			},
			wantErr: false,
		},
		{
			name: "template_with_variables",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "config.yaml",
							Content: "name: {{.name}}\nversion: {{.version}}",
						},
					},
					Variables: map[string]any{
						"name":    "myapp",
						"version": "1.0",
					},
				}
				return pkg, outputDir
			},
			wantFiles: map[string]string{
				"config.yaml": "name: myapp\nversion: 1.0",
			},
			wantErr: false,
		},
		{
			name: "nested_directory_structure",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    filepath.Join("internal", "handlers", "handler.go"),
							Content: `package handlers`,
						},
						{
							Path:    "main.go",
							Content: `package main`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			wantFiles: map[string]string{
				filepath.Join("internal", "handlers", "handler.go"): "package handlers",
				"main.go": "package main",
			},
			wantErr: false,
		},
		{
			name: "template_with_tmpl_extension_removed",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "main.go.tmpl",
							Content: `package main`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			options: []writer.Option{writer.WithVarExtension(".tmpl")},
			wantFiles: map[string]string{
				"main.go": "package main",
			},
			wantErr: false,
		},
		{
			name: "dry_run_should_not_create_files",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "main.go",
							Content: `package main`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			options:          []writer.Option{writer.WithDryRun(true)},
			wantFiles:        map[string]string{},
			wantErr:          false,
			checkDirNotExist: true,
		},
		{
			name: "empty_output_directory_returns_error",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "main.go",
							Content: `package main`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, ""
			},
			options:     []writer.Option{writer.WithOutputDir("")},
			wantErr:     true,
			errContains: "output directory cannot be empty",
		},
		{
			name: "invalid_template_syntax",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "bad.go",
							Content: `package {{.unclosed`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			wantErr:     true,
			errContains: "failed to parse template",
		},
		{
			name: "multiple_templates",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "file1.txt",
							Content: `content 1`,
						},
						{
							Path:    "file2.txt",
							Content: `content 2`,
						},
						{
							Path:    "file3.txt",
							Content: `content 3`,
						},
					},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			wantFiles: map[string]string{
				"file1.txt": "content 1",
				"file2.txt": "content 2",
				"file3.txt": "content 3",
			},
			wantErr: false,
		},
		{
			name: "template_with_complex_variables",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{
						{
							Path:    "readme.md",
							Content: "# {{.project}}\n\nVersion: {{.version}}\nAuthor: {{.author}}",
						},
					},
					Variables: map[string]any{
						"project": "MyProject",
						"version": "2.0.0",
						"author":  "John Doe",
					},
				}
				return pkg, outputDir
			},
			wantFiles: map[string]string{
				"readme.md": "# MyProject\n\nVersion: 2.0.0\nAuthor: John Doe",
			},
			wantErr: false,
		},
		{
			name: "empty_template_package",
			setupFunc: func(t *testing.T) (*parser.TemplatePackage, string) {
				outputDir := t.TempDir()
				pkg := &parser.TemplatePackage{
					Templates: []parser.Template{},
					Variables: map[string]any{},
				}
				return pkg, outputDir
			},
			wantFiles: map[string]string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, outputDir := tt.setupFunc(t)

			opts := append([]writer.Option{writer.WithOutputDir(outputDir)}, tt.options...)
			w := writer.New(opts...)
			err := w.Write(pkg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.checkDirNotExist {
				// For dry run, check that output directory was not created with files
				entries, _ := os.ReadDir(outputDir)
				if len(entries) > 0 {
					t.Errorf("dry run should not create files, but found %d entries", len(entries))
				}
				return
			}

			// Verify all expected files exist with correct content
			for filePath, expectedContent := range tt.wantFiles {
				fullPath := filepath.Join(outputDir, filePath)
				content, err := os.ReadFile(fullPath)
				if err != nil {
					t.Errorf("failed to read expected file %s: %v", filePath, err)
					continue
				}
				if string(content) != expectedContent {
					t.Errorf("file %s content mismatch:\ngot:  %q\nwant: %q", filePath, string(content), expectedContent)
				}
			}

			// Verify no extra files were created
			err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				relPath, err := filepath.Rel(outputDir, path)
				if err != nil {
					return err
				}
				if _, expected := tt.wantFiles[relPath]; !expected {
					t.Errorf("unexpected file created: %s", relPath)
				}
				return nil
			})
			if err != nil {
				t.Errorf("error walking output directory: %v", err)
			}
		})
	}
}

func TestPackageWriter_Options(t *testing.T) {
	tests := []struct {
		name    string
		options []writer.Option
	}{
		{
			name:    "default_options",
			options: []writer.Option{},
		},
		{
			name: "with_output_dir",
			options: []writer.Option{
				writer.WithOutputDir("/tmp/output"),
			},
		},
		{
			name: "with_dry_run",
			options: []writer.Option{
				writer.WithDryRun(true),
			},
		},
		{
			name: "with_verbose",
			options: []writer.Option{
				writer.WithVerbose(true),
			},
		},
		{
			name: "with_var_extension",
			options: []writer.Option{
				writer.WithVarExtension(".tmpl"),
			},
		},
		{
			name: "combined_options",
			options: []writer.Option{
				writer.WithOutputDir("/tmp/custom"),
				writer.WithDryRun(false),
				writer.WithVerbose(true),
				writer.WithVarExtension(".template"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := writer.New(tt.options...)
			if w == nil {
				t.Error("expected non-nil writer")
			}
		})
	}
}
