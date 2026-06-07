package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	apiURL    = "https://api.anthropic.com/v1/messages"
	model     = "claude-sonnet-4-20250514"
	apiVersion = "2023-06-01"
	maxTokens = 1024
)

// SummaryStyle controls how the summary is formatted.
type SummaryStyle string

const (
	StyleParagraph SummaryStyle = "paragraph"
	StyleBullet    SummaryStyle = "bullet"
	StyleDetailed  SummaryStyle = "detailed"
)

// SummaryLength controls how long the summary is.
type SummaryLength string

const (
	LengthShort  SummaryLength = "short"
	LengthMedium SummaryLength = "medium"
	LengthLong   SummaryLength = "long"
)

// Message represents a chat message in the API request.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is the payload sent to the Anthropic API.
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system"`
}

// ContentBlock is a block in the API response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Response is the API response structure.
type Response struct {
	Content []ContentBlock `json:"content"`
	Error   *APIError      `json:"error,omitempty"`
}

// APIError represents an error from the Anthropic API.
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Client wraps the Anthropic HTTP API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Anthropic API client.
// It reads ANTHROPIC_API_KEY from the environment if apiKey is empty.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not found.\n" +
			"Set the ANTHROPIC_API_KEY environment variable:\n" +
			"  export ANTHROPIC_API_KEY=sk-ant-...")
	}
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// Summarize sends extracted PDF text to Claude and returns a summary.
func (c *Client) Summarize(text string, style SummaryStyle, length SummaryLength) (string, error) {
	systemPrompt := buildSystemPrompt(style, length)
	userPrompt := fmt.Sprintf("Please summarize the following document:\n\n%s", text)

	reqBody := Request{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiResp Response
		if jsonErr := json.Unmarshal(respBytes, &apiResp); jsonErr == nil && apiResp.Error != nil {
			return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, apiResp.Error.Message)
		}
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var apiResp Response
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("API returned empty content")
	}

	var result string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	return result, nil
}

func buildSystemPrompt(style SummaryStyle, length SummaryLength) string {
	lengthDesc := map[SummaryLength]string{
		LengthShort:  "a concise summary in 2-3 sentences",
		LengthMedium: "a moderate summary covering all key points in a few paragraphs",
		LengthLong:   "a comprehensive and detailed summary",
	}[length]

	styleDesc := map[SummaryStyle]string{
		StyleParagraph: "Write it as well-structured prose paragraphs.",
		StyleBullet:    "Format it as clear bullet points for easy scanning.",
		StyleDetailed:  "Include key sections, main arguments, important data, and conclusions.",
	}[style]

	return fmt.Sprintf(
		"You are an expert document summarizer. Your task is to produce %s of the provided document text. "+
			"%s "+
			"Be accurate, informative, and preserve the core meaning of the original content. "+
			"Do not add information not present in the document.",
		lengthDesc, styleDesc,
	)
}
