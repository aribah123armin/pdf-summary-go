package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"pdf-summary/internal/anthropic"
	"pdf-summary/internal/extractor"
)

const maxTextChars = 80000 // ~20k tokens, safe for the API

var (
	flagStyle  string
	flagLength string
	flagOutput string
	flagAPIKey string
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize <file.pdf>",
	Short: "Extract text from a PDF and generate an AI summary",
	Args:  cobra.ExactArgs(1),
	RunE:  runSummarize,
	Example: `  pdf-summary summarize report.pdf
  pdf-summary summarize report.pdf --style bullet --length short
  pdf-summary summarize report.pdf --output summary.txt
  pdf-summary summarize report.pdf --api-key sk-ant-...`,
}

func init() {
	summarizeCmd.Flags().StringVarP(&flagStyle, "style", "s", "paragraph",
		"Summary style: paragraph | bullet | detailed")
	summarizeCmd.Flags().StringVarP(&flagLength, "length", "l", "medium",
		"Summary length: short | medium | long")
	summarizeCmd.Flags().StringVarP(&flagOutput, "output", "o", "",
		"Save summary to a file (optional)")
	summarizeCmd.Flags().StringVar(&flagAPIKey, "api-key", "",
		"Anthropic API key (defaults to ANTHROPIC_API_KEY env var)")
}

func runSummarize(cmd *cobra.Command, args []string) error {
	pdfPath := args[0]

	// ── Validate flags ──────────────────────────────────────────────────────
	style := anthropic.SummaryStyle(flagStyle)
	switch style {
	case anthropic.StyleParagraph, anthropic.StyleBullet, anthropic.StyleDetailed:
	default:
		return fmt.Errorf("invalid style %q — must be: paragraph, bullet, detailed", flagStyle)
	}

	length := anthropic.SummaryLength(flagLength)
	switch length {
	case anthropic.LengthShort, anthropic.LengthMedium, anthropic.LengthLong:
	default:
		return fmt.Errorf("invalid length %q — must be: short, medium, long", flagLength)
	}

	// ── Extract text ────────────────────────────────────────────────────────
	fmt.Printf("📄 Extracting text from: %s\n", pdfPath)
	start := time.Now()

	result, err := extractor.ExtractFromFile(pdfPath)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	if strings.TrimSpace(result.FullText) == "" {
		return fmt.Errorf("no text could be extracted from the PDF (it may be scanned/image-based)")
	}

	fmt.Printf("✅ Extracted %d pages (%d characters) in %.2fs\n",
		result.PageCount, len(result.FullText), time.Since(start).Seconds())

	// ── Truncate if needed ──────────────────────────────────────────────────
	text, wasTruncated := extractor.Truncate(result.FullText, maxTextChars)
	if wasTruncated {
		fmt.Printf("⚠️  Text truncated to %d characters to fit API limits\n", maxTextChars)
	}

	// ── Summarize via Claude ────────────────────────────────────────────────
	client, err := anthropic.NewClient(flagAPIKey)
	if err != nil {
		return err
	}

	fmt.Printf("🤖 Generating %s %s summary...\n", flagLength, flagStyle)
	aiStart := time.Now()

	summary, err := client.Summarize(text, style, length)
	if err != nil {
		return fmt.Errorf("summarization failed: %w", err)
	}

	fmt.Printf("✅ Summary generated in %.2fs\n\n", time.Since(aiStart).Seconds())

	// ── Output ──────────────────────────────────────────────────────────────
	header := buildHeader(pdfPath, result, flagStyle, flagLength)
	output := header + summary + "\n"

	fmt.Println(strings.Repeat("═", 60))
	fmt.Println(output)
	fmt.Println(strings.Repeat("═", 60))

	if flagOutput != "" {
		if err := os.WriteFile(flagOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("💾 Summary saved to: %s\n", flagOutput)
	}

	return nil
}

func buildHeader(pdfPath string, result *extractor.Result, style, length string) string {
	return fmt.Sprintf(
		"SUMMARY\n"+
			"File   : %s\n"+
			"Pages  : %d\n"+
			"Style  : %s\n"+
			"Length : %s\n"+
			"Date   : %s\n\n"+
			"%s\n\n",
		filepath.Base(pdfPath),
		result.PageCount,
		style,
		length,
		time.Now().Format("2006-01-02 15:04:05"),
		strings.Repeat("─", 60),
	)
}
