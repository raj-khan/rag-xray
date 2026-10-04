package ai

import (
	"fmt"
	"os"
	"strings"
)

// Settings come from environment variables so every lesson, the web app
// and Docker share one configuration. See .env.example for recipes.
//
//	EMBED_PROVIDER  ollama | openai          (default ollama)
//	EMBED_MODEL     default nomic-embed-text / text-embedding-3-small
//	EMBED_BASE_URL  default http://localhost:11434 / https://api.openai.com/v1
//	EMBED_API_KEY   falls back to OPENAI_API_KEY
//
//	CHAT_PROVIDER   ollama | openai | anthropic   (default ollama)
//	CHAT_MODEL      default llama3.2 / gpt-4o-mini / claude-haiku-4-5
//	CHAT_BASE_URL   default http://localhost:11434 / https://api.openai.com/v1 / https://api.anthropic.com
//	CHAT_API_KEY    falls back to OPENAI_API_KEY / ANTHROPIC_API_KEY
//
// OLLAMA_HOST is honoured as the Ollama default for both.

type defaults struct{ model, baseURL, keyEnv string }

var embedDefaults = map[string]defaults{
	"ollama": {"nomic-embed-text", ollamaHost(), ""},
	"openai": {"text-embedding-3-small", "https://api.openai.com/v1", "OPENAI_API_KEY"},
}

var chatDefaults = map[string]defaults{
	"ollama":    {"llama3.2", ollamaHost(), ""},
	"openai":    {"gpt-4o-mini", "https://api.openai.com/v1", "OPENAI_API_KEY"},
	"anthropic": {"claude-haiku-4-5", "https://api.anthropic.com", "ANTHROPIC_API_KEY"},
}

func ollamaHost() string {
	h := os.Getenv("OLLAMA_HOST")
	if h == "" {
		return "http://localhost:11434"
	}
	if !strings.HasPrefix(h, "http") {
		h = "http://" + h
	}
	return h
}

type settings struct{ provider, model, baseURL, apiKey string }

func load(prefix string, table map[string]defaults) (settings, error) {
	s := settings{provider: strings.ToLower(env(prefix+"_PROVIDER", "ollama"))}
	d, ok := table[s.provider]
	if !ok {
		names := make([]string, 0, len(table))
		for n := range table {
			names = append(names, n)
		}
		return s, fmt.Errorf("%s_PROVIDER=%q is not supported (use one of %v)", prefix, s.provider, names)
	}
	s.model = env(prefix+"_MODEL", d.model)
	s.baseURL = strings.TrimRight(env(prefix+"_BASE_URL", d.baseURL), "/")
	s.apiKey = os.Getenv(prefix + "_API_KEY")
	if s.apiKey == "" && d.keyEnv != "" {
		s.apiKey = os.Getenv(d.keyEnv)
	}
	return s, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// NewEmbedder builds the embedder described by the EMBED_* variables.
func NewEmbedder() (Embedder, error) {
	s, err := load("EMBED", embedDefaults)
	if err != nil {
		return nil, err
	}
	if s.provider == "openai" {
		return &OpenAI{BaseURL: s.baseURL, APIKey: s.apiKey, EmbedModel: s.model}, nil
	}
	return &Ollama{Host: s.baseURL, EmbedModel: s.model}, nil
}

// NewChatter builds the chat model described by the CHAT_* variables.
func NewChatter() (Chatter, error) {
	s, err := load("CHAT", chatDefaults)
	if err != nil {
		return nil, err
	}
	switch s.provider {
	case "openai":
		return &OpenAI{BaseURL: s.baseURL, APIKey: s.apiKey, ChatModel: s.model}, nil
	case "anthropic":
		if s.apiKey == "" {
			return nil, fmt.Errorf("CHAT_PROVIDER=anthropic needs CHAT_API_KEY or ANTHROPIC_API_KEY")
		}
		return &Anthropic{BaseURL: s.baseURL, APIKey: s.apiKey, ChatModel: s.model}, nil
	}
	return &Ollama{Host: s.baseURL, ChatModel: s.model}, nil
}
