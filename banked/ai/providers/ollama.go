// Copyright (c) 2025 Youzill
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaProvider struct {
	HTTPClient *http.Client
	BaseURL    string
	Model      string
}

// NewOllamaProvider 初始化 Ollama 服务
// 默认地址为 http://127.0.0.1:11434
func NewOllamaProvider(baseURL, model string, timeout time.Duration) *OllamaProvider {
	if baseURL == "" {
		panic("Ollama baseURL is empty")
	}
	return &OllamaProvider{
		HTTPClient: &http.Client{Timeout: timeout},
		BaseURL:    baseURL,
		Model:      model,
	}
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  map[string]any  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

// CompleteText 通过 Ollama Chat API 请求模型仅返回文本
func (p *OllamaProvider) CompleteText(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaChatRequest{
		Model: p.Model,
		Messages: []ollamaMessage{
			{Role: "system", Content: "You output ONLY the song title text. No extra words, no quotes."},
			{Role: "user", Content: prompt},
		},
		Stream:  false,
		Options: map[string]any{"temperature": 0.2},
	}
	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/chat", p.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama http %d: %s", resp.StatusCode, string(body))
	}

	var cr ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	content := cr.Message.Content
	return content, nil
}
