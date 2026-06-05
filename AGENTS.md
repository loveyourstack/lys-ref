# AGENTS

Guidance for AI coding agents working in this repository.

## Improvement Suggestion Policy

When suggesting or making improvements, prioritize in this order:

1. Correctness and regression risk.
2. Error handling and data safety.
3. Naming errors or inconsistencies.
4. Idiomatic patterns.
5. Simplicity and readability.
6. Missing or weak tests.
7. Performance.

Prefer one clear, obvious approach over configurable or extensible designs.

Breaking changes are allowed unless I specify otherwise.

## Testing Expectations

- Prefer targeted tests first, then full suite when risk is broader.
- Keep table-driven style where already used.
