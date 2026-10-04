---
title: Chunking
order: 2
minutes: 15
run: go run ./lessons/02-chunking data/docs/kestrel-api.md
summary: Split documents into pieces that are small enough to be precise and big enough to make sense.
---

# Chunking

## Why split documents at all?

One vector has to summarise the whole text it was made from. Embed an entire handbook as one vector and it becomes a blurry average of everything in it: it matches every question a little and none of them well. You also cannot paste whole documents into a prompt once you have many.

So we cut documents into **chunks** and embed each one separately. The trade-off:

- **Too small** (one sentence): precise, but loses context. "It is capped at 180 USD" means nothing alone.
- **Too big** (a whole page): has context, but the vector is blurry and wastes prompt space.

A few hundred to a thousand characters is a common starting point. **Measure it** for your own documents (lesson 6) instead of trusting a magic number.

## Two strategies

**Fixed-size windows** cut every N characters, with some overlap so that a sentence cut at a boundary appears whole in at least one chunk.

```go
chunk.Fixed(text, 300, 50) // 300 chars, 50 shared with the previous chunk
```

**Structure-aware** chunking (`chunk.Markdown`) follows the document's own shape:

- every heading starts a new chunk;
- paragraphs are packed together up to a size limit, never split mid-paragraph unless one is huge;
- each chunk remembers its **heading path**, for example `Kestrel API Reference > Rate limits`.

## Run it

```sh
go run ./lessons/02-chunking data/docs/kestrel-api.md
go run ./lessons/02-chunking -size 150 -overlap 0 data/docs/leave-policy.md
```

## What to notice

A fixed chunk from the run:

```text
--- #1 (300 chars)
    ated every 90 days. Keys are created and revoked in the Kestrel console.

    ## Rate limits

    Standard plan customers can send 600 requests per minute. ...
    When a limit is exceeded the API returns HTTP 4
```

It starts mid-word, mixes two topics, and ends mid-number. The Markdown chunker instead gives one clean chunk per section, labelled with where it came from.

## The heading path is context

When indexing (lesson 3) we embed not just the chunk text but:

```text
Document: kestrel-api.md
Section: Kestrel API Reference > Rate limits

Standard plan customers can send 600 requests per minute. ...
```

This is a lightweight version of Anthropic's **Contextual Retrieval** idea: give every chunk enough context to stand on its own. Their full version asks an LLM to write a short summary of where each chunk sits in its document; we get much of the benefit for free from headings.

## Experiments

1. Run the fixed chunker with `-overlap 0`. Find a fact that is now split across two chunks.
2. Point lesson 2 at one of your own Markdown notes. Do your headings produce sensible chunks?
3. Re-index with `-max 200` and `-max 1500` (lesson 3), then run the eval (lesson 6). Which size wins on this corpus?

## Check yourself

<details>
<summary>What problem does overlap solve, and what does it cost?</summary>

It keeps sentences near a boundary intact in at least one chunk. The cost is duplicated text: more chunks to embed and store, and near-duplicate results in search.
</details>

<details>
<summary>Your documents are PDFs with no headings. What would you do?</summary>

Extract the text first, then split on paragraphs or sentences with a size limit and some overlap. If the PDF has visual structure (titles, sections), a layout-aware extractor can recover headings to use as context.
</details>

## Further reading

- [Anthropic: Introducing Contextual Retrieval](https://www.anthropic.com/news/contextual-retrieval)
- [RAG_Techniques](https://github.com/NirDiamant/RAG_Techniques) by Nir Diamant has notebooks on semantic chunking, proposition chunking and chunk-size selection.
