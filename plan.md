# Plan

## TODO
- Implement JSONL enrichment mode (`--jsonl [name]`) with wrapping/merging logic.
- Update README and CLI help once JSONL mode is available.
- Add integration tests covering JSONL output alongside existing text flow.

## COMPLETED
- Designed brace-based template language with tokens for elapsed, delta, line, time, iso, and unix stamps.
- Integrated template parser into runtime and refactored line buffering/emission for unit testing.
- Expanded CLI to accept positional template and documented usage in README/help.
