package lessons

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadFrontMatterAndOrder(t *testing.T) {
	fsys := fstest.MapFS{
		"b.md":     {Data: []byte("---\ntitle: Second\norder: 2\nrun: go run .\n---\n# Ignored\n\n| a | b |\n|---|---|\n| 1 | 2 |\n")},
		"a.md":     {Data: []byte("---\ntitle: First\norder: 1\nminutes: 5\n---\nHello\n")},
		"plain.md": {Data: []byte("# From heading\n\ntext")},
	}
	ls, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{ls[0].Title, ls[1].Title, ls[2].Title}
	if strings.Join(got, "|") != "First|Second|From heading" {
		t.Errorf("titles = %v", got)
	}
	if ls[0].Minutes != 5 || ls[1].Run != "go run ." {
		t.Errorf("front matter not parsed: %+v %+v", ls[0], ls[1])
	}
	if !strings.Contains(ls[1].HTML, "<table>") {
		t.Errorf("GFM tables not rendered: %s", ls[1].HTML)
	}
}
