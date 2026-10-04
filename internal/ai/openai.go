package ai

import (
	"bufio"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

// OpenAI speaks the OpenAI REST format, which most providers and local
// servers also accept. Point BaseURL at any of them, for example:
//
//	https://api.openai.com/v1
//	https://openrouter.ai/api/v1
//	https://api.groq.com/openai/v1
//	https://generativelanguage.googleapis.com/v1beta/openai
//	http://localhost:1234/v1   (LM Studio)
//	http://localhost:8080/v1   (llama.cpp server)
//	http://localhost:11434/v1  (Ollama)
type OpenAI struct {
	BaseURL    string
	APIKey     string
	EmbedModel string
	ChatModel  string
}

func (o *OpenAI) Model() string {
	if o.ChatModel != "" {
		return o.ChatModel
	}
	return o.EmbedModel
}

func (o *OpenAI) headers() map[string]string {
	if o.APIKey == "" {
		return nil // local servers usually need no key
	}
	return map[string]string{"Authorization": "Bearer " + o.APIKey}
}

func (o *OpenAI) Embed(texts []string) ([][]float64, error) {
	return batched(texts, func(batch []string) ([][]float64, error) {
		body, err := post(o.BaseURL+"/embeddings", o.headers(), map[string]any{
			"model": o.EmbedModel,
			"input": batch,
		})
		if err != nil {
			return nil, err
		}
		defer body.Close()
		var out struct {
			Data []struct {
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			} `json:"data"`
		}
		if err := json.NewDecoder(body).Decode(&out); err != nil {
			return nil, err
		}
		sort.Slice(out.Data, func(i, j int) bool { return out.Data[i].Index < out.Data[j].Index })
		vecs := make([][]float64, len(out.Data))
		for i, d := range out.Data {
			vecs[i] = d.Embedding
		}
		return vecs, nil
	})
}

func (o *OpenAI) Chat(msgs []Message, onToken func(string)) (string, error) {
	body, err := post(o.BaseURL+"/chat/completions", o.headers(), map[string]any{
		"model":       o.ChatModel,
		"messages":    msgs,
		"stream":      true,
		"temperature": 0,
	})
	if err != nil {
		return "", err
	}
	defer body.Close()

	// Server-sent events: lines of "data: {json}", ending with "data: [DONE]".
	var sb strings.Builder
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data: ")
		if !ok {
			continue
		}
		if data == "[DONE]" {
			break
		}
		var ev struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			return sb.String(), err
		}
		if ev.Error != nil {
			return sb.String(), errors.New(ev.Error.Message)
		}
		if len(ev.Choices) > 0 && ev.Choices[0].Delta.Content != "" {
			t := ev.Choices[0].Delta.Content
			sb.WriteString(t)
			if onToken != nil {
				onToken(t)
			}
		}
	}
	return sb.String(), sc.Err()
}
