---
title: Hybrid search
order: 5
minutes: 20
run: go run ./lessons/05-hybrid
summary: Combine meaning-based vector search with exact keyword search, and see where each one fails.
---

# Hybrid search

## Two kinds of search, two kinds of failure

**Vector search** matches *meaning*. It finds "I'm ill and can't come to work" → *Sick leave*, though they share no words. But it is fuzzy about **exact tokens**: product names, error codes, IDs, rare words the embedding model never learned.

**Keyword search (BM25)** matches *words*. It nails `Vaultbox` and `KST-401` instantly. But it is blind to synonyms, and common words can mislead it.

BM25 in one paragraph: a chunk scores higher the more often it contains the query's words, words that are **rare across the corpus** count much more than common ones, and repetition gives diminishing returns. See `internal/bm25`, about 80 lines.

## Fusing the two: Reciprocal Rank Fusion

Cosine scores (around 0.5) and BM25 scores (around 3 to 6) are on different scales, so you cannot just add them. **RRF** ignores scores and uses only **ranks**:

```text
score(chunk) = Σ over each result list   1 / (60 + rank)
```

A chunk that ranks high in both lists beats one that ranks first in only one. The constant 60 comes from the original paper and rarely needs tuning.

## Run it

```sh
go run ./lessons/05-hybrid
go run ./lessons/05-hybrid "your own query"
```

Real results on the sample corpus:

| Query | Vector #1 | Keyword #1 | Hybrid #1 |
|---|---|---|---|
| `Vaultbox` | Kestrel > Authentication ✗ | Security > Passwords ✓ | Security > Passwords ✓ |
| `KST-401` | Kestrel > Error codes ✓ | Kestrel > Error codes ✓ | Kestrel > Error codes ✓ |
| `I'm ill and can't come to work` | Leave > Sick leave ✓ | Security > Passwords ✗ | Office Guide ✗ (sick leave drops to #3) |
| `too many requests error` | Kestrel > Rate limits ✓ | Kestrel > Rate limits ✓ | Kestrel > Rate limits ✓ |

## What to notice

- **Hybrid is not magic.** On the "ill" query, keyword search matched the word "work" in unrelated chunks and dragged the right answer from #1 to #3. Hybrid usually *raises the floor* (fewer total misses), but it can lower the top result.
- This is why production systems add a **reranker**: retrieve ~20 candidates with hybrid search, then let a cross-encoder model (or an LLM) re-score each candidate against the question and keep the best few. Anthropic reported that contextual embeddings plus contextual BM25 cut retrieval failures by 49%, and adding reranking brought it to 67%.
- Look at the measured effect in lesson 6 rather than trusting a single example.

## Experiments

1. Add "work" to the stopword list in `internal/bm25` and rerun the "ill" query. Then think about why stopword lists are a blunt tool.
2. Give vector search more weight: change RRF so the vector list counts twice. Does the eval improve?
3. Write a query using a synonym the documents never use ("PTO", "holiday allowance"). Which method survives?

## Check yourself

<details>
<summary>Why does BM25 weight rare words more?</summary>

A word that appears in every chunk (like "the" or "policy") says nothing about which chunk is relevant. A word that appears in one chunk (like "Vaultbox") almost identifies it. That is inverse document frequency (IDF).
</details>

## Further reading

- Cormack, Clarke & Büttcher (2009), *Reciprocal Rank Fusion outperforms Condorcet and individual rank learning methods*.
- [Anthropic: Introducing Contextual Retrieval](https://www.anthropic.com/news/contextual-retrieval), on contextual BM25, fusion and reranking.
- Robertson & Zaragoza (2009), *The Probabilistic Relevance Framework: BM25 and Beyond*.
