package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pdf-summary",
	Short: "PDF to Text Summary Generator",
	Long: `pdf-summary extracts text from PDF files and generates
AI-powered summaries using the Anthropic Claude API.

Examples:
  pdf-summary summarize document.pdf
  pdf-summary summarize document.pdf --style bullet
  pdf-summary summarize document.pdf --length short --output summary.txt
  pdf-summary extract document.pdf`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
	rootCmd.AddCommand(extractCmd)
}
