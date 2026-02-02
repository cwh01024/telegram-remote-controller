package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	// GeminiAPIEndpoint is the endpoint for Gemini API
	GeminiAPIEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
)

// Client is the Gemini API client
type Client struct {
	apiKey string
	client *http.Client
}

// NewClient creates a new Gemini client
func NewClient() *Client {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: GEMINI_API_KEY not set, summarization will be disabled")
	}
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Request represents the API request
type Request struct {
	Contents []Content `json:"contents"`
}

// Content represents content in the request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a part of content
type Part struct {
	Text string `json:"text"`
}

// Response represents the API response
type Response struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content CandidateContent `json:"content"`
}

// CandidateContent represents the content of a candidate
type CandidateContent struct {
	Parts []Part `json:"parts"`
}

// Summarize summarizes text using Gemini
func (c *Client) Summarize(text string, maxLength int) (string, error) {
	if c.apiKey == "" {
		// Return truncated original if no API key
		if len(text) > maxLength {
			return text[:maxLength] + "...", nil
		}
		return text, nil
	}

	prompt := fmt.Sprintf(`請將以下內容精簡摘要，保留重要資訊，最多 %d 字：

%s`, maxLength, text)

	return c.Generate(prompt)
}

// Generate generates text using Gemini
func (c *Client) Generate(prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	reqBody := Request{
		Contents: []Content{
			{
				Parts: []Part{{Text: prompt}},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", GeminiAPIEndpoint, c.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var geminiResp Response
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// IsAvailable checks if Gemini API is available
func (c *Client) IsAvailable() bool {
	return c.apiKey != ""
}
