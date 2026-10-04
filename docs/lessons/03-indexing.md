---
title: Indexing and vector stores
order: 3
minutes: 15
run: go run ./lessons/03-index
summary: Chunk and embed a folder of documents, then save the vectors so questions can be answered fast.
---

# Indexing and vector stores

## The offline half of RAG

Embedding is the slow part, so we do it once, ahead of time:

1. **Load** every `.md` or `.txt` file in a folder.
2. **Chunk** each one (lesson 2).
3. **Embed** each chunk, in batches.
4. **Save** chunks and vectors to a **vector store**.

The whole pipeline is `retrieve.Build` in about 25 lines.

## Our vector store is a JSON file

```go
type Chunk struct {
	ID      int
	Source  string    // file name
	Heading string    // heading path
	Text    string
	Vector  []float64 // 768 numbers
}
```

Search is a loop: compute cosine against every chunk and keep the top k. This is called **brute-force** or **exact** search, and it is honestly fine up to tens of thousands of chunks.

Real vector databases (pgvector, Qdrant, LanceDB, sqlite-vec, Chroma and others) add:

- **Approximate nearest neighbour (ANN) indexes** such as HNSW, which find *nearly* the best matches in milliseconds among millions of vectors;
- metadata filtering ("only documents from team X");
- updates and deletes without rebuilding;
- persistence, replication and access control.

None of that changes the idea. You can swap them in later without touching the rest of the pipeline.

## Run it

```sh
go run ./lessons/03-index
go run ./lessons/03-index -data ~/my-notes -out store/notes.json
```

Output on the sample corpus:

```text
Embedded 35 chunks with nomic-embed-text in 6.0s (768 dims each)
Saved store/index.json (334.1 KB)
```

## Study anything

Point `-data` at **any folder of Markdown or text files**: your course notes, a project's docs, a book exported to Markdown. Lessons 4 to 6 accept `-index` to use that index. The web app (`ragschool`) does the same by drag and drop.

## The index remembers its model

The store records which embedding model built it. If you change `EMBED_MODEL` and ask a question, you get a clear error instead of silently wrong results, because vectors from different models are not comparable (lesson 1).

## Experiments

1. Look inside `store/index.json`. How much of the file is vectors and how much is text?
2. Estimate: how big would the file be for 100,000 chunks? (Roughly 768 numbers × ~10 bytes each as JSON text, per chunk.) Why do real stores save vectors as binary `float32`?
3. Index your own notes and ask them questions with lesson 4.

## Check yourself

<details>
<summary>When does exact search stop being good enough?</summary>

When the loop over every vector gets too slow for your latency budget, typically somewhere past a few hundred thousand vectors on one machine. That is the point to add an ANN index.
</details>

<details>
<summary>A document is edited. What is the minimum work to keep the index fresh?</summary>

Delete that document's chunks, then re-chunk and re-embed only that document. A content hash per document tells you which ones changed.
</details>
