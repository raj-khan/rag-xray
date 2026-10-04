package retrieve

import (
	"fmt"
	"strings"
)

// SystemPrompt grounds the model, asks for citations and gives it
// permission to say it does not know. See lesson 4.
const SystemPrompt = `You are a helpful assistant answering questions about the user's documents.
Answer the question using ONLY the numbered context passages.
Cite the passages you used like [1] or [2].
If the context does not contain the answer, say "I don't know based on the documents."
Keep answers short.`

// Prompt builds the user message: numbered passages, then the question.
func Prompt(question string, hits []Hit) string {
	var b strings.Builder
	b.WriteString("Context:\n")
	for i, h := range hits {
		fmt.Fprintf(&b, "[%d] (%s > %s)\n%s\n\n", i+1, h.Chunk.Source, h.Chunk.Heading, h.Chunk.Text)
	}
	b.WriteString("Question: " + question)
	return b.String()
}
