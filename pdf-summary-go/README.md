# pdf-summary — PDF to Text Summary Generator (Go)

A command-line tool written in Go that extracts text from PDF files and generates AI-powered summaries.

---

## Features

- 📄 **Text extraction** from PDF files (text-based PDFs)
- 🤖 **AI summarization** generates short summary
- 🎨 **Three summary styles**: paragraph, bullet points, detailed
- 📏 **Three summary lengths**: short, medium, long
- 💾 **Save output** to a text file
- 🔍 **Extract-only mode** to dump raw PDF text

---

## Prerequisites

- Go 1.21 or later → https://go.dev/dl/
- An Anthropic API key → https://console.anthropic.com/

---

## Setup

```bash
# 1. Clone or download this project, then enter the directory
cd pdf-summary-go

# 2. Download dependencies
go mod tidy

# 3. Build the binary
go build -o pdf-summary .

# 4. Set your API key
export ANTHROPIC_API_KEY=sk-ant-your-key-here
```

---

## Usage

### Summarize a PDF

```bash
# Basic summarization (medium length, paragraph style)
./pdf-summary summarize report.pdf

# Bullet-point summary, short length
./pdf-summary summarize report.pdf --style bullet --length short

# Detailed long summary saved to a file
./pdf-summary summarize report.pdf --style detailed --length long --output summary.txt

# Pass API key inline (instead of env var)
./pdf-summary summarize report.pdf --api-key sk-ant-...
```

### Extract raw text only (no AI)

```bash
# Extract all pages
./pdf-summary extract document.pdf

# Extract a single page
./pdf-summary extract document.pdf --page 3

# Save extracted text to a file
./pdf-summary extract document.pdf --output text.txt
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--style` | `-s` | `paragraph` | `paragraph` \| `bullet` \| `detailed` |
| `--length` | `-l` | `medium` | `short` \| `medium` \| `long` |
| `--output` | `-o` | _(none)_ | File path to save the result |
| `--api-key` | | _(env)_ | Anthropic API key override |
| `--page` | `-p` | _(all)_ | Extract a single page (extract cmd) |

---

## Project Structure

```
pdf-summary-go/
├── main.go                     # Entry point
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go                 # Root Cobra command
│   ├── summarize.go            # `summarize` subcommand
│   └── extract.go              # `extract` subcommand
└── internal/
    ├── extractor/
    │   └── extractor.go        # PDF text extraction (dslipak/pdf)
    └── anthropic/
        └── client.go           # Anthropic API client
```

---

## Notes

- Only **text-based PDFs** are supported. Scanned/image PDFs require OCR (e.g. Tesseract) and are not handled.
- Very large PDFs are automatically truncated to ~80,000 characters to respect API token limits.
- The tool uses `claude-sonnet-4-20250514` by default.

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/dslipak/pdf` | Pure-Go PDF text extraction |
| `github.com/spf13/cobra` | CLI framework |

---

## License

MIT
