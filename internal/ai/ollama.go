package ai

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// Ollama runs models on your own machine. Install it from ollama.com, then:
//
//	ollama pull nomic-embed-text && ollama pull llama3.2
type Ollama struct {
	Host       string
	EmbedModel string
	ChatModel  string
}

func (o *Ollama) Model() string {
	if o.ChatModel != "" {
		return o.ChatModel
	}
	return o.EmbedModel
}

func (o *Ollama) Embed(texts []string) ([][]float64, error) {
	return batched(texts, func(batch []string) ([][]float64, error) {
		body, err := post(o.Host+"/api/embed", nil, map[string]any{
			"model": o.EmbedModel,
			"input": batch,
		})
		if err != nil {
			return nil, hintPull(err, o.EmbedModel)
		}
		defer body.Close()
		var out struct {
			Embeddings [][]float64 `json:"embeddings"`
		}
		return out.Embeddings, json.NewDecoder(body).Decode(&out)
	})
}

func (o *Ollama) Chat(msgs []Message, onToken func(string)) (string, error) {
	body, err := post(o.Host+"/api/chat", nil, map[string]any{
		"model":    o.ChatModel,
		"messages": msgs,
		"stream":   true,
		"options":  map[string]any{"temperature": 0},
	})
	if err != nil {
		return "", hintPull(err, o.ChatModel)
	}
	defer body.Close()

	// The stream is newline-delimited JSON, one object per few tokens.
	var sb strings.Builder
	dec := json.NewDecoder(body)
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

func hintPull(err error, model string) error {
	if strings.Contains(err.Error(), "not found") {
		return errors.New(err.Error() + "\nhint: run `ollama pull " + model + "`")
	}
	return err
}
