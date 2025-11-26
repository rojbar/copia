package parser_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/rojbar/copia/internal/parser"
)

func TestPackageParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		options     []parser.Option
		want        *parser.TemplatePackage
		wantErr     bool
		errContains string
	}{
		{
			name: "simple_package_with_one_template",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{"name": "test", "version": "1.0"}`
				templateContent := `package {{.name}}`

				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "main.go.tmpl"), []byte(templateContent), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    "main.go.tmpl",
						Content: `package {{.name}}`,
					},
				},
				Variables: map[string]any{
					"name":    "test",
					"version": "1.0",
				},
			},
			options: []parser.Option{parser.WithVarsFileName("vars.json")},
			wantErr: false,
		},
		{
			name: "package_with_nested_directories",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{"project": "myapp"}`

				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}

				subDir := filepath.Join(dir, "internal", "handlers")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(dir, "main.go.tmpl"), []byte("package main"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(subDir, "handler.go.tmpl"), []byte("package handlers"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    filepath.Join("internal", "handlers", "handler.go.tmpl"),
						Content: "package handlers",
					},
					{
						Path:    "main.go.tmpl",
						Content: "package main",
					},
				},
				Variables: map[string]any{
					"project": "myapp",
				},
			},
			options: []parser.Option{parser.WithVarsFileName("vars.json")},
			wantErr: false,
		},
		{
			name: "package_with_max_depth_limit",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{}`

				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}

				deepDir := filepath.Join(dir, "level1", "level2", "level3")
				if err := os.MkdirAll(deepDir, 0755); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(dir, "root.tmpl"), []byte("root"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "level1", "file1.tmpl"), []byte("level1"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(deepDir, "deep.tmpl"), []byte("deep"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			options: []parser.Option{parser.WithMaxDepth(1), parser.WithVarsFileName("vars.json")},
			want:    nil,
			wantErr: true,

			errContains: "maximum directory depth exceeded",
		},
		{
			name: "package_with_ignore_patterns",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{}`

				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "main.go.tmpl"), []byte("main"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "test.backup"), []byte("backup"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "data.tmp"), []byte("tmp"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			options: []parser.Option{parser.WithIgnorePatterns("*.backup", "*.tmp"), parser.WithVarsFileName("vars.json")},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    "main.go.tmpl",
						Content: "main",
					},
				},
				Variables: map[string]any{},
			},
			wantErr: false,
		},
		{
			name: "package_with_custom_vars_file_name",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{"custom": "value"}`

				if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "main.go.tmpl"), []byte("main"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			options: []parser.Option{parser.WithVarsFileName("config.json")},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    "main.go.tmpl",
						Content: "main",
					},
				},
				Variables: map[string]any{
					"custom": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "package_without_vars_json",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "main.go.tmpl"), []byte("main"), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want:        nil,
			wantErr:     true,
			errContains: "failed to fetch variables",
		},
		{
			name: "empty_package",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{}`
				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{},
				Variables: map[string]any{},
			},
			wantErr: false,
			options: []parser.Option{parser.WithVarsFileName("vars.json")},
		},
		{
			name: "non_existent_directory",
			setupFunc: func(t *testing.T) string {
				return "/non/existent/path"
			},
			want:        nil,
			wantErr:     true,
			errContains: "failed to access package path",
		},
		{
			name: "path_is_a_file_not_directory",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				filePath := filepath.Join(dir, "notadir")
				if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
					t.Fatal(err)
				}
				return filePath
			},
			want:        nil,
			wantErr:     true,
			errContains: "package path must be a directory",
		},
		{
			name: "invalid_vars_json",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{invalid json}`
				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want:        nil,
			wantErr:     true,
			errContains: "failed to parse vars file",
			options:     []parser.Option{parser.WithVarsFileName("vars.json")},
		},
		{
			name: "vars_json_should_not_be_parsed_as_template",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{"key": "value"}`
				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "main.tmpl"), []byte("template"), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    "main.tmpl",
						Content: "template",
					},
				},
				Variables: map[string]any{
					"key": "value",
				},
			},
			wantErr: false,
			options: []parser.Option{parser.WithVarsFileName("vars.json")},
		},
		{
			name: "multiple_templates_with_shared_variables",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{"shared": "data"}`

				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "file1.tmpl"), []byte("template1"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "file2.tmpl"), []byte("template2"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    "file1.tmpl",
						Content: "template1",
					},
					{
						Path:    "file2.tmpl",
						Content: "template2",
					},
				},
				Variables: map[string]any{
					"shared": "data",
				},
			},
			wantErr: false,
			options: []parser.Option{parser.WithVarsFileName("vars.json")},
		},
		{
			name: "vars_json_in_subdirectory_should_be_included_as_template",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				varsContent := `{"root": "data"}`

				if err := os.WriteFile(filepath.Join(dir, "vars.json"), []byte(varsContent), 0644); err != nil {
					t.Fatal(err)
				}

				subDir := filepath.Join(dir, "subdir")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatal(err)
				}

				subVarsContent := `{"sub": "value"}`
				if err := os.WriteFile(filepath.Join(subDir, "vars.json"), []byte(subVarsContent), 0644); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(dir, "main.tmpl"), []byte("main"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: &parser.TemplatePackage{
				Templates: []parser.Template{
					{
						Path:    "main.tmpl",
						Content: "main",
					},
					{
						Path:    filepath.Join("subdir", "vars.json"),
						Content: `{"sub": "value"}`,
					},
				},
				Variables: map[string]any{
					"root": "data",
				},
			},
			wantErr: false,
			options: []parser.Option{parser.WithVarsFileName("vars.json")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packagePath := tt.setupFunc(t)

			p := parser.New(tt.options...)
			pkg, err := p.Parse(packagePath)

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

			if pkg == nil {
				t.Fatal("expected non-nil package")
			}

			// Set the Path field to match (since it's the temp dir path)
			if tt.want != nil {
				tt.want.Path = pkg.Path
			}

			if !compareTemplates(pkg.Templates, tt.want.Templates) {
				t.Errorf("templates mismatch:\ngot:  %+v\nwant: %+v", pkg.Templates, tt.want.Templates)
			}

			if !reflect.DeepEqual(pkg.Variables, tt.want.Variables) {
				t.Errorf("variables mismatch:\ngot:  %+v\nwant: %+v", pkg.Variables, tt.want.Variables)
			}

			if pkg.Path != tt.want.Path {
				t.Errorf("path mismatch: got %q, want %q", pkg.Path, tt.want.Path)
			}
		})
	}
}

func compareTemplates(got, want []parser.Template) bool {
	if len(got) != len(want) {
		return false
	}

	// Create maps for order-independent comparison
	gotMap := make(map[string]string)
	for _, tmpl := range got {
		gotMap[tmpl.Path] = tmpl.Content
	}

	wantMap := make(map[string]string)
	for _, tmpl := range want {
		wantMap[tmpl.Path] = tmpl.Content
	}

	return reflect.DeepEqual(gotMap, wantMap)
}

func TestPackageParser_Options(t *testing.T) {
	tests := []struct {
		name    string
		options []parser.Option
	}{
		{
			name:    "default_options",
			options: []parser.Option{},
		},
		{
			name: "with_max_depth",
			options: []parser.Option{
				parser.WithMaxDepth(5),
			},
		},
		{
			name: "with_ignore_patterns",
			options: []parser.Option{
				parser.WithIgnorePatterns("*.backup", "*.tmp"),
			},
		},
		{
			name: "with_custom_vars_file",
			options: []parser.Option{
				parser.WithVarsFileName("config.json"),
			},
		},
		{
			name: "combined_options",
			options: []parser.Option{
				parser.WithMaxDepth(3),
				parser.WithIgnorePatterns("*.bak"),
				parser.WithVarsFileName("vars.yaml"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.options...)
			if p == nil {
				t.Error("expected non-nil parser")
			}
		})
	}
}
