package bm25

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	got := Tokenize("What is KST-429, the Rate limit?")
	want := []string{"kst", "429", "rate", "limit"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tokenize = %v, want %v", got, want)
	}
}

func TestScoresPreferRareMatchingTerms(t *testing.T) {
	ix := New([]string{
		"the cat sat on the mat",
		"error code KST-429 means rate limited",
		"the dog sat on the log",
	})
	s := ix.Scores("KST-429 error")
	if !(s[1] > 0 && s[0] == 0 && s[2] == 0) {
		t.Errorf("unexpected scores %v", s)
	}
	s = ix.Scores("cat sat")
	if !(s[0] > s[2] && s[2] > 0) {
		t.Errorf("rare word 'cat' should outweigh common 'sat': %v", s)
	}
}
