# Drafts CLI: LLM-First Redesign

**Date:** 2026-01-19
**Status:** Approved

## Goals

**Primary:** LLM-first tool - optimized for AI agents calling this CLI
**Secondary:** Complete URL scheme coverage, power user CLI, good docs

**Workflows supported:** Create & process, query & analyze, orchestrate - all equally

**Discovery method:** Schema command for system prompts + good docs. MCP planned as future upgrade.

## Core Architecture

### Philosophy

The CLI is a thin wrapper around Drafts URL schemes, optimized for LLM consumption. Every command outputs structured JSON by default, fails fast with actionable errors, and is fully described via a tool-use schema.

### Output Modes

- **JSON is the default** (not a flag) - this is an LLM-first tool
- Add `--plain` flag for human-readable output when needed
- All responses include full draft metadata

### Error Contract

```json
{
  "success": false,
  "error": {
    "code": "DRAFT_NOT_FOUND",
    "message": "No draft found with UUID: ABC123",
    "hint": "Use 'drafts list' to see available drafts"
  }
}
```

### Success Contract

```json
{
  "success": true,
  "draft": {
    "uuid": "ABC123",
    "content": "...",
    "title": "First line",
    "tags": ["work"],
    "createdAt": "2024-01-15T10:30:00Z",
    "modifiedAt": "2024-01-15T11:45:00Z",
    "isFlagged": false,
    "isArchived": false,
    "isTrashed": false,
    "folder": "inbox"
  }
}
```

## Schema & Introspection

The `drafts schema` command outputs tool-use formatted definitions for LLM system prompts.

### Format

```json
{
  "name": "drafts",
  "version": "0.2.0",
  "tools": [
    {
      "name": "drafts_create",
      "description": "Create a new draft in Drafts.app",
      "parameters": {
        "type": "object",
        "properties": {
          "content": {
            "type": "string",
            "description": "The draft content. Required unless reading from stdin."
          },
          "tags": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Tags to apply to the draft"
          },
          "folder": {
            "type": "string",
            "enum": ["inbox", "archive"],
            "default": "inbox"
          },
          "flagged": {
            "type": "boolean",
            "default": false
          },
          "action": {
            "type": "string",
            "description": "Action name to run after creation"
          }
        },
        "required": []
      }
    }
  ]
}
```

### Usage

- `drafts schema` - Full schema for all commands
- `drafts schema create` - Schema for single command
- `drafts schema | pbcopy` - Copy to clipboard for system prompts

### Naming Convention

Tool names use `drafts_<command>` format (e.g., `drafts_create`, `drafts_append`) to namespace clearly.

## Command Set

### Create & Modify

| Command | Description | Key Params |
|---------|-------------|------------|
| `create` | New draft | `content`, `tags`, `folder`, `flagged`, `action` |
| `append` | Add to end | `uuid`, `content`, `tags`, `action` |
| `prepend` | Add to start | `uuid`, `content`, `tags`, `action` |
| `replace` | Replace content | `uuid`, `content` |
| `replaceRange` | Replace at position | `uuid`, `content`, `start`, `length` |

### Read & Query

| Command | Description | Key Params |
|---------|-------------|------------|
| `get` | Get single draft | `uuid` (or active draft) |
| `list` | List drafts | `folder`, `tags`, `limit` |
| `search` | Open search UI | `query`, `tag`, `folder` |

### Actions

| Command | Description | Key Params |
|---------|-------------|------------|
| `run` | Execute action | `action`, `content` or `uuid` |

### Navigation

| Command | Description | Key Params |
|---------|-------------|------------|
| `open` | Open draft in UI | `uuid` or `title` |
| `workspace` | Load workspace | `name` |
| `capture` | Open capture window | `content`, `tags` |

### Introspection

| Command | Description |
|---------|-------------|
| `schema` | Output tool-use schema |
| `info` | Environment info (actions, workspaces, tags available) |

## Environment Discovery (`info` command)

```json
{
  "success": true,
  "environment": {
    "drafts_running": true,
    "pro_subscription": true,
    "helper_installed": true
  },
  "actions": [
    {"name": "Markdown to HTML", "group": "Markdown"},
    {"name": "Export to PDF", "group": "Export"},
    {"name": "Send to Calendar", "group": "Integrations"}
  ],
  "workspaces": ["Default", "Work", "Personal", "Writing"],
  "tags": ["work", "personal", "urgent", "draft", "published"],
  "counts": {
    "inbox": 42,
    "flagged": 5,
    "archive": 128,
    "trash": 3
  }
}
```

## Implementation Phases

### Phase 1: Foundation

1. Invert output default: JSON by default, `--plain` for humans
2. Add `schema` command with tool-use format
3. Expose `run` command (already exists internally)
4. Add `--action` flag to `create`, `append`, `prepend`
5. Add `--tags` flag to `append`, `prepend`
6. Structured error responses with hints

### Phase 2: Enhanced Metadata & Info

7. Expand `get` to return full draft metadata
8. Add `info` command for environment discovery
9. Enhance `list` to return full metadata per draft

### Phase 3: Complete URL Scheme Coverage

10. Add `open` command
11. Add `search` command
12. Add `workspace` command
13. Add `replaceRange` command
14. Add `capture` command
15. Add navigation commands (quickSearch, commandPalette, actionSearch)

### Phase 4: Documentation & Polish

16. Comprehensive README with LLM usage examples
17. Enhanced `--help` with examples per command
18. Example system prompts for Claude, GPT, etc.

### Future (Deferred)

- Query DSL for complex filtering
- Batch operations
- Version management
- Task management

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Default output | JSON | LLM-first tool |
| Schema format | Tool-use style | Multi-LLM compatible |
| Metadata | Full by default | Extra tokens negligible vs. context value |
| Errors | Fail fast + hints | Clear signal + recovery guidance |
| Naming | Follow Drafts terminology | Schema is source of truth anyway |
| MCP | Deferred | Future upgrade, CLI-first now |

## References

- [Drafts URL Schemes](https://docs.getdrafts.com/docs/automation/urlschemes.html)
- [Drafts Scripting Reference](https://scripting.getdrafts.com/classes/Draft)
- [Original Repository](https://github.com/ernstwi/drafts)
