// Command ragxray serves the lesson reader and the RAG X-ray playground.
//
//	go run ./cmd/ragxray                      # http://localhost:8080
//	go run ./cmd/ragxray -lessons ./my-course # read your own Markdown course
package main

import (
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/raj-khan/rag-xray/data"
	"github.com/raj-khan/rag-xray/docs"
	"github.com/raj-khan/rag-xray/internal/ai"
	"github.com/raj-khan/rag-xray/internal/lessons"
	"github.com/raj-khan/rag-xray/internal/retrieve"
	"github.com/raj-khan/rag-xray/internal/server"
	"github.com/raj-khan/rag-xray/web"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", "127.0.0.1:8080"), "listen address (use 0.0.0.0:8080 inside Docker)")
	lessonDir := flag.String("lessons", "", "folder of Markdown lessons (default: the built-in RAG course)")
	docDir := flag.String("docs", "", "folder of .md/.txt documents to start with (default: the built-in sample)")
	flag.Parse()

	var lessonFS fs.FS = mustSub(docs.Lessons, "lessons")
	if *lessonDir != "" {
		lessonFS = os.DirFS(*lessonDir)
	}
	ls, err := lessons.Load(lessonFS)
	if err != nil {
		log.Fatal(err)
	}

	sample, err := server.LoadDocs(data.Docs, "docs")
	if err != nil {
		log.Fatal(err)
	}
	if *docDir != "" {
		if sample, err = retrieve.LoadDir(*docDir); err != nil {
			log.Fatal(err)
		}
	}

	e, err := ai.NewEmbedder()
	if err != nil {
		log.Fatal(err)
	}
	c, err := ai.NewChatter()
	if err != nil {
		log.Fatal(err)
	}

	srv := server.New(ls, sample, e, c)
	log.Printf("rag-xray: %d lessons, %d documents, embed=%s chat=%s", len(ls), len(sample), e.Model(), c.Model())
	log.Printf("open http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, srv.Handler(mustSub(web.Static, "static"))))
}

func mustSub(f fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		log.Fatal(err)
	}
	return sub
}

func envOr(k, v string) string {
	if s := os.Getenv(k); s != "" {
		return s
	}
	return v
}
