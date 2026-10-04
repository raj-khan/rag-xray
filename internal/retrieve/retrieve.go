// Package retrieve finds the chunks most relevant to a question using
// vector search, keyword search (BM25), or both fused together (hybrid).
package retrieve

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/raj-khan/rag-xray/internal/ai"
	"github.com/raj-khan/rag-xray/internal/bm25"
	"github.com/raj-khan/rag-xray/internal/chunk"
	"github.com/raj-khan/rag-xray/internal/store"
	"github.com/raj-khan/rag-xray/internal/vec"
)

// Some embedding models were trained with task prefixes and retrieve
// noticeably better when you use them. nomic-embed-text is one; most
// hosted models (e.g. OpenAI) need none.
func prefixes(model string) (doc, query string) {
	if strings.Contains(model, "nomic") {
		return "search_document: ", "search_query: "
	}
	return "", ""
}

// DocText is what gets embedded for a chunk. Adding the file name and
// heading path ("contextual chunking") helps short chunks that would
// otherwise be ambiguous, e.g. a bare "Limit: 600 per minute".
func DocText(model string, c store.Chunk) string {
	p, _ := prefixes(model)
	return fmt.Sprintf("%sDocument: %s\nSection: %s\n\n%s", p, c.Source, c.Heading, c.Text)
}

// Doc is one input document.
type Doc struct {
	Name string
	Text string
}

// LoadDir reads every .md and .txt file in dir.
func LoadDir(dir string) ([]Doc, error) {
	var docs []Doc
	for _, pattern := range []string{"*.md", "*.markdown", "*.txt"} {
		files, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			docs = append(docs, Doc{filepath.Base(f), string(b)})
		}
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no .md or .txt files in %s", dir)
	}
	return docs, nil
}

// Build chunks and embeds documents into a new store: the offline half of RAG.
func Build(docs []Doc, maxChars int, e ai.Embedder) (*store.Store, error) {
	s := &store.Store{EmbedModel: e.Model()}
	for _, d := range docs {
		for _, p := range chunk.Markdown(d.Text, maxChars) {
			s.Chunks = append(s.Chunks, store.Chunk{
				ID: len(s.Chunks), Source: d.Name, Heading: p.Heading, Text: p.Text,
			})
		}
	}
	if len(s.Chunks) == 0 {
		return nil, fmt.Errorf("documents produced no chunks")
	}
	texts := make([]string, len(s.Chunks))
	for i, c := range s.Chunks {
		texts[i] = DocText(e.Model(), c)
	}
	vecs, err := e.Embed(texts)
	if err != nil {
		return nil, err
	}
	for i := range s.Chunks {
		s.Chunks[i].Vector = vecs[i]
	}
	return s, nil
}

type Hit struct {
	Chunk *store.Chunk
	Score float64
}

type Retriever struct {
	Store *store.Store
	BM25  *bm25.Index
	Embed ai.Embedder
	LLM   ai.Chatter
}

// New wraps a store for searching. Queries must be embedded with the same
// model as the chunks, otherwise the vectors are not comparable.
func New(s *store.Store, e ai.Embedder, c ai.Chatter) (*Retriever, error) {
	if s.EmbedModel != e.Model() {
		return nil, fmt.Errorf("index was built with embedding model %q but EMBED_MODEL is %q: rebuild the index (lesson 03)", s.EmbedModel, e.Model())
	}
	texts := make([]string, len(s.Chunks))
	for i, c := range s.Chunks {
		texts[i] = c.Source + " " + c.Heading + " " + c.Text
	}
	return &Retriever{Store: s, BM25: bm25.New(texts), Embed: e, LLM: c}, nil
}

// Open loads a saved index and connects the models configured in the environment.
func Open(path string) (*Retriever, error) {
	s, err := store.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load index (run lesson 03 first): %w", err)
	}
	e, err := ai.NewEmbedder()
	if err != nil {
		return nil, err
	}
	c, err := ai.NewChatter()
	if err != nil {
		return nil, err
	}
	return New(s, e, c)
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
	_, qp := prefixes(r.Embed.Model())
	qv, err := r.Embed.Embed([]string{qp + query})
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
