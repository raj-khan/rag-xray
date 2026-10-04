// Lesson 3: indexing. Reads every Markdown file, chunks it, embeds the
// chunks and saves them to a JSON vector store. This is the offline half
// of RAG; you rerun it whenever documents change.
//
//	go run ./lessons/03-index [-data data/docs] [-out store/index.json] [-max 500]
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/raj-khan/rag-from-scratch/internal/chunk"
	"github.com/raj-khan/rag-from-scratch/internal/ollama"
	"github.com/raj-khan/rag-from-scratch/internal/retrieve"
	"github.com/raj-khan/rag-from-scratch/internal/store"
)

func main() {
	dir := flag.String("data", "data/docs", "folder of .md files")
	out := flag.String("out", "store/index.json", "where to save the index")
	maxChars := flag.Int("max", 500, "max characters per chunk")
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*dir, "*.md"))
	if err != nil || len(files) == 0 {
		log.Fatalf("no .md files in %s", *dir)
	}

	// 1. Load and chunk.
	s := &store.Store{EmbedModel: ollama.EmbedModel}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		parts := chunk.Markdown(string(b), *maxChars)
		for _, p := range parts {
			s.Chunks = append(s.Chunks, store.Chunk{
				ID:      len(s.Chunks),
				Source:  filepath.Base(f),
				Heading: p.Heading,
				Text:    p.Text,
			})
		}
		fmt.Printf("%-20s %2d chunks\n", filepath.Base(f), len(parts))
	}

	// 2. Embed every chunk (with its context, see retrieve.DocText).
	texts := make([]string, len(s.Chunks))
	for i, c := range s.Chunks {
		texts[i] = retrieve.DocText(c)
	}
	start := time.Now()
	vecs, err := ollama.New().Embed(texts)
	if err != nil {
		log.Fatal(err)
	}
	for i := range s.Chunks {
		s.Chunks[i].Vector = vecs[i]
	}
	fmt.Printf("\nEmbedded %d chunks in %s (%d dims each)\n", len(vecs), time.Since(start).Round(time.Millisecond), len(vecs[0]))

	// 3. Save.
	if err := s.Save(*out); err != nil {
		log.Fatal(err)
	}
	info, _ := os.Stat(*out)
	fmt.Printf("Saved %s (%.1f KB)\n", *out, float64(info.Size())/1024)
	fmt.Println("\nExample of what was embedded for chunk 0:\n" + texts[0])
}
