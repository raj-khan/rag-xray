// Package chunk splits documents into retrievable pieces.
package chunk

import "strings"

type Chunk struct {
	Heading string // e.g. "Kestrel API > Rate limits"
	Text    string
}

// Fixed cuts text into windows of size characters, each sharing overlap
// characters with the previous one. Simple, but it cuts mid-sentence.
func Fixed(text string, size, overlap int) []Chunk {
	r := []rune(text)
	if overlap >= size {
		overlap = 0
	}
	var out []Chunk
	for start := 0; start < len(r); start += size - overlap {
		end := min(start+size, len(r))
		if t := strings.TrimSpace(string(r[start:end])); t != "" {
			out = append(out, Chunk{Text: t})
		}
		if end == len(r) {
			break
		}
	}
	return out
}

// Markdown respects document structure: every heading starts a new chunk,
// paragraphs are packed together up to maxChars, and each chunk remembers
// its heading path so it keeps its context after being cut out.
func Markdown(text string, maxChars int) []Chunk {
	var (
		out    []Chunk
		path   []string // current heading path
		buf    []string // paragraphs in the chunk being built
		bufLen int
		para   []string // lines of the paragraph being read
	)
	heading := func() string { return strings.Join(path, " > ") }
	flushChunk := func() {
		if len(buf) > 0 {
			out = append(out, Chunk{Heading: heading(), Text: strings.Join(buf, "\n\n")})
			buf, bufLen = nil, 0
		}
	}
	endPara := func() {
		if len(para) == 0 {
			return
		}
		p := strings.Join(para, "\n")
		para = nil
		if len(p) > maxChars { // one huge paragraph: fall back to fixed windows
			flushChunk()
			for _, c := range Fixed(p, maxChars, maxChars/5) {
				c.Heading = heading()
				out = append(out, c)
			}
			return
		}
		if bufLen+len(p) > maxChars {
			flushChunk()
		}
		buf = append(buf, p)
		bufLen += len(p)
	}

	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "#"):
			endPara()
			flushChunk()
			level := len(t) - len(strings.TrimLeft(t, "#"))
			if level-1 < len(path) {
				path = path[:level-1]
			}
			path = append(path, strings.TrimSpace(t[level:]))
		case t == "":
			endPara()
		default:
			para = append(para, line)
		}
	}
	endPara()
	flushChunk()
	return out
}
