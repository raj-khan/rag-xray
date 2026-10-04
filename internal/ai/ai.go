// Package ai talks to whichever model provider you configure: a local
// Ollama server (the default), any OpenAI-compatible API (OpenAI,
// OpenRouter, Groq, Gemini, LM Studio, llama.cpp, vLLM...) or Anthropic.
//
// RAG needs exactly two capabilities:
//   - an Embedder, which turns text into vectors (for search)
//   - a Chatter, which turns a prompt into an answer (for generation)
//
// They are configured separately, so you can embed locally for free and
// answer with a hosted model, or the other way around.
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Message struct {
	Role    string `json:"role"` // "system", "user" or "assistant"
	Content string `json:"content"`
}

type Embedder interface {
	// Embed returns one vector per input text, in order.
	Embed(texts []string) ([][]float64, error)
	Model() string
}

type Chatter interface {
	// Chat streams the reply to onToken (may be nil) and returns it whole.
	Chat(msgs []Message, onToken func(string)) (string, error)
	Model() string
}

const batchSize = 32

var httpClient = &http.Client{Timeout: 10 * time.Minute}

// post sends JSON and returns the response body for a 200 reply.
func post(url string, headers map[string]string, body any) (io.ReadCloser, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach %s (is the server running?): %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s: %s", url, resp.Status, bytes.TrimSpace(msg))
	}
	return resp.Body, nil
}

// batched calls fn on slices of at most batchSize texts and joins results.
func batched(texts []string, fn func([]string) ([][]float64, error)) ([][]float64, error) {
	out := make([][]float64, 0, len(texts))
	for start := 0; start < len(texts); start += batchSize {
		vecs, err := fn(texts[start:min(start+batchSize, len(texts))])
		if err != nil {
			return nil, err
		}
		out = append(out, vecs...)
	}
	if len(out) != len(texts) {
		return nil, fmt.Errorf("asked for %d embeddings, got %d", len(texts), len(out))
	}
	return out, nil
}
