# Acknowledgements

This project stands on other people's work. Thank you.

## Learning resources that shaped it

- [NirDiamant/RAG_Techniques](https://github.com/NirDiamant/RAG_Techniques): the catalogue of advanced techniques that lesson 7 points to.
- [langchain-ai/rag-from-scratch](https://github.com/langchain-ai/rag-from-scratch): the "build it yourself" framing and video series.
- [pguso/rag-from-scratch](https://github.com/pguso/rag-from-scratch): the "local models, no black boxes" spirit, in JavaScript.
- [microsoft/rag-time](https://github.com/microsoft/rag-time): a structured course, including evaluation.
- [mlabonne/llm-course](https://github.com/mlabonne/llm-course): the wider LLM roadmap.

## Ideas and research used in the code

- Contextual chunk headers and hybrid search with fusion follow [Anthropic's Contextual Retrieval](https://www.anthropic.com/news/contextual-retrieval) write-up.
- Reciprocal Rank Fusion: Cormack, Clarke and Büttcher, SIGIR 2009.
- BM25: Robertson and Zaragoza, *The Probabilistic Relevance Framework: BM25 and Beyond*, 2009.
- Task prefixes for embeddings: [Nomic Embed](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5).
- Evaluation metrics inspired by [Ragas](https://github.com/explodinggradients/ragas).

## Tools

- [Ollama](https://ollama.com) for running models locally, and the open-weight models `nomic-embed-text` and `llama3.2`.

If your work influenced this project and is missing here, please open an issue or PR.
