# Copia

A simple template parser and writer CLI tool for generating files from templates.

## Installation

### Install using `go install`

```bash
go install github.com/rojbar/copia@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Verify installation

```bash
copia --help
```

## Usage

```bash
copia [template-package-path] [flags]
```

### Flags

- `-o, --output string`: Output directory for generated files (default: `./output`)
- `-d, --dry-run`: Preview what would be generated without writing files
- `-v, --verbose`: Verbose output
- `-h, --help`: Help for copia

### Example

```bash
# Generate files from a template package
copia ./templates/my-project -o ./output --verbose

# Dry run to preview what would be generated
copia ./templates/my-project -o ./output --dry-run
```

## Template Package Structure

A template package is a directory containing:

- Template files (with any extension, commonly `.tmpl`)
- A `package_vars.json` file at the root with variables for template substitution

### Example

```
my-template/
├── package_vars.json
├── main.go.tmpl
└── internal/
    └── handler.go.tmpl
```

**package_vars.json:**
```json
{
  "projectName": "myapp",
  "version": "1.0.0"
}
```

**main.go.tmpl:**
```go
package main

const ProjectName = "{{.projectName}}"
const Version = "{{.version}}"
```

Running `copia my-template -o output` will generate the files with variables substituted.

## Development

### Build from source

```bash
git clone https://github.com/rojbar/copia.git
cd copia
go build
```

### Run tests

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) file for details.
