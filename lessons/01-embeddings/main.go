// Lesson 1: embeddings and similarity.
//
// This file deliberately uses raw HTTP and no project packages, so you can
// see exactly what an embedding call is.
//
//	go run ./lessons/01-embeddings "how do I make pasta"
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
)

var sentences = []string{
	"The cat curled up on the warm windowsill and fell asleep.",
	"Kittens need to eat several small meals a day.",
	"Goroutines are lightweight threads managed by the Go runtime.",
	"Use a channel to pass values safely between goroutines.",
	"Boil the spaghetti in salted water for about nine minutes.",
	"A good tomato sauce simmers slowly with garlic and olive oil.",
	"The stock market fell sharply after the interest rate decision.",
}

// embed calls Ollama's /api/embed and returns one vector per input.
func embed(inputs []string) [][]float64 {
	body, _ := json.Marshal(map[string]any{"model": "nomic-embed-text", "input": inputs})
	resp, err := http.Post("http://localhost:11434/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("is Ollama running? %v", err)
	}
	defer resp.Body.Close()
	var out struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		log.Fatal(err)
	}
	return out.Embeddings
}

func cosine(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func main() {
	query := "how do I make pasta"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}

	vecs := embed(append([]string{query}, sentences...))
	q, docs := vecs[0], vecs[1:]

	fmt.Printf("Each text became a vector of %d numbers. The query starts with:\n  %.3f\n\n", len(q), q[:6])
	fmt.Printf("Query: %q\n\n", query)

	type scored struct {
		text  string
		score float64
	}
	var ranked []scored
	for i, d := range docs {
		ranked = append(ranked, scored{sentences[i], cosine(q, d)})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })

	for _, r := range ranked {
		bar := strings.Repeat("#", int(math.Max(0, r.score)*40))
		fmt.Printf("%.3f %-28s %s\n", r.score, bar, r.text)
	}

	fmt.Println("\nNotice: unrelated sentences still score well above 0.")
	fmt.Println("Scores are only meaningful relative to each other, so rank, don't threshold.")
}
