// Package retrieve finds the chunks most relevant to a question using
// vector search, keyword search (BM25), or both fused together (hybrid).
package retrieve

import (
	"fmt"
	"sort"

	"github.com/raj-khan/rag-from-scratch/internal/bm25"
	"github.com/raj-khan/rag-from-scratch/internal/ollama"
	"github.com/raj-khan/rag-from-scratch/internal/store"
	"github.com/raj-khan/rag-from-scratch/internal/vec"
)

// nomic-embed-text was trained with these task prefixes. Using them
// noticeably improves retrieval; try removing them in lesson 6.
const (
	DocPrefix   = "search_document: "
	QueryPrefix = "search_query: "
)

// DocText is what gets embedded for a chunk. Adding the file name and
// heading path ("contextual chunking") helps short chunks that would
// otherwise be ambiguous, e.g. a bare "Limit: 600 per minute".
func DocText(c store.Chunk) string {
	return fmt.Sprintf("%sDocument: %s\nSection: %s\n\n%s", DocPrefix, c.Source, c.Heading, c.Text)
}

type Hit struct {
	Chunk *store.Chunk
	Score float64
}

type Retriever struct {
	Store *store.Store
	BM25  *bm25.Index
	LLM   *ollama.Client
}

func Open(path string) (*Retriever, error) {
	s, err := store.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load index (run lesson 03 first): %w", err)
	}
	texts := make([]string, len(s.Chunks))
	for i, c := range s.Chunks {
		texts[i] = c.Source + " " + c.Heading + " " + c.Text
	}
	return &Retriever{Store: s, BM25: bm25.New(texts), LLM: ollama.New()}, nil
}

// Search runs one of the modes: "vector", "keyword" or "hybrid".
func (r *Retriever) Search(query, mode string, k int) ([]Hit, error) {
	switch mode {
	case "vector":
		return r.Vector(query, k)
	case "keyword":
		return r.Keyword(query, k), nil
	case "hybrid":
		v, err := r.Vector(query, 20)
		if err != nil {
			return nil, err
		}
		return top(RRF(v, r.Keyword(query, 20)), k), nil
	}
	return nil, fmt.Errorf("unknown mode %q (vector, keyword, hybrid)", mode)
}

func (r *Retriever) Vector(query string, k int) ([]Hit, error) {
	qv, err := r.LLM.Embed([]string{QueryPrefix + query})
	if err != nil {
		return nil, err
	}
	hits := make([]Hit, len(r.Store.Chunks))
	for i := range r.Store.Chunks {
		c := &r.Store.Chunks[i]
		hits[i] = Hit{c, vec.Cosine(qv[0], c.Vector)}
	}
	return top(hits, k), nil
}

func (r *Retriever) Keyword(query string, k int) []Hit {
	var hits []Hit
	for i, s := range r.BM25.Scores(query) {
		if s > 0 {
			hits = append(hits, Hit{&r.Store.Chunks[i], s})
		}
	}
	return top(hits, k)
}

// RRF (Reciprocal Rank Fusion) merges ranked lists using only ranks, so it
// does not care that cosine and BM25 scores live on different scales.
// Each list adds 1/(60+rank) to a chunk's score.
func RRF(lists ...[]Hit) []Hit {
	const c = 60
	scores := map[int]float64{}
	chunks := map[int]*store.Chunk{}
	for _, list := range lists {
		for rank, h := range list {
			scores[h.Chunk.ID] += 1 / float64(c+rank+1)
			chunks[h.Chunk.ID] = h.Chunk
		}
	}
	hits := make([]Hit, 0, len(scores))
	for id, s := range scores {
		hits = append(hits, Hit{chunks[id], s})
	}
	return hits
}

func top(hits []Hit, k int) []Hit {
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Chunk.ID < hits[j].Chunk.ID
	})
	if len(hits) > k {
		hits = hits[:k]
	}
	return hits
}
