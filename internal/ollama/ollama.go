// Package ollama is a tiny client for the two Ollama endpoints RAG needs:
// /api/embed (text -> vectors) and /api/chat (messages -> answer).
package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	EmbedModel = "nomic-embed-text" // 768-dim embeddings
	ChatModel  = "llama3.2"         // 3B chat model
	batchSize  = 32
)

type Client struct {
	Host string
	HTTP *http.Client
}

func New() *Client {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	return &Client{
		Host: strings.TrimRight(host, "/"),
		HTTP: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (c *Client) post(path string, body any) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Post(c.Host+path, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("is Ollama running at %s? %w", c.Host, err)
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama %s: %s: %s", path, resp.Status, msg)
	}
	return resp, nil
}

// Embed returns one vector per input, sending inputs in batches.
func (c *Client) Embed(inputs []string) ([][]float64, error) {
	out := make([][]float64, 0, len(inputs))
	for start := 0; start < len(inputs); start += batchSize {
		end := min(start+batchSize, len(inputs))
		resp, err := c.post("/api/embed", map[string]any{
			"model": EmbedModel,
			"input": inputs[start:end],
		})
		if err != nil {
			return nil, err
		}
		var body struct {
			Embeddings [][]float64 `json:"embeddings"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		out = append(out, body.Embeddings...)
	}
	return out, nil
}

type Message struct {
	Role    string `json:"role"` // "system", "user" or "assistant"
	Content string `json:"content"`
}

// Chat streams the model's reply, calling onToken for each piece (may be nil),
// and returns the full reply. Temperature 0 keeps answers repeatable.
func (c *Client) Chat(msgs []Message, onToken func(string)) (string, error) {
	resp, err := c.post("/api/chat", map[string]any{
		"model":    ChatModel,
		"messages": msgs,
		"stream":   true,
		"options":  map[string]any{"temperature": 0},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// The stream is newline-delimited JSON: one object per token batch.
	var sb strings.Builder
	dec := json.NewDecoder(resp.Body)
	for {
		var part struct {
			Message Message `json:"message"`
			Done    bool    `json:"done"`
			Error   string  `json:"error"`
		}
		if err := dec.Decode(&part); err == io.EOF {
			break
		} else if err != nil {
			return sb.String(), err
		}
		if part.Error != "" {
			return sb.String(), errors.New(part.Error)
		}
		sb.WriteString(part.Message.Content)
		if onToken != nil {
			onToken(part.Message.Content)
		}
		if part.Done {
			break
		}
	}
	return sb.String(), nil
}
