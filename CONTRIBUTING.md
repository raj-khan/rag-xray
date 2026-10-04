# Contributing to rag-xray

Thanks for helping people learn RAG. Every kind of contribution is welcome: fixing a typo, clarifying a lesson, adding an experiment, supporting a new provider, or improving the UI.

## Ways to help

- **Improve a lesson.** If something confused you, it confuses others. Small clarity PRs are the most valuable kind.
- **Share an experiment.** Ran the capstone from lesson 7 on your own docs? Add your result to a lesson's "Experiments" section.
- **Translate** a lesson (add `docs/lessons/<lang>/`).
- **Add a provider** that is not OpenAI-compatible.
- **Fix bugs** or pick up an issue labelled `good first issue`.

## Setup

You need Go 1.27+ and either [Ollama](https://ollama.com) or an API key for any OpenAI-compatible provider (see `.env.example`).

```sh
git clone https://github.com/raj-khan/rag-xray && cd rag-xray
ollama pull nomic-embed-text && ollama pull llama3.2
make test    # unit tests, no model needed
make run     # http://localhost:8080
```

Or run everything in Docker with `docker compose up`.

## Project layout

| Path | What lives there |
|---|---|
| `lessons/NN-*/main.go` | One small runnable program per lesson |
| `docs/lessons/*.md` | Lesson text (front matter + Markdown), shown in the reader |
| `internal/ai` | Providers: Ollama, OpenAI-compatible, Anthropic |
| `internal/chunk`, `bm25`, `vec`, `retrieve` | The RAG building blocks |
| `internal/server`, `web/static` | The web app (plain HTML/CSS/JS, no build step) |
| `data/docs`, `data/eval.json` | Sample corpus and labelled eval set |

## Guidelines

- **Clarity over cleverness.** This is teaching code. Prefer the obvious 10 lines over the clever 3. Comments explain *why*.
- **Standard library first.** New dependencies need a strong reason.
- **Show real output.** When a lesson quotes output, paste what the program actually printed, including failures.
- **Test what you can offline.** Use fakes like `internal/server/server_test.go` instead of calling real models.
- **Credit others.** If an idea comes from a paper, post or project, link it and add it to `ACKNOWLEDGEMENTS.md`.
- Before pushing: `make lint test`.

## Writing a lesson

Create `docs/lessons/NN-title.md`:

```markdown
---
title: Reranking
order: 8
minutes: 20
run: go run ./lessons/08-rerank
summary: One sentence for the sidebar.
---

# Reranking

The idea, run it, what to notice, experiments, check yourself, further reading.
```

Use `<details><summary>Question</summary>Answer</details>` for self-check questions.

## Commits and pull requests

- Small, focused commits in the imperative mood: `Add`, `Fix`, `Update`, `Remove`, `Refactor`, `Test`.
- One logical change per PR, with a short description of what and why.
- Link the issue it closes, if any.

By contributing you agree that your contributions are licensed under the MIT License, and you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
