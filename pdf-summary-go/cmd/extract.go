package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pdf-summary/internal/extractor"
)

var (
	extractOutput   string
	extractPageOnly int
)

var extractCmd = &cobra.Command{
	Use:   "extract <file.pdf>",
	Short: "Extract raw text from a PDF without summarizing",
	Args:  cobra.ExactArgs(1),
	RunE:  runExtract,
	Example: `  pdf-summary extract document.pdf
  pdf-summary extract document.pdf --output extracted.txt
  pdf-summary extract document.pdf --page 3`,
}

func init() {
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "",
		"Save extracted text to a file (optional)")
	extractCmd.Flags().IntVarP(&extractPageOnly, "page", "p", 0,
		"Extract a specific page only (1-indexed)")
}

func runExtract(cmd *cobra.Command, args []string) error {
	pdfPath := args[0]

	fmt.Printf("📄 Extracting text from: %s\n", pdfPath)

	result, err := extractor.ExtractFromFile(pdfPath)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Printf("✅ Extracted %d pages (%d characters)\n\n",
		result.PageCount, len(result.FullText))

	var output string

	if extractPageOnly > 0 {
		if extractPageOnly > result.PageCount {
			return fmt.Errorf("page %d out of range (document has %d pages)", extractPageOnly, result.PageCount)
		}
		for _, p := range result.Pages {
			if p.PageNum == extractPageOnly {
				output = fmt.Sprintf("--- Page %d ---\n%s\n", p.PageNum, p.Text)
				break
			}
		}
	} else {
		output = result.FullText
	}

	fmt.Println(strings.Repeat("═", 60))
	fmt.Println(output)
	fmt.Println(strings.Repeat("═", 60))

	if extractOutput != "" {
		if err := os.WriteFile(extractOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("💾 Text saved to: %s\n", extractOutput)
	}

	return nil
}
