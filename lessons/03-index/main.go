// Lesson 3: indexing. Reads every .md/.txt file (point -data at your own
// notes to study anything), chunks it, embeds the
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
	"time"

	"github.com/raj-khan/rag-xray/internal/ai"
	"github.com/raj-khan/rag-xray/internal/retrieve"
)

func main() {
	dir := flag.String("data", "data/docs", "folder of .md or .txt files")
	out := flag.String("out", "store/index.json", "where to save the index")
	maxChars := flag.Int("max", 500, "max characters per chunk")
	flag.Parse()

	// 1. Load documents.
	docs, err := retrieve.LoadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}
	e, err := ai.NewEmbedder()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Chunk and embed (see retrieve.Build and retrieve.DocText).
	start := time.Now()
	s, err := retrieve.Build(docs, *maxChars, e)
	if err != nil {
		log.Fatal(err)
	}
	perFile := map[string]int{}
	for _, c := range s.Chunks {
		perFile[c.Source]++
	}
	for _, d := range docs {
		fmt.Printf("%-24s %3d chunks\n", d.Name, perFile[d.Name])
	}
	fmt.Printf("\nEmbedded %d chunks with %s in %s (%d dims each)\n",
		len(s.Chunks), e.Model(), time.Since(start).Round(time.Millisecond), len(s.Chunks[0].Vector))

	// 3. Save.
	if err := s.Save(*out); err != nil {
		log.Fatal(err)
	}
	info, _ := os.Stat(*out)
	fmt.Printf("Saved %s (%.1f KB)\n", *out, float64(info.Size())/1024)
	fmt.Println("\nExample of what was embedded for chunk 0:\n" + retrieve.DocText(e.Model(), s.Chunks[0]))
}
