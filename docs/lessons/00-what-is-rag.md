---
title: What is RAG?
order: 0
minutes: 10
summary: Why language models need retrieval, and the two halves of every RAG system.
---

# What is RAG?

A language model only knows what was in its training data. It has never seen your company handbook, your notes, or anything written after its training cutoff. Ask it anyway and it will either refuse or, worse, invent a confident answer.

**Retrieval-Augmented Generation (RAG)** fixes this without retraining the model:

1. **Retrieve** the few passages from your documents that are most relevant to the question.
2. **Augment** the prompt by pasting those passages in front of the question.
3. **Generate** an answer that is grounded in, and cites, those passages.

That is the whole idea. Everything else in this course is about doing step 1 well.

## The two halves

```text
OFFLINE (once, and whenever documents change)

  documents ──► chunk ──► embed ──► vector store
                 (L2)      (L1)        (L3)

ONLINE (every question)

  question ──► embed ──► search store ──► top-k chunks ──► prompt ──► LLM ──► answer
                (L1)       (L3, L5)                         (L4)               + sources
```

- **Indexing** (offline) turns documents into small, searchable pieces, each with a vector that captures its meaning.
- **Querying** (online) finds the pieces closest to the question and hands them to the model.

## Why not just paste everything into the prompt?

Modern models accept very long prompts, and for a handful of short documents that genuinely works. RAG still wins when:

- the documents are larger than the context window, or keep growing;
- you pay per token and send many questions;
- you need **citations**, so people can check where an answer came from;
- smaller local models get confused by long, noisy prompts and do better with a few focused passages.

## Why not fine-tune?

Fine-tuning changes *how* a model writes, but it is a poor way to teach it *facts*: it is slow, costly, hard to update, and the model still cannot tell you its source. RAG keeps knowledge in plain documents you can edit at any time.

## What you will build

Six small Go programs, each adding one piece, all running against a local model by default:

| Lesson | You learn | Program |
|---|---|---|
| 1 | Embeddings and similarity | `lessons/01-embeddings` |
| 2 | Chunking | `lessons/02-chunking` |
| 3 | Indexing into a vector store | `lessons/03-index` |
| 4 | The full RAG loop with citations | `lessons/04-ask` |
| 5 | Hybrid search (vectors + keywords) | `lessons/05-hybrid` |
| 6 | Measuring quality | `lessons/06-eval` |
| 7 | Where to go next | reading list |

The sample documents in `data/docs` describe **Orbita Labs**, a company that does not exist. That is deliberate: the model cannot possibly know these facts, so every correct answer has to come from retrieval.

## Check yourself

<details>
<summary>A model answers a question about your private wiki correctly without RAG. What probably happened?</summary>

The answer was common knowledge (or a lucky guess), not something learned from your wiki. Always test RAG with facts the model cannot know, as this course does with a fictional company.
</details>

<details>
<summary>Which half of RAG do you re-run when a document changes?</summary>

Indexing. Only the changed document needs to be re-chunked and re-embedded.
</details>
