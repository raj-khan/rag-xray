---
title: Embeddings and similarity
order: 1
minutes: 15
run: go run ./lessons/01-embeddings "how do I make pasta"
summary: Turn text into vectors and measure how close two meanings are.
---

# Embeddings and similarity

## The idea

An **embedding model** reads a piece of text and outputs a fixed-length list of numbers, a **vector**. `nomic-embed-text` outputs 768 numbers for any input, whether it is one word or a page.

The model is trained so that texts with similar *meaning* produce vectors that point in similar *directions*, even when they share no words. "How do I make pasta" lands near "Boil the spaghetti in salted water" and far from "Goroutines are lightweight threads".

## Measuring closeness: cosine similarity

```go
func cosine(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
```

It compares only the *angle* between two vectors, ignoring their length: `1` means the same direction, `0` unrelated, `-1` opposite. Retrieval is simply "embed the question, then find the stored vectors with the highest cosine".

## Run it

```sh
go run ./lessons/01-embeddings "how do I make pasta"
```

Typical output with `nomic-embed-text`:

```text
0.585 #######################      Boil the spaghetti in salted water for about nine minutes.
0.579 #######################      A good tomato sauce simmers slowly with garlic and olive oil.
0.462 ##################           Goroutines are lightweight threads managed by the Go runtime.
0.431 #################            Kittens need to eat several small meals a day.
0.424 ################             The cat curled up on the warm windowsill and fell asleep.
0.396 ###############              Use a channel to pass values safely between goroutines.
0.296 ###########                  The stock market fell sharply after the interest rate decision.
```

## What to notice

- **The ranking is right**: both cooking sentences come first, though neither contains the word "pasta".
- **The scores are not near 0 for unrelated text.** Goroutines score 0.46 against a pasta question. Every embedding model has its own "baseline" similarity, so a fixed cutoff like "ignore anything below 0.5" does not transfer between models. **Rank, don't threshold.**
- **The gap is small** (0.58 vs 0.46). Embeddings are a fuzzy signal, which is why later lessons add keyword search and evaluation.

## Experiments

1. Try `"my kitten is hungry"` and `"concurrency in Go"`. Does the right pair win each time?
2. Try a query in another language, such as `"comment faire des pâtes"`. Small embedding models handle some languages much better than others.
3. Add a sentence that is a *negation*, like "I never eat pasta". Does it rank near the top for the pasta query? (Usually yes: embeddings capture topic far better than logic.)
4. Switch to a hosted model (`EMBED_PROVIDER=openai`) and compare the score ranges.

## Check yourself

<details>
<summary>Why can't you compare a vector from one embedding model with a vector from another?</summary>

Each model invents its own coordinate system. Dimension 17 means something different in each, and they often have different lengths (768 vs 1536). Vectors are only comparable within one model, which is why the index records the model it was built with.
</details>

<details>
<summary>Why cosine instead of plain Euclidean distance?</summary>

Cosine ignores vector length, which mostly reflects things like text length rather than meaning. Many models also return unit-length vectors, in which case cosine, dot product and Euclidean distance all produce the same ranking.
</details>

## Further reading

- [Nomic Embed model card](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5), including the task prefixes this project uses.
- [MTEB leaderboard](https://huggingface.co/spaces/mteb/leaderboard), which compares embedding models on retrieval tasks.
