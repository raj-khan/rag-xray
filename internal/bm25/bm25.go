// Package bm25 is classic keyword search: it rewards documents that contain
// the query's words, especially rare words, without needing any model.
package bm25

import (
	"math"
	"strings"
	"unicode"
)

const (
	k1 = 1.2  // how fast repeated words stop adding score
	b  = 0.75 // how much long documents are penalised
)

var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true, "of": true,
	"to": true, "in": true, "on": true, "for": true, "and": true, "or": true,
	"what": true, "how": true, "do": true, "does": true, "i": true, "my": true,
	"can": true, "be": true, "it": true, "with": true, "at": true, "by": true,
	"if": true, "we": true, "you": true, "your": true, "much": true, "many": true,
}

func Tokenize(s string) []string {
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := words[:0]
	for _, w := range words {
		if !stopwords[w] {
			out = append(out, w)
		}
	}
	return out
}

type Index struct {
	tf    []map[string]int // term counts per document
	lens  []int
	df    map[string]int // how many documents contain each term
	avgdl float64
}

func New(texts []string) *Index {
	ix := &Index{df: map[string]int{}}
	total := 0
	for _, t := range texts {
		counts := map[string]int{}
		toks := Tokenize(t)
		for _, w := range toks {
			counts[w]++
		}
		for w := range counts {
			ix.df[w]++
		}
		ix.tf = append(ix.tf, counts)
		ix.lens = append(ix.lens, len(toks))
		total += len(toks)
	}
	if len(texts) > 0 {
		ix.avgdl = float64(total) / float64(len(texts))
	}
	return ix
}

// Scores returns one BM25 score per document (0 = no query word matched).
func (ix *Index) Scores(query string) []float64 {
	n := float64(len(ix.tf))
	seen := map[string]bool{}
	scores := make([]float64, len(ix.tf))
	for _, w := range Tokenize(query) {
		if seen[w] {
			continue
		}
		seen[w] = true
		df := float64(ix.df[w])
		if df == 0 {
			continue
		}
		idf := math.Log(1 + (n-df+0.5)/(df+0.5))
		for i, counts := range ix.tf {
			f := float64(counts[w])
			if f == 0 {
				continue
			}
			norm := 1 - b + b*float64(ix.lens[i])/ix.avgdl
			scores[i] += idf * f * (k1 + 1) / (f + k1*norm)
		}
	}
	return scores
}
