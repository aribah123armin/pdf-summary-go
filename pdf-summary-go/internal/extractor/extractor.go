package extractor

import (
	"fmt"
	"strings"

	"github.com/dslipak/pdf"
)

// PageText holds the extracted text and page number.
type PageText struct {
	PageNum int
	Text    string
}

// Result holds full extraction output.
type Result struct {
	Pages     []PageText
	FullText  string
	PageCount int
	FilePath  string
}

// ExtractFromFile reads a PDF and returns extracted text per page.
func ExtractFromFile(filePath string) (*Result, error) {
	r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	result := &Result{
		FilePath:  filePath,
		PageCount: r.NumPage(),
	}

	var sb strings.Builder

	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		content, err := page.GetPlainText(nil)
		if err != nil {
			// Non-fatal: skip the page but continue
			content = fmt.Sprintf("[Could not extract text from page %d]\n", i)
		}

		pt := PageText{
			PageNum: i,
			Text:    strings.TrimSpace(content),
		}
		result.Pages = append(result.Pages, pt)
		sb.WriteString(fmt.Sprintf("--- Page %d ---\n%s\n\n", i, pt.Text))
	}

	result.FullText = sb.String()
	return result, nil
}

// Truncate limits text to maxChars to stay within API token limits.
func Truncate(text string, maxChars int) (string, bool) {
	if len(text) <= maxChars {
		return text, false
	}
	return text[:maxChars] + "\n\n[...text truncated due to length...]", true
}
