package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	// GeminiVisionEndpoint for image understanding
	GeminiVisionEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
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
		log.Println("Warning: GEMINI_API_KEY not set, OCR/summarization will be disabled")
	}
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// VisionRequest represents the API request with image
type VisionRequest struct {
	Contents []VisionContent `json:"contents"`
}

// VisionContent represents content in the request
type VisionContent struct {
	Parts []VisionPart `json:"parts"`
}

// VisionPart represents a part of content (text or image)
type VisionPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

// InlineData represents base64 encoded image data
type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
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

// Part represents a text part
type Part struct {
	Text string `json:"text"`
}

// ExtractTextFromImage reads an image and extracts text/response using Gemini Vision
func (c *Client) ExtractTextFromImage(imagePath string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	log.Printf("Extracting text from image: %s", imagePath)

	// Read and encode image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Build request
	reqBody := VisionRequest{
		Contents: []VisionContent{
			{
				Parts: []VisionPart{
					{
						Text: `這是一個 IDE (Antigravity) 的螢幕截圖。請提取 AI 助手的回應內容。
規則：
1. 只提取 AI 的回應文字，不要包含使用者的問題
2. 如果有代碼區塊，保留格式
3. 用簡潔的方式呈現，去除不必要的 UI 元素描述
4. 如果截圖中沒有明確的回應，說明「未偵測到回應內容」
5. 用繁體中文回覆（如果原本是中文的話）

請直接輸出提取的回應內容：`,
					},
					{
						InlineData: &InlineData{
							MimeType: "image/png",
							Data:     base64Image,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", GeminiVisionEndpoint, c.apiKey)
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
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp Response
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	extractedText := geminiResp.Candidates[0].Content.Parts[0].Text
	log.Printf("Extracted %d characters from image", len(extractedText))

	return extractedText, nil
}

// Summarize summarizes text using Gemini
func (c *Client) Summarize(text string, maxLength int) (string, error) {
	if c.apiKey == "" {
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

	reqBody := VisionRequest{
		Contents: []VisionContent{
			{
				Parts: []VisionPart{{Text: prompt}},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", GeminiVisionEndpoint, c.apiKey)
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
