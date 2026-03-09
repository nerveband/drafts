# Drafts CLI Agent DX Audit

Date: 2026-03-09

This audit uses two references:

- Ship Types philosophy: the executable contract should live in types and machine-readable surfaces, not just prose docs.
- `agent-dx-cli-scale`: a 0-21 scoring rubric for agent-readiness.

## Result

Current score: **11/21**

Rating: **Agent-ready**

This repo is no longer just "JSON output plus a README". The command contract now lives in typed request structs and a generated schema surface, which is much closer to "ship types, not docs". The remaining gaps are mostly around safety rails, tighter validation, and context control on read paths.

## Scorecard

| Axis | Score | Notes |
| --- | --- | --- |
| Machine-readable output | 2/3 | JSON is the default and errors are structured. There is no NDJSON streaming or page-wise output. |
| Raw payload input | 3/3 | Mutating draft commands now accept `--input` JSON and the payload maps directly to typed request structs. |
| Schema introspection | 2/3 | `drafts schema` returns machine-readable parameter contracts for all commands. The schema is generated from local types, not a remote discovery document. |
| Context window discipline | 1/3 | `drafts list` defaults to summaries and supports `--limit` plus `--full`, but there are no field masks or streaming pagination. |
| Input hardening | 1/3 | Unknown JSON fields are rejected and invalid filters now fail fast, but IDs and names are not yet hardened against agent-specific malformed input patterns. |
| Safety rails | 0/3 | There is still no `--dry-run` for mutating commands and no response sanitization layer. |
| Agent knowledge packaging | 2/3 | The repo ships a structured `skills/SKILL.md` plus a schema surface. Guardrails are present, but not yet comprehensive or versioned as a broader skill library. |

## What Improved

- Removed unsupported syntax/language grammar behavior from the CLI surface.
- Aligned AppleScript property access with the live Drafts dictionary (`tag list`, `creation date`, `modification date`, latitude/longitude fields, workspace open support).
- Fixed `run -u` so actions run on the existing draft instead of copying content into a transient draft.
- Added `actions` and `workspace --open`.
- Added raw JSON input to mutating commands.
- Made schema output type-driven via `cmd/drafts/contracts.go` and `cmd/drafts/schema.go`.
- Made `drafts list` cheaper by default with summary output and `--limit`.
- Promoted structured error handling instead of swallowed package errors.
- Aligned docs and skill text with the actual module/repo identity.

## Ship Types Assessment

The repo is now materially closer to the "ship types, not docs" model:

- The CLI contract is encoded in Go structs, not hand-maintained prose tables.
- Schema output is generated from those structs, so agents can discover the same contract the code uses.
- README and `skills/SKILL.md` are now secondary surfaces that explain the contract instead of defining it.

The remaining weakness is that the contract is still local-only. It is typed, but not runtime-discovered from Drafts itself. That is acceptable for a desktop AppleScript CLI, but it limits the maximum score on schema introspection.

## Remaining Gaps

1. Add `--dry-run` to all mutating commands.
2. Add stronger validation for UUIDs, action names, and workspace names before AppleScript execution.
3. Add field selection or slimmer read modes beyond `list --full`.
4. Add page-wise or streaming output for large result sets.
5. Consider exposing a non-shell agent surface such as MCP or a local JSON-RPC mode.
