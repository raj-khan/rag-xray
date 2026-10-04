package chunk

import (
	"strings"
	"testing"
)

func TestFixedOverlap(t *testing.T) {
	got := Fixed("abcdefghij", 4, 1)
	want := []string{"abcd", "defg", "ghij"}
	if len(got) != len(want) {
		t.Fatalf("got %d chunks, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Text != want[i] {
			t.Errorf("chunk %d = %q, want %q", i, got[i].Text, want[i])
		}
	}
}

func TestFixedBadOverlapDoesNotLoop(t *testing.T) {
	if got := Fixed("abcdef", 2, 5); len(got) != 3 {
		t.Fatalf("got %d chunks, want 3", len(got))
	}
}

func TestMarkdownHeadingPath(t *testing.T) {
	doc := "# Guide\n\nIntro.\n\n## Setup\n\nStep one.\n\n### Linux\n\nUse apt.\n\n## Usage\n\nRun it.\n"
	got := Markdown(doc, 500)
	want := []Chunk{
		{"Guide", "Intro."},
		{"Guide > Setup", "Step one."},
		{"Guide > Setup > Linux", "Use apt."},
		{"Guide > Usage", "Run it."},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("chunk %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestMarkdownPacksAndSplits(t *testing.T) {
	doc := "# T\n\naaaa\n\nbbbb\n\ncccc\n"
	if got := Markdown(doc, 9); len(got) != 2 {
		t.Errorf("expected paragraphs packed into 2 chunks, got %+v", got)
	}
	long := "# T\n\n" + strings.Repeat("x", 250)
	for _, c := range Markdown(long, 100) {
		if len(c.Text) > 100 || c.Heading != "T" {
			t.Errorf("bad split chunk: %d chars, heading %q", len(c.Text), c.Heading)
		}
	}
}
