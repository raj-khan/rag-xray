package ai

import (
	"bufio"
	"encoding/json"
	"errors"
	"strings"
)

// Anthropic calls the Claude Messages API. Anthropic has no embedding
// endpoint, so pair it with an Ollama or OpenAI-compatible embedder.
type Anthropic struct {
	BaseURL   string
	APIKey    string
	ChatModel string
}

func (a *Anthropic) Model() string { return a.ChatModel }

func (a *Anthropic) Chat(msgs []Message, onToken func(string)) (string, error) {
	// Claude takes the system prompt as a separate field.
	var system string
	var turns []Message
	for _, m := range msgs {
		if m.Role == "system" {
			system += m.Content + "\n"
			continue
		}
		turns = append(turns, m)
	}
	req := map[string]any{
		"model":       a.ChatModel,
		"max_tokens":  1024,
		"messages":    turns,
		"stream":      true,
		"temperature": 0,
	}
	if system != "" {
		req["system"] = strings.TrimSpace(system)
	}
	body, err := post(a.BaseURL+"/v1/messages", map[string]string{
		"x-api-key":         a.APIKey,
		"anthropic-version": "2023-06-01",
	}, req)
	if err != nil {
		return "", err
	}
	defer body.Close()

	// Server-sent events; text arrives in content_block_delta events.
	var sb strings.Builder
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data: ")
		if !ok {
			continue
		}
		var ev struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			return sb.String(), err
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Type == "text_delta" {
				sb.WriteString(ev.Delta.Text)
				if onToken != nil {
					onToken(ev.Delta.Text)
				}
			}
		case "error":
			return sb.String(), errors.New(ev.Error.Message)
		case "message_stop":
			return sb.String(), nil
		}
	}
	return sb.String(), sc.Err()
}
