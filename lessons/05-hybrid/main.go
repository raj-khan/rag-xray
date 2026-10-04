// Lesson 5: hybrid search. Shows vector search, BM25 keyword search and
// their RRF fusion side by side, so you can see where each one wins.
//
//	go run ./lessons/05-hybrid                    (built-in examples)
//	go run ./lessons/05-hybrid "your question"
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/raj-khan/rag-xray/internal/retrieve"
)

var examples = []string{
	"Vaultbox",                       // rare name: only keywords find it
	"KST-401",                        // exact code: keywords are sharper
	"I'm ill and can't come to work", // paraphrase: vectors win, keywords drag hybrid down
	"too many requests error",        // both agree: hybrid is confident
	"what happens if my integration gets throttled", // everything struggles: why rerankers exist
}

func main() {
	k := flag.Int("k", 3, "results per method")
	index := flag.String("index", "store/index.json", "index built by lesson 03")
	flag.Parse()

	r, err := retrieve.Open(*index)
	if err != nil {
		log.Fatal(err)
	}
	queries := examples
	if flag.NArg() > 0 {
		queries = []string{strings.Join(flag.Args(), " ")}
	}

	for _, q := range queries {
		fmt.Printf("\n=== %q ===\n", q)
		for _, mode := range []string{"vector", "keyword", "hybrid"} {
			hits, err := r.Search(q, mode, *k)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  %s\n", mode)
			if len(hits) == 0 {
				fmt.Println("    (no results: no query word appears in any chunk)")
			}
			for i, h := range hits {
				fmt.Printf("    %d. %8.4f  %s > %s\n", i+1, h.Score, h.Chunk.Source, h.Chunk.Heading)
			}
		}
	}
}
