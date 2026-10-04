// Lesson 4: the full RAG loop. Retrieve the best chunks for a question,
// put them in the prompt, and let the LLM answer with citations.
//
//	go run ./lessons/04-ask "How many vacation days do I get?"
//	go run ./lessons/04-ask -no-rag "How many vacation days do I get?"
//	go run ./lessons/04-ask -show-prompt -k 2 "What does KST-503 mean?"
//	go run ./lessons/04-ask              (interactive)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/raj-khan/rag-from-scratch/internal/ai"
	"github.com/raj-khan/rag-from-scratch/internal/retrieve"
)

const systemPrompt = `You are a helpful assistant for Orbita Labs employees.
Answer the question using ONLY the numbered context passages.
Cite the passages you used like [1] or [2].
If the context does not contain the answer, say "I don't know based on the documents."
Keep answers short.`

var (
	k          = flag.Int("k", 4, "number of chunks to retrieve")
	mode       = flag.String("mode", "hybrid", "retrieval mode: vector, keyword or hybrid")
	noRAG      = flag.Bool("no-rag", false, "ask the LLM without any retrieved context")
	showPrompt = flag.Bool("show-prompt", false, "print the full prompt sent to the LLM")
	index      = flag.String("index", "store/index.json", "index built by lesson 03")
)

func main() {
	flag.Parse()
	r, err := retrieve.Open(*index)
	if err != nil {
		log.Fatal(err)
	}

	if flag.NArg() > 0 {
		ask(r, strings.Join(flag.Args(), " "))
		return
	}
	in := bufio.NewScanner(os.Stdin)
	for fmt.Print("\nquestion> "); in.Scan(); fmt.Print("\nquestion> ") {
		if q := strings.TrimSpace(in.Text()); q != "" {
			ask(r, q)
		}
	}
}

func ask(r *retrieve.Retriever, question string) {
	if *noRAG {
		fmt.Println("(no retrieval: the model only has its training data)")
		chat(r.LLM, []ai.Message{{Role: "user", Content: question}})
		return
	}

	// 1. Retrieve.
	hits, err := r.Search(question, *mode, *k)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Augment: build a prompt with numbered sources.
	var ctx strings.Builder
	for i, h := range hits {
		fmt.Fprintf(&ctx, "[%d] (%s > %s)\n%s\n\n", i+1, h.Chunk.Source, h.Chunk.Heading, h.Chunk.Text)
	}
	user := fmt.Sprintf("Context:\n%s\nQuestion: %s", ctx.String(), question)
	if *showPrompt {
		fmt.Printf("----- SYSTEM -----\n%s\n----- USER -----\n%s\n------------------\n", systemPrompt, user)
	}

	// 3. Generate.
	chat(r.LLM, []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: user},
	})

	fmt.Println("\nSources:")
	for i, h := range hits {
		fmt.Printf("  [%d] %.4f  %s > %s\n", i+1, h.Score, h.Chunk.Source, h.Chunk.Heading)
	}
}

func chat(c ai.Chatter, msgs []ai.Message) {
	fmt.Println()
	if _, err := c.Chat(msgs, func(t string) { fmt.Print(t) }); err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}
