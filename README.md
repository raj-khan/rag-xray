<div align="center">

<img src="web/static/favicon.svg" alt="" width="64">

# rag-xray

**Learn Retrieval-Augmented Generation by building it, then see inside every step.**

A hands-on RAG course in plain Go, plus a playground that X-rays the whole pipeline on *your own* documents.<br>
Runs fully local with [Ollama](https://ollama.com), or with any AI API.

[![CI](https://github.com/raj-khan/rag-xray/actions/workflows/ci.yml/badge.svg)](https://github.com/raj-khan/rag-xray/actions/workflows/ci.yml)
[![Docker](https://github.com/raj-khan/rag-xray/actions/workflows/docker.yml/badge.svg)](https://github.com/raj-khan/rag-xray/actions/workflows/docker.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Go](https://img.shields.io/badge/go-1.27-00ADD8?logo=go&logoColor=white)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[Quick start](#quick-start) · [The course](#the-course) · [X-ray playground](#the-x-ray-playground) · [Configuration](#configuration) · [Contributing](#contributing) · [Security](#security)

<img src="docs/media/demo.webp" alt="Asking a question in rag-xray: the answer streams in while vector, keyword and hybrid rankings appear side by side" width="860">

▶ **[Watch the full demo video (MP4)](docs/media/demo.mp4)**

</div>

## Table of contents

- [Why rag-xray?](#why-rag-xray)
- [Features](#features)
- [Quick start](#quick-start)
- [The course](#the-course)
- [The X-ray playground](#the-x-ray-playground)
- [Study anything](#study-anything)
- [Configuration](#configuration)
- [Results on the sample corpus](#results-on-the-sample-corpus)
- [Project layout](#project-layout)
- [Troubleshooting](#troubleshooting)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Security](#security)
- [Learn more elsewhere](#learn-more-elsewhere)
- [License](#license)

## Why rag-xray?

Most RAG tutorials hand you a framework and a notebook. It works, but you never *see* why it works, or why it fails.

rag-xray shows every step of the pipeline, **R**etrieve, **A**ugment, **G**enerate, side by side, and lets you turn the knobs and watch what changes. The code is small enough to read in an afternoon, uses no framework, and the lessons quote what the programs really print, including the failures.

**Who it is for:** developers who want to understand RAG well enough to build and debug their own, students, and anyone who wants to ask questions of their notes with a local model.

## Features

- 📚 **An 8-lesson course** (about two hours), each lesson with a runnable Go program, real output, experiments and self-check questions.
- 🔬 **X-ray playground**: vector search, BM25 keyword search and their fusion side by side, the exact prompt, and the streamed answer with clickable citations and timings.
- 🎛️ **Live knobs**: chunk size, top-k, retrieval mode, and a no-retrieval toggle to compare against the model's memory.
- 📂 **Bring your own documents**: drag and drop `.md` / `.txt` files and study anything.
- ✍️ **Bring your own course**: any folder of Markdown becomes a course in the reader.
- 🔌 **Any model**: local Ollama by default; OpenAI and any OpenAI-compatible API (Gemini, Groq, OpenRouter, LM Studio, llama.cpp, vLLM); Anthropic Claude.
- 📏 **Evaluation built in**: a labelled question set with hit@k, MRR and answer accuracy.
- 🐳 **One command with Docker**, or a single Go binary with everything embedded. Light and dark themes, works on phones.

## Quick start

### Option 1: Docker, fully local (no installs, no API keys)

```sh
git clone https://github.com/raj-khan/rag-xray && cd rag-xray
docker compose up
```

Open **http://localhost:8080**. The first start downloads two small open models (about 2.3 GB) into a Docker volume.

### Option 2: Go + Ollama

Requires [Go 1.27+](https://go.dev/dl/) and [Ollama](https://ollama.com/download).

```sh
git clone https://github.com/raj-khan/rag-xray && cd rag-xray
ollama pull nomic-embed-text && ollama pull llama3.2
go run ./cmd/ragxray          # http://localhost:8080
```

### Option 3: any AI provider

```sh
cp .env.example .env          # uncomment a recipe and add your key
set -a; source .env; set +a   # or export the variables yourself
go run ./cmd/ragxray
```

With Docker and a hosted provider (no Ollama needed): `docker compose up --no-deps --build app`.

### Option 4: published image, with Ollama already on your machine

```sh
docker run --rm -p 8080:8080 --add-host=host.docker.internal:host-gateway \
  -e OLLAMA_HOST=http://host.docker.internal:11434 ghcr.io/raj-khan/rag-xray
```

## The course

Read it in the app (**Learn** tab) or right here on GitHub.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/media/learn-dark.webp">
  <img src="docs/media/learn-light.webp" alt="Lesson reader with sidebar, progress and table of contents" width="860">
</picture>

| # | Lesson | You will learn | Program |
|---|---|---|---|
| 0 | [What is RAG?](docs/lessons/00-what-is-rag.md) | Why models need retrieval; the two halves of RAG | |
| 1 | [Embeddings and similarity](docs/lessons/01-embeddings.md) | Text to vectors, cosine similarity, why you rank instead of threshold | [`01-embeddings`](lessons/01-embeddings/main.go) |
| 2 | [Chunking](docs/lessons/02-chunking.md) | Fixed vs structure-aware chunks, overlap, heading context | [`02-chunking`](lessons/02-chunking/main.go) |
| 3 | [Indexing and vector stores](docs/lessons/03-indexing.md) | Building an index; what vector databases add | [`03-index`](lessons/03-index/main.go) |
| 4 | [The RAG loop](docs/lessons/04-the-rag-loop.md) | Grounded prompts, citations, "I don't know" and its cost | [`04-ask`](lessons/04-ask/main.go) |
| 5 | [Hybrid search](docs/lessons/05-hybrid-search.md) | BM25, Reciprocal Rank Fusion, where each method fails | [`05-hybrid`](lessons/05-hybrid/main.go) |
| 6 | [Evaluation](docs/lessons/06-evaluation.md) | hit@k, MRR, answer grading, reading failures | [`06-eval`](lessons/06-eval/main.go) |
| 7 | [Beyond the basics](docs/lessons/07-beyond-the-basics.md) | Reranking, HyDE, agentic RAG, GraphRAG, long context vs RAG | reading map |

Run any lesson from the repo root, for example:

```sh
go run ./lessons/01-embeddings "how do I make pasta"
go run ./lessons/03-index                      # builds store/index.json
go run ./lessons/04-ask "What does KST-503 mean?"
go run ./lessons/06-eval -answers
```

## The X-ray playground

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/media/xray-dark.webp">
  <img src="docs/media/xray-light.webp" alt="X-ray view: documents and settings on the left, streamed answer with citations and retrieval columns on the right" width="860">
</picture>

Ask a question and you get:

| Stage | What you see |
|---|---|
| **R**etrieve | Top results from vector search, keyword search (BM25) and hybrid fusion, each with scores; for hybrid, each chunk's rank in both lists. Chunks that reached the prompt are highlighted. Click any chunk to read it. |
| **A**ugment | The exact system and user messages sent to the model. |
| **G**enerate | The streamed answer with clickable `[n]` citations, plus retrieval, first-token and total timings. |

Where methods disagree, you can see it. For "I'm ill and can't come to work", vector search finds *Sick leave*, but keyword search matches the word "work" elsewhere, and fusion pushes the right chunk down:

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/media/retrieval-dark.webp">
  <img src="docs/media/retrieval-light.webp" alt="Vector, keyword and hybrid result lists side by side, with each hybrid hit's rank in both lists" width="860">
</picture>

## Study anything

- **Your documents.** Drag `.md` or `.txt` files into the X-ray (notes, docs, a book) and ask them questions. From the CLI: `go run ./lessons/03-index -data ~/notes -out store/notes.json`, then `go run ./lessons/04-ask -index store/notes.json`. Start the app with them preloaded: `go run ./cmd/ragxray -docs ~/notes`.
- **Your own course.** Any folder of Markdown files becomes a course: `go run ./cmd/ragxray -lessons ./my-course`. See [writing a lesson](CONTRIBUTING.md#writing-a-lesson) for the optional front matter.
- **Shareable questions.** `http://localhost:8080/#/xray?q=your+question` opens the X-ray and asks right away.

<p align="center"><img src="docs/media/chunks-light.webp" alt="Chunk browser showing how each document was split" width="640">&nbsp;&nbsp;<img src="docs/media/mobile.webp" alt="Mobile layout" width="180"></p>

## Configuration

Everything is configured with environment variables, shared by the app, the lessons and Docker. With nothing set, everything runs locally on Ollama. See [`.env.example`](.env.example) for ready-made recipes.

| Variable | Default | Description |
|---|---|---|
| `EMBED_PROVIDER` | `ollama` | `ollama` or `openai` (any OpenAI-compatible API) |
| `EMBED_MODEL` | `nomic-embed-text` / `text-embedding-3-small` | Embedding model |
| `EMBED_BASE_URL` | `http://localhost:11434` / `https://api.openai.com/v1` | API endpoint |
| `EMBED_API_KEY` | falls back to `OPENAI_API_KEY` | Key for hosted embeddings |
| `CHAT_PROVIDER` | `ollama` | `ollama`, `openai` or `anthropic` |
| `CHAT_MODEL` | `llama3.2` / `gpt-4o-mini` / `claude-haiku-4-5` | Chat model |
| `CHAT_BASE_URL` | per provider | API endpoint |
| `CHAT_API_KEY` | falls back to `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` | Key for hosted chat |
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama address for both, if not overridden |
| `ADDR` | `127.0.0.1:8080` | Address the app listens on (`0.0.0.0:8080` in Docker) |

App flags: `-addr`, `-docs <folder>`, `-lessons <folder>`. Run `go run ./cmd/ragxray -h` for details.

> Changing the embedding model? Rebuild the index (lesson 3). The app re-indexes automatically; the CLI refuses to mix vectors from different models and tells you so.

## Results on the sample corpus

The sample documents describe a fictional company, so the model cannot know the answers without retrieval. With `nomic-embed-text` and `llama3.2` (3B) on 18 paraphrased questions:

| Retrieval | hit@1 | hit@4 | MRR |
|---|---|---|---|
| Vector | 0.83 | 1.00 | 0.90 |
| Keyword (BM25) | 0.78 | 0.94 | 0.85 |
| Hybrid (RRF) | 0.83 | 1.00 | 0.91 |

Answers: 15/18 correct. Every failure had the right passage in the prompt; [lesson 6](docs/lessons/06-evaluation.md) digs into why. Reproduce with `make eval`.

## Project layout

```text
cmd/ragxray        the web app (one binary, assets embedded)
lessons/NN-*       one small program per lesson
docs/lessons       lesson text, rendered by the app and readable on GitHub
internal/ai        providers: Ollama, OpenAI-compatible, Anthropic
internal/chunk     fixed-size and Markdown-aware chunking
internal/bm25      keyword search in ~80 lines
internal/retrieve  indexing, vector/keyword/hybrid search, prompt building
internal/server    HTTP API with a streamed X-ray endpoint
web/static         plain HTML, CSS and JS, no build step
data               sample corpus and labelled eval set
scripts            screenshot and demo capture
```

| Command | What it does |
|---|---|
| `make run` | Start the app on http://localhost:8080 |
| `make test` | Unit tests (no model needed) |
| `make lint` | gofmt check and `go vet` |
| `make index` / `make eval` | Build the CLI index / run the evaluation |
| `make up` | Everything in Docker, including Ollama |
| `make capture` | Regenerate screenshots and the demo |

## Troubleshooting

<details>
<summary><b>"cannot reach http://localhost:11434 (is the server running?)"</b></summary>

Ollama is not running. Start the Ollama app, or run `ollama serve`. In Docker on Linux, reach Ollama on the host with `--add-host=host.docker.internal:host-gateway -e OLLAMA_HOST=http://host.docker.internal:11434`.
</details>

<details>
<summary><b>"model not found" or a 404 from Ollama</b></summary>

Pull the models first: `ollama pull nomic-embed-text && ollama pull llama3.2`. If you set `CHAT_MODEL` or `EMBED_MODEL`, pull those instead.
</details>

<details>
<summary><b>"index was built with embedding model X but EMBED_MODEL is Y"</b></summary>

Vectors from different models are not comparable. Rebuild: `go run ./lessons/03-index`.
</details>

<details>
<summary><b>Answers are slow</b></summary>

On a CPU-only machine a 3B model can take 10 to 30 seconds per answer, and the first question also loads the model into memory. Use a smaller k, a GPU, or a hosted provider (see `.env.example`). Retrieval itself takes milliseconds; the X-ray timings show where time goes.
</details>

<details>
<summary><b>The model says "I don't know" although the right chunk was retrieved</b></summary>

Small models sometimes refuse when the wording differs ("vacation" vs "annual leave"). Lesson 4 measures this trade-off. Try a larger chat model, or compare with the no-retrieval toggle.
</details>

<details>
<summary><b>Port 8080 is already in use</b></summary>

`go run ./cmd/ragxray -addr 127.0.0.1:9090`, or change the port mapping in `compose.yaml`.
</details>

## Roadmap

- [ ] Lesson 8: reranking (cross-encoder and LLM-based) ([#10](https://github.com/raj-khan/rag-xray/issues/10))
- [ ] Contextual retrieval: LLM-written chunk context, measured against heading context ([#11](https://github.com/raj-khan/rag-xray/issues/11))
- [ ] PDF and HTML ingestion ([#12](https://github.com/raj-khan/rag-xray/issues/12))
- [ ] A persistent vector store option (for example sqlite-vec) with incremental updates ([#13](https://github.com/raj-khan/rag-xray/issues/13))
- [ ] LLM-as-judge answer grading in lesson 6 ([#14](https://github.com/raj-khan/rag-xray/issues/14))
- [ ] Translations of the course ([#17](https://github.com/raj-khan/rag-xray/issues/17))

Ideas and votes are welcome in [issues](https://github.com/raj-khan/rag-xray/issues).

## Contributing

Contributions of every size are welcome: fixing a typo, clarifying a lesson, sharing an experiment, adding a provider, improving the UI.

1. Read [CONTRIBUTING.md](CONTRIBUTING.md) for setup, layout and guidelines.
2. Pick an issue, especially ones labelled [`good first issue`](https://github.com/raj-khan/rag-xray/labels/good%20first%20issue), or open one to discuss your idea.
3. Fork, create a branch, and run `make lint test` before pushing.
4. Open a pull request; the template will guide you.

Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).

<a href="https://github.com/raj-khan/rag-xray/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=raj-khan/rag-xray" alt="Contributors">
</a>

## Security

rag-xray is a local learning tool without authentication. It listens on `127.0.0.1` by default; do not expose it to the internet. API keys stay on the server side and are never sent to the browser.

Found a vulnerability? Please **do not open a public issue**. Report it privately via [GitHub security advisories](https://github.com/raj-khan/rag-xray/security/advisories/new). See [SECURITY.md](SECURITY.md) for details.

## Learn more elsewhere

rag-xray covers the core pipeline. These excellent projects go further, and inspired parts of this one:

- [NirDiamant/RAG_Techniques](https://github.com/NirDiamant/RAG_Techniques): the biggest catalogue of advanced techniques
- [langchain-ai/rag-from-scratch](https://github.com/langchain-ai/rag-from-scratch): videos and notebooks
- [pguso/rag-from-scratch](https://github.com/pguso/rag-from-scratch): local-first RAG in JavaScript
- [microsoft/rag-time](https://github.com/microsoft/rag-time): a five-week course
- [Anthropic: Contextual Retrieval](https://www.anthropic.com/news/contextual-retrieval): the case for contextual chunks, hybrid search and reranking

Full credits in [ACKNOWLEDGEMENTS.md](ACKNOWLEDGEMENTS.md).

## License

[MIT](LICENSE). If rag-xray helped you, a ⭐ helps others find it.
