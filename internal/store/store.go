// Package store is the simplest possible vector store: a slice of chunks
// with their vectors, saved as one JSON file. Fine for thousands of chunks.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Chunk struct {
	ID      int       `json:"id"`
	Source  string    `json:"source"`  // file name
	Heading string    `json:"heading"` // heading path inside the file
	Text    string    `json:"text"`
	Vector  []float64 `json:"vector"`
}

type Store struct {
	EmbedModel string  `json:"embed_model"`
	Chunks     []Chunk `json:"chunks"`
}

func (s *Store) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Load(path string) (*Store, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Store
	return &s, json.Unmarshal(b, &s)
}
