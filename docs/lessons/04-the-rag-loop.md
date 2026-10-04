---
title: The RAG loop
order: 4
minutes: 20
run: go run ./lessons/04-ask "How many vacation days do I get?"
summary: Retrieve, build a grounded prompt with numbered sources, and generate an answer with citations.
---

# The RAG loop

## Three steps, about 30 lines

```go
// 1. Retrieve
hits, _ := r.Search(question, "hybrid", 4)

// 2. Augment: "[1] (file > heading)\ntext" for each hit, then the question
user := retrieve.Prompt(question, hits)

// 3. Generate
r.LLM.Chat([]ai.Message{
	{Role: "system", Content: retrieve.SystemPrompt},
	{Role: "user", Content: user},
}, print)
```

## The system prompt does three jobs

```text
You are a helpful assistant answering questions about the user's documents.
Answer the question using ONLY the numbered context passages.
Cite the passages you used like [1] or [2].
If the context does not contain the answer, say "I don't know based on the documents."
Keep answers short.
```

1. **Grounding**: use the passages, not memory.
2. **Citations**: numbered sources let readers verify the answer.
3. **An escape hatch**: without explicit permission to say "I don't know", models tend to make something up.

## Run it: with and without retrieval

```sh
go run ./lessons/04-ask "How many vacation days do I get?"
go run ./lessons/04-ask -no-rag "How many vacation days do I get at Orbita Labs?"
go run ./lessons/04-ask "What is the CEO's name?"
go run ./lessons/04-ask -show-prompt -k 2 "What does KST-503 mean?"
go run ./lessons/04-ask          # interactive
```

What `llama3.2` (3B) answered:

| Question | Answer |
|---|---|
| What does KST-503 mean? (RAG) | "According to [1], KST-503 means the API is in scheduled maintenance, specifically on Sundays from 02:00 to 04:00 UTC." |
| Vacation days (no RAG) | "I don't have information about your employer, Orbita Labs..." |
| CEO's name (RAG) | "I don't know based on the documents." |

The CEO is never mentioned in the documents, so "I don't know" is the **correct** answer. A RAG system that admits ignorance is worth far more than one that guesses.

## The escape hatch has a price

Ask `How many vacation days do I get?` and the 3B model may reply:

```text
I don't know based on the documents. The context only mentions "Annual leave" and does not
specify whether this is referring to paid annual leave or vacation days specifically.
```

The right passage was retrieved, but the model took "vacation" ≠ "annual leave" literally and used its permission to refuse. We measured three prompt variants on the eval set (lesson 6):

| System prompt | Answers correct |
|---|---|
| An earlier, shorter prompt: "Answer using only the context. One short sentence." (no "I don't know" instruction, no headings in the context) | 16 / 18 |
| Strict "only if nothing is relevant, say I don't know" | 14 / 18 |
| The prompt above | 15 / 18 |

More permission to refuse means fewer invented answers but more false refusals, and smaller models feel this most. Larger models usually handle paraphrases better, so try a hosted one (for example `CHAT_PROVIDER=anthropic`) and compare. Measure it rather than guess.

## Always look at the prompt

Run with `-show-prompt`. Most RAG bugs are visible right there: the wrong chunks, a chunk cut mid-sentence, the right chunk at position 4 where a small model ignores it. When an answer is wrong, first ask: **was the answer in the prompt?**

- **No** → a retrieval problem (lessons 2, 3, 5).
- **Yes** → a generation problem: prompt wording, model size, or too much noise in the context.

## Experiments

1. Ask the same question with `-k 1` and `-k 8`. When does more context help, and when does it confuse the model?
2. Delete the "I don't know" line from the system prompt and ask about the CEO again.
3. Switch the chat model (for example `CHAT_PROVIDER=anthropic` or a bigger Ollama model) and compare answers on the hardest eval questions.
4. Try `-mode vector` and `-mode keyword` on "I left my laptop on the train".

## Check yourself

<details>
<summary>An answer is wrong but the right chunk was retrieved. Name three possible causes.</summary>

The model ignored it (small models often favour the first passages), other chunks contradicted or distracted it, or the prompt did not insist on using only the context. A larger model, fewer but better chunks (reranking), or a clearer prompt each help.
</details>

<details>
<summary>Why temperature 0?</summary>

It makes answers as repeatable as possible, which you need in order to compare changes fairly. Creativity is not the goal when the job is quoting documents.
</details>
