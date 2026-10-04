---
title: Evaluation
order: 6
minutes: 20
run: go run ./lessons/06-eval -answers
summary: Measure retrieval and answers with a labelled question set, so every change is a measurement instead of a guess.
---

# Evaluation

## Without numbers, you are guessing

Every knob in this course (chunk size, overlap, k, hybrid vs vector, prompt wording, model) changes results in ways that single examples hide. A **small labelled set** of questions turns tuning into measurement.

`data/eval.json` has 18 questions, each labelled with the file that holds the answer and a string the answer must contain:

```json
{"q": "What is the food budget per day when travelling abroad?", "source": "expenses.md", "answer": "70"}
```

Notice the questions are **paraphrased** ("food budget", "abroad") rather than copied from the text ("meal allowance", "international"). Real users do not use your document's vocabulary.

## Two layers, two sets of metrics

**Retrieval**: did the right source reach the prompt?

- **hit@1**: the right file was ranked first.
- **hit@k**: the right file was anywhere in the top k (what the LLM actually sees).
- **MRR** (mean reciprocal rank): average of 1/rank of the first correct result; 1.0 is perfect.

**Generation**: given the context, was the answer right? Here we check the expected text appears. Larger projects use an LLM as a judge for faithfulness and relevance.

## Run it

```sh
go run ./lessons/06-eval            # retrieval only, a few seconds
go run ./lessons/06-eval -answers   # also generates and grades answers
```

Real results with `nomic-embed-text` and `llama3.2`:

```text
mode      hit@1  hit@k    MRR
vector     0.83   1.00   0.90
keyword    0.78   0.94   0.85
           miss: How quickly must I respond when I get paged at night?
hybrid     0.83   1.00   0.91

Answer accuracy: 15/18
```

## Read the failures, not just the score

All three answer failures had **perfect retrieval**:

```text
FAIL What is the food budget per day when travelling abroad?
     -> I don't know based on the documents. The context passages do not mention a specific
        food budget for traveling abroad.
FAIL I bought a 700 dollar monitor for a client demo. Who has to sign off?
     -> I don't know based on the documents. There is no information about signing off for
        expenses related to client demos or equipment purchases.
FAIL Can my friend come to the Lisbon office with me?
     -> According to [2], visitors must be accompanied by an employee at all times, so your
        answer would depend on whether you are accompanying your friend or not.
```

- The first two are **false refusals**: the 3B model did not connect "food budget" with "meal allowance", or a 700 dollar purchase with "any single expense above 500 USD". Lesson 4 shows how the "I don't know" instruction trades hallucinations for refusals.
- The third is arguably **correct**, but our grader wanted the word "Perch" (visitors must be registered in Perch). That is a *grader* failure, which is why you always read the failures instead of trusting the number.

Retrieval was fine; generation failed. The fixes are on the generation side: a bigger chat model, fewer distracting chunks (lower k, or a reranker), or a prompt that asks the model to quote the passage before answering.

## Experiments

1. Change one thing at a time and re-run: chunk `-max` in lesson 3, `-k` here, or the chat model. Keep a small table of results.
2. Remove the task prefixes for nomic in `internal/retrieve` (`prefixes` returns `"", ""`), rebuild the index and compare MRR.
3. Write 10 questions for your own documents and evaluate them. This is the single most useful thing you can do for a real RAG project.

## Check yourself

<details>
<summary>hit@k is 1.00 but answer accuracy is 15/18. Where should you spend your effort?</summary>

On generation: prompt, model, or context noise. Retrieval already delivers the right file every time.
</details>

<details>
<summary>Why is checking "answer contains '70'" a weak grader?</summary>

It passes "not 70" and fails "seventy". It is fine for a fast smoke test with carefully chosen strings. For anything serious, use an LLM judge with a rubric, and spot-check its verdicts by hand.
</details>

## Further reading

- [Ragas](https://github.com/explodinggradients/ragas): metrics such as faithfulness, answer relevance and context precision.
- [Microsoft RAG Time](https://github.com/microsoft/rag-time) has a module on evaluation.
