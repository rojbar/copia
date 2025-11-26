/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/rojbar/copia/internal/parser"
	"github.com/rojbar/copia/internal/writer"
	"github.com/spf13/cobra"
)

var (
	outputDir string
	dryRun    bool
	verbose   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "copia [template-package-path]",
	Short: "A simple template parser and writer",
	Long: `A simple template parser and writer, use it to generate files from templates.

The template package should contain .tmpl files that will be processed and written
to the output directory. Each template can have an associated package_vars.json file in
the same directory that provides variables for template substitution.

Example:
  copia ./templates/code/storage/basic -o ./output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templatePath := args[0]

		if verbose {
			fmt.Printf("Parsing templates from: %s\n", templatePath)
		}

		p := parser.New()
		templateFolder, err := p.Parse(templatePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Found %d template(s)\n", len(templateFolder.Templates))
		}

		// Write the templates
		w := writer.New(
			writer.WithOutputDir(outputDir),
			writer.WithDryRun(dryRun),
			writer.WithVerbose(verbose),
		)

		if err := w.Write(templateFolder); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing templates: %v\n", err)
			os.Exit(1)
		}

		if !dryRun {
			fmt.Printf("Successfully generated files in: %s\n", outputDir)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define flags
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Output directory for generated files")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Preview what would be generated without writing files")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
