# Security Policy

## Supported versions

Only the latest release and the `main` branch receive fixes.

## Reporting a vulnerability

Please **do not open a public issue**. Report it privately through GitHub:
**Security → Report a vulnerability** on this repository ([direct link](https://github.com/raj-khan/rag-xray/security/advisories/new)).

Include steps to reproduce, the impact, and the version or commit. You can expect an acknowledgement within a few days, and credit in the release notes once a fix ships, if you want it.

## Using rag-xray safely

rag-xray is a **local learning tool**. It has no authentication.

- By default it listens on `127.0.0.1` only. Do not expose it to the internet: anyone who can reach it can upload documents and spend your model credits.
- API keys are read from environment variables and never sent to the browser. Keep `.env` out of version control (it is in `.gitignore`).
- Uploaded documents live in memory only and are sent to the embedding and chat providers you configure. Do not upload sensitive documents to a hosted provider you do not trust.
- Retrieved text is untrusted input to the model (prompt injection). Treat answers about untrusted documents with care, as lesson 7 explains.
