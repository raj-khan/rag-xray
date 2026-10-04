// Package lessons loads a course from Markdown files. Any folder of .md
// files works, so anyone can write their own course for the reader:
//
//	---
//	title: Embeddings
//	order: 1
//	minutes: 15
//	run: go run ./lessons/01-embeddings
//	summary: One line shown in the sidebar.
//	---
//	# Markdown body...
//
// All front matter fields are optional.
package lessons

import (
	"bytes"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Lesson struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Order   int    `json:"order"`
	Minutes int    `json:"minutes,omitempty"`
	Run     string `json:"run,omitempty"`
	Summary string `json:"summary,omitempty"`
	HTML    string `json:"html,omitempty"`
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	// Lessons are trusted author content and use <details> for answers.
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// Load reads every .md file in fsys (non-recursive), sorted by order then name.
func Load(fsys fs.FS) ([]Lesson, error) {
	files, err := fs.Glob(fsys, "*.md")
	if err != nil {
		return nil, err
	}
	var out []Lesson
	for i, f := range files {
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		l := parse(strings.TrimSuffix(path.Base(f), ".md"), b)
		if l.Order == 0 && !strings.HasPrefix(l.Slug, "00") {
			l.Order = 1000 + i
		}
		out = append(out, l)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out, nil
}

func parse(slug string, src []byte) Lesson {
	l := Lesson{Slug: slug}
	body := src
	if rest, ok := bytes.CutPrefix(src, []byte("---\n")); ok {
		if fm, after, ok := bytes.Cut(rest, []byte("\n---\n")); ok {
			body = after
			for _, line := range strings.Split(string(fm), "\n") {
				k, v, ok := strings.Cut(line, ":")
				if !ok {
					continue
				}
				v = strings.TrimSpace(v)
				switch strings.TrimSpace(k) {
				case "title":
					l.Title = v
				case "order":
					l.Order, _ = strconv.Atoi(v)
				case "minutes":
					l.Minutes, _ = strconv.Atoi(v)
				case "run":
					l.Run = v
				case "summary":
					l.Summary = v
				}
			}
		}
	}
	if l.Title == "" {
		l.Title = firstHeading(body, slug)
	}
	var buf bytes.Buffer
	if err := md.Convert(body, &buf); err != nil {
		buf.Reset()
		buf.WriteString("<pre>" + err.Error() + "</pre>")
	}
	l.HTML = buf.String()
	return l
}

func firstHeading(body []byte, fallback string) string {
	for _, line := range strings.Split(string(body), "\n") {
		if t, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(t)
		}
	}
	return fallback
}
