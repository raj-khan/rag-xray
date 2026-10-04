// Lesson 2: chunking. Compares naive fixed-size windows with
// structure-aware Markdown chunks on the same document.
//
//	go run ./lessons/02-chunking [file] [-size 300] [-overlap 50] [-max 500]
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/raj-khan/rag-xray/internal/chunk"
)

func main() {
	size := flag.Int("size", 300, "fixed chunk size in characters")
	overlap := flag.Int("overlap", 50, "fixed chunk overlap in characters")
	maxChars := flag.Int("max", 500, "max characters per markdown chunk")
	flag.Parse()

	path := "data/docs/kestrel-api.md"
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	text := string(b)

	show("FIXED", chunk.Fixed(text, *size, *overlap))
	show("MARKDOWN", chunk.Markdown(text, *maxChars))

	fmt.Println("Look at where FIXED chunks start and end: mid-sentence, mid-list,")
	fmt.Println("and a chunk can lose the heading that says what it is about.")
}

func show(name string, chunks []chunk.Chunk) {
	fmt.Printf("==================== %s: %d chunks ====================\n", name, len(chunks))
	for i, c := range chunks {
		fmt.Printf("--- #%d (%d chars)", i, len(c.Text))
		if c.Heading != "" {
			fmt.Printf("  [%s]", c.Heading)
		}
		fmt.Println()
		fmt.Println(indent(c.Text))
	}
	fmt.Println()
}

func indent(s string) string {
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
}
