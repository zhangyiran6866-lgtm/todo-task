# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Layout

This repo is configured as `single-context`.

- Preferred context file: `CONTEXT.md` at the repo root.
- ADR directory: `docs/adr/` at the repo root.

If these files/directories do not exist yet, proceed silently and continue implementation work.

## Consumer rules

- Read `CONTEXT.md` before deep exploration when it exists.
- Read related ADRs under `docs/adr/` when they exist.
- Use glossary terms from `CONTEXT.md` consistently in issues, plans, and refactor proposals.
- If a proposal conflicts with an ADR, call out the conflict explicitly instead of silently overriding.
