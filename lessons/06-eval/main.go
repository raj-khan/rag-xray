// Lesson 6: evaluation. Measures retrieval quality on a labelled question
// set, and optionally checks the final answers too. Without numbers like
// these, every "improvement" to a RAG system is a guess.
//
//	go run ./lessons/06-eval            (retrieval only, fast)
//	go run ./lessons/06-eval -answers   (also generate and check answers)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/raj-khan/rag-xray/internal/ai"
	"github.com/raj-khan/rag-xray/internal/retrieve"
)

type example struct {
	Q      string `json:"q"`
	Source string `json:"source"` // file that holds the answer
	Answer string `json:"answer"` // text the answer must contain
}

func main() {
	k := flag.Int("k", 4, "chunks retrieved per question")
	answers := flag.Bool("answers", false, "also generate answers (slower)")
	index := flag.String("index", "store/index.json", "index built by lesson 03")
	set := flag.String("set", "data/eval.json", "labelled questions")
	flag.Parse()

	r, err := retrieve.Open(*index)
	if err != nil {
		log.Fatal(err)
	}
	b, err := os.ReadFile(*set)
	if err != nil {
		log.Fatal(err)
	}
	var examples []example
	if err := json.Unmarshal(b, &examples); err != nil {
		log.Fatal(err)
	}

	// Retrieval metrics:
	//   hit@1  = right file ranked first
	//   hit@k  = right file anywhere in the top k (what the LLM gets to see)
	//   MRR    = mean of 1/rank of the first right result (1.0 is perfect)
	fmt.Printf("%d questions, k=%d\n\n%-8s %6s %6s %6s\n", len(examples), *k, "mode", "hit@1", "hit@k", "MRR")
	for _, mode := range []string{"vector", "keyword", "hybrid"} {
		var hit1, hitK, mrr float64
		var misses []string
		for _, ex := range examples {
			hits, err := r.Search(ex.Q, mode, *k)
			if err != nil {
				log.Fatal(err)
			}
			rank := 0
			for i, h := range hits {
				if h.Chunk.Source == ex.Source {
					rank = i + 1
					break
				}
			}
			if rank == 1 {
				hit1++
			}
			if rank > 0 {
				hitK++
				mrr += 1 / float64(rank)
			} else {
				misses = append(misses, ex.Q)
			}
		}
		n := float64(len(examples))
		fmt.Printf("%-8s %6.2f %6.2f %6.2f\n", mode, hit1/n, hitK/n, mrr/n)
		for _, m := range misses {
			fmt.Printf("           miss: %s\n", m)
		}
	}

	if !*answers {
		fmt.Println("\nRun with -answers to also grade generated answers.")
		return
	}

	fmt.Println("\nAnswer check (hybrid retrieval, answer must contain the expected text):")
	correct := 0
	for _, ex := range examples {
		hits, err := r.Search(ex.Q, "hybrid", *k)
		if err != nil {
			log.Fatal(err)
		}
		var ctx strings.Builder
		for i, h := range hits {
			fmt.Fprintf(&ctx, "[%d] %s\n\n", i+1, h.Chunk.Text)
		}
		got, err := r.LLM.Chat([]ai.Message{
			{Role: "system", Content: "Answer using only the context. One short sentence."},
			{Role: "user", Content: "Context:\n" + ctx.String() + "Question: " + ex.Q},
		}, nil)
		if err != nil {
			log.Fatal(err)
		}
		ok := strings.Contains(strings.ToLower(got), strings.ToLower(ex.Answer))
		mark := "FAIL"
		if ok {
			correct++
			mark = "ok  "
		}
		fmt.Printf("  %s %s\n       -> %s\n", mark, ex.Q, strings.TrimSpace(strings.ReplaceAll(got, "\n", " ")))
	}
	fmt.Printf("\nAnswer accuracy: %d/%d\n", correct, len(examples))
}
