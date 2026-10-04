package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/raj-khan/rag-xray/internal/ai"
	"github.com/raj-khan/rag-xray/internal/lessons"
	"github.com/raj-khan/rag-xray/internal/retrieve"
)

// fakeEmbedder maps text to a tiny bag-of-letters vector: deterministic and offline.
type fakeEmbedder struct{ calls int }

func (f *fakeEmbedder) Model() string { return "fake-embed" }
func (f *fakeEmbedder) Embed(texts []string) ([][]float64, error) {
	f.calls += len(texts)
	out := make([][]float64, len(texts))
	for i, t := range texts {
		v := make([]float64, 26)
		for _, r := range strings.ToLower(t) {
			if r >= 'a' && r <= 'z' {
				v[r-'a']++
			}
		}
		out[i] = v
	}
	return out, nil
}

type fakeChat struct{}

func (fakeChat) Model() string { return "fake-chat" }
func (fakeChat) Chat(msgs []ai.Message, on func(string)) (string, error) {
	for _, t := range []string{"It is ", "24 days [1]."} {
		on(t)
	}
	return "It is 24 days [1].", nil
}

func newTestServer(t *testing.T) (*Server, *fakeEmbedder, http.Handler) {
	t.Helper()
	e := &fakeEmbedder{}
	sample := []retrieve.Doc{
		{Name: "leave.md", Text: "# Leave\n\n## Annual\n\nEmployees get 24 days of annual leave."},
		{Name: "api.md", Text: "# API\n\n## Errors\n\nKST-503 means maintenance."},
	}
	ls := []lessons.Lesson{{Slug: "01-intro", Title: "Intro", HTML: "<p>hi</p>"}}
	s := New(ls, sample, e, fakeChat{})
	waitReady(t, s)
	return s, e, s.Handler(fstest.MapFS{"index.html": {Data: []byte("ok")}})
}

func waitReady(t *testing.T, s *Server) {
	t.Helper()
	for range 200 {
		s.mu.RLock()
		ready, building := s.ret != nil, s.building
		s.mu.RUnlock()
		if ready && !building {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("index never became ready")
}

func TestAskStreamsEveryStage(t *testing.T) {
	_, _, h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/ask", strings.NewReader(`{"question":"annual leave days","k":1}`)))
	body := rec.Body.String()
	for _, ev := range []string{"event: retrieval", "event: prompt", "event: token", "event: done"} {
		if !strings.Contains(body, ev) {
			t.Errorf("missing %q in stream:\n%s", ev, body)
		}
	}
	if !strings.Contains(body, `"source":"leave.md"`) {
		t.Errorf("expected leave.md to be retrieved:\n%s", body)
	}
}

func TestAskRequiresQuestion(t *testing.T) {
	_, _, h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/ask", strings.NewReader(`{"question":"  "}`)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("code = %d", rec.Code)
	}
}

func TestUploadReindexesAndCachesEmbeddings(t *testing.T) {
	_, e, h := newTestServer(t)
	before := e.calls

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("files", "../../notes.md") // path must be stripped
	fw.Write([]byte("# Notes\n\nCoffee ratio is 1:16."))
	skip, _ := mw.CreateFormFile("files", "evil.exe")
	skip.Write([]byte("nope"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/docs", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}
	var info struct {
		Docs []struct{ Name string } `json:"docs"`
	}
	json.Unmarshal(rec.Body.Bytes(), &info)
	names := []string{}
	for _, d := range info.Docs {
		names = append(names, d.Name)
	}
	if strings.Join(names, ",") != "leave.md,api.md,notes.md" {
		t.Errorf("docs = %v", names)
	}
	if got := e.calls - before; got != 1 {
		t.Errorf("re-index embedded %d texts, want 1 (others cached)", got)
	}
}

func TestLessonEndpoints(t *testing.T) {
	_, _, h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/lessons", nil))
	if strings.Contains(rec.Body.String(), `"html"`) {
		t.Error("lesson list should not include HTML bodies")
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/lessons/01-intro", nil))
	var l lessons.Lesson
	json.Unmarshal(rec.Body.Bytes(), &l)
	if l.HTML != "<p>hi</p>" {
		t.Errorf("lesson body = %q", l.HTML)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/lessons/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing lesson code = %d", rec.Code)
	}
}
