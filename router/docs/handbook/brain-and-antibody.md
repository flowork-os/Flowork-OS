# The Brain & the Anti-Hallucination Antibody

Flow Router carries a local **knowledge brain** — a SQLite FTS5 "Memory Palace" for fast, offline
retrieval — and a **mistakes journal**. Together they power the **Anti-Hallucination Antibody**:
before the model answers, the router pulls the most relevant past mistakes, ranked by *karma* (how
often a mistake recurred) **×** *relevance*, and injects the top few into the prompt. A hallucination
gets harder to repeat over time — deterministically, with no retraining and no GPU.

This is the runtime engine of *Educational Errors*. Read the dated blueprints:
- [Anti-Hallucination Antibody](https://github.com/flowork-os/doc/blob/main/ANTI_HALLUCINATION_ANTIBODY.md)
- [Educational Errors](https://github.com/flowork-os/doc/blob/main/EDUCATIONAL_ERRORS.md)
