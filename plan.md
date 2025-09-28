# Plan

## TODO
(All planned features are complete)

## COMPLETED
- Designed brace-based template language with tokens for elapsed, delta, line, time, iso, and unix stamps.
- Integrated template parser into runtime and refactored line buffering/emission for unit testing.
- Expanded CLI to accept positional template and documented usage in README/help.
- Implemented JSONL enrichment mode (`--json [name]`) with wrapping/merging logic.
- Added comprehensive unit and integration tests covering JSONL output alongside existing text flow.
- Updated all documentation and CLI help with JSONL mode examples and usage.
