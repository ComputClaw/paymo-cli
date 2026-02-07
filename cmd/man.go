package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// manCmd generates man pages
var manCmd = &cobra.Command{
	Use:   "man [output-dir]",
	Short: "Generate man pages",
	Long: `Generate man pages for all paymo commands.

By default, man pages are written to ./man/ directory.
You can specify a custom output directory as an argument.

Examples:
  paymo man                    # Generate to ./man/
  paymo man /usr/local/man/man1   # Generate to system man directory
  paymo man ~/.local/share/man/man1

After generating, you can view with:
  man ./man/paymo.1
  man ./man/paymo-time.1
  man ./man/paymo-time-start.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := "./man"
		if len(args) > 0 {
			outputDir = args[0]
		}

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		// Generate man pages
		header := &doc.GenManHeader{
			Title:   "PAYMO",
			Section: "1",
			Source:  "paymo-cli " + version,
			Manual:  "Paymo CLI Manual",
		}

		if err := doc.GenManTree(rootCmd, header, outputDir); err != nil {
			return fmt.Errorf("generating man pages: %w", err)
		}

		// Count generated files
		files, _ := filepath.Glob(filepath.Join(outputDir, "*.1"))
		
		fmt.Printf("✅ Generated %d man pages in %s\n", len(files), outputDir)
		fmt.Println("\nView with:")
		fmt.Printf("  man %s/paymo.1\n", outputDir)
		fmt.Printf("  man %s/paymo-time.1\n", outputDir)
		
		return nil
	},
}

// markdownCmd generates markdown documentation
var markdownCmd = &cobra.Command{
	Use:   "markdown [output-dir]",
	Short: "Generate markdown documentation",
	Long: `Generate markdown documentation for all paymo commands.

By default, docs are written to ./docs/ directory.

Examples:
  paymo markdown              # Generate to ./docs/
  paymo markdown ./wiki       # Generate to custom directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := "./docs"
		if len(args) > 0 {
			outputDir = args[0]
		}

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		// Generate markdown docs
		if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
			return fmt.Errorf("generating markdown docs: %w", err)
		}

		// Count generated files
		files, _ := filepath.Glob(filepath.Join(outputDir, "*.md"))
		
		fmt.Printf("✅ Generated %d markdown files in %s\n", len(files), outputDir)
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
	rootCmd.AddCommand(markdownCmd)
}