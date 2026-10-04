package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// These tests fake each provider's HTTP API, so they run offline.

func TestOpenAIStreamAndEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Errorf("Authorization = %q", got)
		}
		switch r.URL.Path {
		case "/chat/completions":
			for _, tok := range []string{"Hel", "lo"} {
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", tok)
			}
			fmt.Fprint(w, "data: [DONE]\n\n")
		case "/embeddings":
			// Out of order on purpose: results must be sorted by index.
			fmt.Fprint(w, `{"data":[{"index":1,"embedding":[0,1]},{"index":0,"embedding":[1,0]}]}`)
		}
	}))
	defer srv.Close()

	o := &OpenAI{BaseURL: srv.URL, APIKey: "k", ChatModel: "m", EmbedModel: "e"}
	var streamed string
	got, err := o.Chat([]Message{{Role: "user", Content: "hi"}}, func(s string) { streamed += s })
	if err != nil || got != "Hello" || streamed != "Hello" {
		t.Fatalf("Chat = %q, streamed %q, err %v", got, streamed, err)
	}
	vecs, err := o.Embed([]string{"a", "b"})
	if err != nil || vecs[0][0] != 1 || vecs[1][1] != 1 {
		t.Fatalf("Embed = %v, err %v", vecs, err)
	}
}

func TestAnthropicStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		if req["system"] != "be brief" {
			t.Errorf("system prompt not moved to top-level field: %v", req["system"])
		}
		if r.Header.Get("x-api-key") != "k" {
			t.Errorf("missing api key header")
		}
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Hi\"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\" there\"}}\n\n")
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer srv.Close()

	a := &Anthropic{BaseURL: srv.URL, APIKey: "k", ChatModel: "m"}
	got, err := a.Chat([]Message{{Role: "system", Content: "be brief"}, {Role: "user", Content: "hi"}}, nil)
	if err != nil || got != "Hi there" {
		t.Fatalf("Chat = %q, err %v", got, err)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("CHAT_PROVIDER", "openai")
	t.Setenv("CHAT_BASE_URL", "http://localhost:1234/v1/")
	t.Setenv("CHAT_MODEL", "qwen")
	c, err := NewChatter()
	if err != nil {
		t.Fatal(err)
	}
	o := c.(*OpenAI)
	if o.BaseURL != "http://localhost:1234/v1" || o.Model() != "qwen" {
		t.Errorf("got %+v", o)
	}

	t.Setenv("CHAT_PROVIDER", "nope")
	if _, err := NewChatter(); err == nil {
		t.Error("unknown provider should fail")
	}
}
