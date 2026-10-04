---
title: Beyond the basics
order: 7
minutes: 15
summary: A map of modern RAG techniques, what problem each one solves, and where to learn it.
---

# Beyond the basics

You now have every core piece: chunking, embeddings, a vector store, hybrid retrieval, grounded generation and evaluation. Here is a map of what people add next. **Add one at a time, and only when your eval shows the problem it solves.**

## Better retrieval

| Technique | Problem it solves | Idea |
|---|---|---|
| **Reranking** | The right chunk is in the top 20 but not the top 4. | Re-score candidates with a cross-encoder (e.g. a BGE or Cohere reranker) or an LLM, keep the best few. |
| **Contextual retrieval** | Chunks are ambiguous on their own. | Ask an LLM to write one or two sentences placing each chunk in its document, and prepend it before embedding and BM25. |
| **Query rewriting** | Users phrase questions badly or vaguely. | Have an LLM rewrite the question, or split it into sub-questions, before searching. |
| **HyDE** | Questions and answers look different as text. | Have an LLM write a *hypothetical* answer, then embed and search with that. |
| **Multi-vector / parent-child** | Small chunks are precise but lack context. | Search over small chunks, but hand the LLM the larger section each one belongs to. |
| **Metadata filters** | Results from the wrong product, date or team. | Store fields with each chunk and filter before ranking. |

## Different shapes of RAG

- **Agentic RAG**: instead of always retrieving once, give the model a `search` tool and let it decide when, what and how many times to search. This is how most AI assistants and coding agents use retrieval today, often through **MCP** servers that expose search over docs, tickets or code.
- **GraphRAG**: extract entities and relationships into a knowledge graph, so questions like "what are all the themes across these reports?" can be answered, which top-k chunk search handles poorly.
- **Multimodal RAG**: embed images, tables and page screenshots, not only text.

## Long context vs RAG

Context windows now hold hundreds of thousands of tokens or more. For a small, stable set of documents, **just putting everything in the prompt** (with prompt caching to control cost) is often simpler and better. RAG remains the right tool when the corpus is large or changing, when cost and latency matter, when you need citations, or when a small local model has to stay focused. Many real systems do both: retrieve generously, then let a long-context model read it all.

## Production concerns

- **Freshness**: re-index changed documents incrementally (content hashes).
- **Access control**: filter chunks by what the asking user is allowed to see, *before* they reach the prompt.
- **Prompt injection**: retrieved documents are untrusted input. A document saying "ignore your instructions" must not be obeyed. Keep instructions in the system prompt, and treat context as data.
- **Observability**: log the question, retrieved chunks, prompt and answer for every request, so failures can be replayed.

## Where to learn more

These projects go deeper, mostly in Python notebooks. They are excellent; we link instead of repeating them.

- [**NirDiamant/RAG_Techniques**](https://github.com/NirDiamant/RAG_Techniques): the largest catalogue of advanced techniques, one notebook each.
- [**langchain-ai/rag-from-scratch**](https://github.com/langchain-ai/rag-from-scratch): a video series and notebooks covering query translation, routing, indexing and more.
- [**pguso/rag-from-scratch**](https://github.com/pguso/rag-from-scratch): the same "no black boxes, local models" spirit, in JavaScript.
- [**microsoft/rag-time**](https://github.com/microsoft/rag-time): a structured five-week course.
- [**mlabonne/llm-course**](https://github.com/mlabonne/llm-course): the broader LLM landscape, with roadmaps.
- [**Anthropic: Contextual Retrieval**](https://www.anthropic.com/news/contextual-retrieval): the measured case for contextual chunks, hybrid search and reranking.

## Your capstone

Pick a folder of documents you actually care about. Index it, write 10 eval questions, record your baseline, then add **one** technique from this page and measure whether it helped. Open a pull request with your result; we would love to include learner experiments.
