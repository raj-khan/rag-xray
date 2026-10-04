// Package server is the rag-xray web app: a lesson reader plus a
// playground that shows every step of the RAG pipeline on any documents.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/raj-khan/rag-xray/internal/ai"
	"github.com/raj-khan/rag-xray/internal/lessons"
	"github.com/raj-khan/rag-xray/internal/retrieve"
)

const (
	maxUpload   = 10 << 20 // 10 MB per upload request
	maxDocs     = 200
	candidates  = 10 // results shown per retrieval method
	defaultSize = 500
)

type Server struct {
	lessons []lessons.Lesson
	sample  []retrieve.Doc
	embed   *cachedEmbedder
	chat    ai.Chatter

	buildMu sync.Mutex // one index build at a time

	mu       sync.RWMutex // guards the fields below
	docs     []retrieve.Doc
	maxChars int
	ret      *retrieve.Retriever
	building bool
	indexErr string
	buildMs  int64
}

func New(ls []lessons.Lesson, sample []retrieve.Doc, e ai.Embedder, c ai.Chatter) *Server {
	s := &Server{
		lessons:  ls,
		sample:   sample,
		embed:    &cachedEmbedder{Embedder: e, cache: map[string][]float64{}},
		chat:     c,
		docs:     append([]retrieve.Doc(nil), sample...),
		maxChars: defaultSize,
	}
	go s.rebuild() // index in the background so the reader is usable at once
	return s
}

func (s *Server) Handler(static fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", s.info)
	mux.HandleFunc("GET /api/lessons", s.listLessons)
	mux.HandleFunc("GET /api/lessons/{slug}", s.getLesson)
	mux.HandleFunc("GET /api/chunks", s.chunks)
	mux.HandleFunc("POST /api/docs", s.upload)
	mux.HandleFunc("DELETE /api/docs/{name}", s.deleteDoc)
	mux.HandleFunc("POST /api/docs/reset", s.resetDocs)
	mux.HandleFunc("POST /api/index", s.reindex)
	mux.HandleFunc("POST /api/ask", s.ask)
	mux.Handle("GET /", http.FileServerFS(static))
	return securityHeaders(mux)
}

// ---------- indexing ----------

// rebuild re-indexes a snapshot of the documents without holding the
// read lock, so the UI can keep polling /api/info while it runs.
func (s *Server) rebuild() {
	s.buildMu.Lock()
	defer s.buildMu.Unlock()

	s.mu.Lock()
	docs := slices.Clone(s.docs)
	size := s.maxChars
	s.building = true
	s.mu.Unlock()

	start := time.Now()
	var ret *retrieve.Retriever
	st, err := retrieve.Build(docs, size, s.embed)
	if err == nil {
		ret, err = retrieve.New(st, s.embed, s.chat)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ret, s.indexErr, s.building = ret, "", false
	if err != nil {
		s.indexErr = err.Error()
	}
	s.buildMs = time.Since(start).Milliseconds()
}

func (s *Server) reindex(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MaxChars int `json:"maxChars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	s.mu.Lock()
	s.maxChars = min(max(req.MaxChars, 100), 4000)
	s.mu.Unlock()
	s.rebuild()
	s.info(w, r)
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		httpError(w, http.StatusBadRequest, fmt.Errorf("upload too large or malformed (limit %d MB)", maxUpload>>20))
		return
	}
	s.mu.Lock()
	added := 0
	for _, fh := range r.MultipartForm.File["files"] {
		name := filepath.Base(fh.Filename)
		ext := strings.ToLower(path.Ext(name))
		if ext != ".md" && ext != ".markdown" && ext != ".txt" {
			continue
		}
		f, err := fh.Open()
		if err != nil {
			continue
		}
		b, err := io.ReadAll(io.LimitReader(f, maxUpload))
		f.Close()
		if err != nil {
			continue
		}
		s.docs = slices.DeleteFunc(s.docs, func(d retrieve.Doc) bool { return d.Name == name })
		if len(s.docs) >= maxDocs {
			break
		}
		s.docs = append(s.docs, retrieve.Doc{Name: name, Text: string(b)})
		added++
	}
	s.mu.Unlock()
	if added == 0 {
		httpError(w, http.StatusBadRequest, errors.New("no .md, .markdown or .txt files found in the upload"))
		return
	}
	s.rebuild()
	s.info(w, r)
}

func (s *Server) deleteDoc(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	s.mu.Lock()
	s.docs = slices.DeleteFunc(s.docs, func(d retrieve.Doc) bool { return d.Name == name })
	s.mu.Unlock()
	s.rebuild()
	s.info(w, r)
}

func (s *Server) resetDocs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.docs = append([]retrieve.Doc(nil), s.sample...)
	s.mu.Unlock()
	s.rebuild()
	s.info(w, r)
}

// ---------- read-only endpoints ----------

func (s *Server) info(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	type docInfo struct {
		Name   string `json:"name"`
		Chars  int    `json:"chars"`
		Chunks int    `json:"chunks"`
	}
	perDoc := map[string]int{}
	total := 0
	if s.ret != nil {
		for _, c := range s.ret.Store.Chunks {
			perDoc[c.Source]++
		}
		total = len(s.ret.Store.Chunks)
	}
	docs := []docInfo{}
	for _, d := range s.docs {
		docs = append(docs, docInfo{d.Name, len(d.Text), perDoc[d.Name]})
	}
	writeJSON(w, map[string]any{
		"embedModel": s.embed.Model(),
		"chatModel":  s.chat.Model(),
		"docs":       docs,
		"chunks":     total,
		"maxChars":   s.maxChars,
		"ready":      s.ret != nil,
		"building":   s.building,
		"indexError": s.indexErr,
		"buildMs":    s.buildMs,
	})
}

func (s *Server) listLessons(w http.ResponseWriter, _ *http.Request) {
	out := make([]lessons.Lesson, len(s.lessons))
	for i, l := range s.lessons {
		l.HTML = ""
		out[i] = l
	}
	writeJSON(w, out)
}

func (s *Server) getLesson(w http.ResponseWriter, r *http.Request) {
	for _, l := range s.lessons {
		if l.Slug == r.PathValue("slug") {
			writeJSON(w, l)
			return
		}
	}
	httpError(w, http.StatusNotFound, errors.New("no such lesson"))
}

type chunkView struct {
	ID      int     `json:"id"`
	Source  string  `json:"source"`
	Heading string  `json:"heading"`
	Text    string  `json:"text"`
	Score   float64 `json:"score,omitempty"`
	// Ranks in each list (0 = absent), so the UI can show how fusion moved it.
	VectorRank  int `json:"vectorRank,omitempty"`
	KeywordRank int `json:"keywordRank,omitempty"`
}

func (s *Server) chunks(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []chunkView{}
	if s.ret != nil {
		for _, c := range s.ret.Store.Chunks {
			out = append(out, chunkView{ID: c.ID, Source: c.Source, Heading: c.Heading, Text: c.Text})
		}
	}
	writeJSON(w, out)
}

// ---------- the X-ray ----------

// ask streams server-sent events, one per pipeline stage:
// retrieval -> prompt -> token... -> done (or error).
func (s *Server) ask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question string `json:"question"`
		K        int    `json:"k"`
		Mode     string `json:"mode"`
		NoRAG    bool   `json:"noRag"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req); err != nil || strings.TrimSpace(req.Question) == "" {
		httpError(w, http.StatusBadRequest, errors.New("question is required"))
		return
	}
	req.K = min(max(req.K, 1), candidates)

	s.mu.RLock()
	ret := s.ret
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, _ := w.(http.Flusher)
	send := func(event string, data any) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		if flusher != nil {
			flusher.Flush()
		}
	}
	fail := func(err error) { send("error", map[string]string{"message": err.Error()}) }

	msgs := []ai.Message{{Role: "user", Content: req.Question}}
	if !req.NoRAG {
		if ret == nil {
			fail(errors.New("the index is not ready yet (or failed: see the Documents panel)"))
			return
		}
		t0 := time.Now()
		vector, err := ret.Vector(req.Question, candidates)
		if err != nil {
			fail(err)
			return
		}
		vectorMs := time.Since(t0).Milliseconds()
		t1 := time.Now()
		keyword := ret.Keyword(req.Question, candidates)
		keywordMs := time.Since(t1).Milliseconds()
		hybrid := retrieve.Fuse(vector, keyword, candidates)

		var final []retrieve.Hit
		switch req.Mode {
		case "vector":
			final = vector
		case "keyword":
			final = keyword
		default:
			req.Mode, final = "hybrid", hybrid
		}
		final = final[:min(req.K, len(final))]

		vr, kr := ranks(vector), ranks(keyword)
		view := func(hits []retrieve.Hit) []chunkView {
			out := []chunkView{}
			for _, h := range hits {
				c := h.Chunk
				out = append(out, chunkView{c.ID, c.Source, c.Heading, c.Text, h.Score, vr[c.ID], kr[c.ID]})
			}
			return out
		}
		send("retrieval", map[string]any{
			"mode":      req.Mode,
			"vector":    view(vector),
			"keyword":   view(keyword),
			"hybrid":    view(hybrid),
			"final":     view(final),
			"vectorMs":  vectorMs,
			"keywordMs": keywordMs,
			"chunks":    len(ret.Store.Chunks),
		})
		msgs = []ai.Message{
			{Role: "system", Content: retrieve.SystemPrompt},
			{Role: "user", Content: retrieve.Prompt(req.Question, final)},
		}
	}

	send("prompt", msgs)
	t2 := time.Now()
	first := int64(-1)
	_, err := s.chat.Chat(msgs, func(tok string) {
		if first < 0 {
			first = time.Since(t2).Milliseconds()
		}
		send("token", tok)
	})
	if err != nil {
		fail(err)
		return
	}
	send("done", map[string]any{"firstTokenMs": first, "totalMs": time.Since(t2).Milliseconds(), "model": s.chat.Model()})
}

func ranks(hits []retrieve.Hit) map[int]int {
	m := map[int]int{}
	for i, h := range hits {
		m[h.Chunk.ID] = i + 1
	}
	return m
}

// ---------- helpers ----------

// cachedEmbedder remembers vectors by text, so changing the chunk size or
// adding one file only embeds text it has not seen before.
type cachedEmbedder struct {
	ai.Embedder
	mu    sync.Mutex
	cache map[string][]float64
}

func (c *cachedEmbedder) Embed(texts []string) ([][]float64, error) {
	c.mu.Lock()
	var missing []string
	seen := map[string]bool{}
	for _, t := range texts {
		if _, ok := c.cache[t]; !ok && !seen[t] {
			missing = append(missing, t)
			seen[t] = true
		}
	}
	c.mu.Unlock()
	if len(missing) > 0 {
		vecs, err := c.Embedder.Embed(missing)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		for i, t := range missing {
			c.cache[t] = vecs[i]
		}
		c.mu.Unlock()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([][]float64, len(texts))
	for i, t := range texts {
		out[i] = c.cache[t]
	}
	return out, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func securityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		h.ServeHTTP(w, r)
	})
}

// LoadDocs reads the sample corpus from an fs.FS.
func LoadDocs(fsys fs.FS, dir string) ([]retrieve.Doc, error) {
	files, err := fs.Glob(fsys, path.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var docs []retrieve.Doc
	for _, f := range files {
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		docs = append(docs, retrieve.Doc{Name: path.Base(f), Text: string(b)})
	}
	return docs, nil
}
