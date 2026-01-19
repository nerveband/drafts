# Drafts CLI - Feature Gap Analysis & Improvement Plan

A comprehensive comparison between the current CLI implementation and the official Drafts URL Scheme documentation, plus LLM-friendliness improvements.

---

## Executive Summary

**Currently Implemented:** 8 commands (new, prepend, append, replace, edit, get, select, list)
**Total URL Scheme Actions Available:** 19 actions
**Missing Actions:** 11 complete actions + action execution not exposed
**Incomplete Parameters:** Multiple parameters missing from existing commands
**LLM-Friendliness:** Needs significant improvements for machine consumption

---

## Part 1: Missing Features

### CURRENTLY IMPLEMENTED COMMANDS

| Command | CLI Command | Status | Notes |
|---------|-------------|--------|-------|
| `/create` | `new` | Partial | Missing `--action`, `--allow-empty` |
| `/prepend` | `prepend` | Partial | Missing `--action`, `--tag` |
| `/append` | `append` | Partial | Missing `--action`, `--tag` |
| `/getCurrentDraft` | (internal) | Full | Used for "active draft" resolution |
| `/runAction` | **NOT EXPOSED** | Internal only | Function exists but no CLI command |

---

### ACTION-RELATED FEATURES (High Priority)

The library has a `RunAction()` function but it's **not exposed as a CLI command**. This is a major gap.

#### 1. `run-action` Command (NEW - HIGH PRIORITY)

**Purpose:** Execute any Drafts action on text or a draft

```bash
# Run action on inline text
drafts run-action "Markdown to HTML" "# Hello World"

# Run action on piped input
echo "# Hello" | drafts run-action "Markdown to HTML"

# Run action on existing draft (by UUID)
drafts run-action "Export to PDF" --uuid ABC123

# Run action on active draft
drafts run-action "Share to Twitter"
```

**Parameters:**
- `action` (positional, required) - Name of the action to run
- `text` (positional, optional) - Text to process (or stdin)
- `--uuid` / `-u` - Run on existing draft instead of text
- `--allow-empty` - Allow action to run even if text is empty

**Implementation:** Already exists as `drafts.RunAction()` in `pkg/drafts/drafts.go:113`

---

#### 2. `--action` Flag on Modification Commands

Add `--action` flag to `new`, `prepend`, `append` to chain action execution:

```bash
# Create draft and immediately run an action
drafts new "Meeting notes" --action "Add to Calendar"

# Append and then process
drafts append "New item" --uuid ABC123 --action "Sort Lines"
```

---

#### 3. Action Group Commands

```bash
# Load action group in side panel
drafts load-action-group "Markdown"

# Load action group in action bar
drafts load-action-bar "Quick Actions"

# Search actions
drafts action-search "export"
```

---

### MISSING URL SCHEME COMMANDS

#### Priority 1: High Value

| URL Action | Proposed CLI | Description |
|------------|--------------|-------------|
| `/open` | `open` | Open draft by UUID or title |
| `/search` | `search` | Open search with query |
| `/workspace` | `workspace` | Load a workspace |
| `/runAction` | `run-action` | Execute action (internal exists, not exposed) |

#### Priority 2: Navigation/UI

| URL Action | Proposed CLI | Description |
|------------|--------------|-------------|
| `/quickSearch` | `quick-search` | Open quick search |
| `/commandPalette` | `command-palette` | Open command palette |
| `/actionSearch` | `action-search` | Search actions |
| `/loadActionGroup` | `load-action-group` | Load action group |
| `/loadActionBarGroup` | `load-action-bar` | Load action bar group |

#### Priority 3: Content Manipulation

| URL Action | Proposed CLI | Description |
|------------|--------------|-------------|
| `/replaceRange` | `replace-range` | Replace at character position |
| `/capture` | `capture` | Open capture window |

#### Priority 4: Input Methods

| URL Action | Proposed CLI | Description |
|------------|--------------|-------------|
| `/dictate` | `dictate` | Open dictation |
| `/arrange` | `arrange` | Open arrange interface |

---

### MISSING DRAFT PROPERTIES (from Scripting API)

The current `get` command only returns basic properties. The Draft class has many more:

**Currently Returned:**
- `uuid`, `content`, `tags`, `isFlagged`, `isArchived`, `isTrashed`

**Missing Properties:**
- [ ] `createdAt` - Creation timestamp
- [ ] `modifiedAt` - Last modified timestamp
- [ ] `title` / `displayTitle` - First line / display title
- [ ] `syntax` - Syntax definition
- [ ] `folder` - Current folder location
- [ ] `permalink` - Permanent link URL
- [ ] `versions` - Version history
- [ ] `actionLogs` - Action execution history
- [ ] `tasks` / `completedTasks` / `incompleteTasks` - Task lists
- [ ] `linkedItems` - Wiki-style cross-links
- [ ] `urls` - Extracted URLs from content

---

### MISSING GLOBAL FEATURES

- [ ] `||clipboard||` markup support in text parameters
- [ ] `--allow-empty` flag for action execution
- [ ] Find draft by title (`Draft.queryByTitle()`)
- [ ] Version management (`saveVersion()`, list versions)
- [ ] Task management (complete, reset, advance tasks)
- [ ] Template processing (`processTemplate()`)

---

## Part 2: LLM-Friendliness Improvements

### Current Problems for LLM Consumption

1. **No JSON output** - All output is plain text, hard to parse
2. **Inconsistent output formats** - `list` uses tabs, `get` returns raw content
3. **Minimal help text** - No examples, no parameter descriptions
4. **No schema documentation** - LLMs can't discover available options
5. **Error handling** - Errors not structured for machine parsing
6. **No introspection** - Can't query available actions, tags, workspaces

---

### Proposed LLM-Friendly Features

#### 1. Global `--json` / `--output` Flag

Add structured JSON output to ALL commands:

```bash
# Current (hard to parse)
$ drafts list
ABC123	Meeting notes from Monday
DEF456	Shopping list

# With --json (easy to parse)
$ drafts list --json
{
  "drafts": [
    {
      "uuid": "ABC123",
      "title": "Meeting notes from Monday",
      "content": "Meeting notes from Monday\n- Item 1\n- Item 2",
      "tags": ["work", "meetings"],
      "isFlagged": false,
      "isArchived": false,
      "createdAt": "2024-01-15T10:30:00Z",
      "modifiedAt": "2024-01-15T11:45:00Z"
    }
  ],
  "count": 2,
  "filter": "inbox"
}
```

---

#### 2. `--help` with Examples

Enhance help text with usage examples:

```bash
$ drafts new --help
Usage: drafts new [OPTIONS] [MESSAGE]

Create a new draft in Drafts.app

Arguments:
  MESSAGE    Draft content. Omit to read from stdin.

Options:
  -t, --tag TAG        Add tag (repeatable)
  -a, --archive        Create in archive folder
  -f, --flagged        Create as flagged
  --action ACTION      Run action after creation
  --allow-empty        Allow action on empty draft
  --json               Output result as JSON

Examples:
  # Create a simple draft
  drafts new "Hello, world!"

  # Create with tags
  drafts new "Meeting notes" -t work -t meetings

  # Create from stdin
  echo "Piped content" | drafts new

  # Create and run action
  drafts new "# Markdown" --action "Preview"

  # JSON output (for scripting/LLMs)
  drafts new "Test" --json
  # Returns: {"uuid": "ABC123", "content": "Test", ...}
```

---

#### 3. `schema` Command for Introspection

Allow LLMs to discover available commands and options:

```bash
$ drafts schema
{
  "version": "0.2.0",
  "commands": {
    "new": {
      "description": "Create a new draft",
      "arguments": [
        {"name": "message", "type": "string", "required": false, "description": "Draft content"}
      ],
      "options": [
        {"name": "tag", "short": "t", "type": "string[]", "description": "Add tags"},
        {"name": "archive", "short": "a", "type": "boolean", "description": "Create in archive"},
        {"name": "flagged", "short": "f", "type": "boolean", "description": "Create flagged"},
        {"name": "action", "type": "string", "description": "Action to run after creation"},
        {"name": "json", "type": "boolean", "description": "Output as JSON"}
      ],
      "returns": {"type": "uuid", "description": "UUID of created draft"}
    },
    ...
  }
}

$ drafts schema new
# Returns schema for just the 'new' command
```

---

#### 4. `info` Command for Environment Discovery

```bash
$ drafts info --json
{
  "drafts_running": true,
  "drafts_version": "42.0",
  "pro_subscription": true,
  "helper_installed": true,
  "available_actions": ["Markdown to HTML", "Export to PDF", ...],
  "available_workspaces": ["Default", "Work", "Personal"],
  "available_tags": ["work", "personal", "urgent"],
  "inbox_count": 42,
  "flagged_count": 5
}
```

---

#### 5. Structured Error Output

```bash
# Current
$ drafts get INVALID
# (crashes or prints unclear error)

# Proposed with --json
$ drafts get INVALID --json
{
  "error": true,
  "code": "DRAFT_NOT_FOUND",
  "message": "No draft found with UUID: INVALID",
  "suggestion": "Use 'drafts list' to see available drafts"
}
```

---

#### 6. Consistent Naming Conventions

Current inconsistencies:
- `new` vs `create` (URL scheme uses `create`)
- `get` (retrieves content) vs `list` (lists drafts)

Proposed aliases for clarity:
```bash
drafts create  # alias for 'new'
drafts show    # alias for 'get'
drafts find    # alias for 'list' with query
```

---

#### 7. Query/Filter DSL

Make complex queries easier:

```bash
# Current (limited)
drafts list -f inbox -t work

# Proposed (richer query language)
drafts query "tag:work AND tag:urgent AND NOT tag:done"
drafts query "modified:>2024-01-01 AND flagged:true"
drafts query "content:meeting AND folder:inbox"
```

---

#### 8. Batch Operations

```bash
# Process multiple drafts
drafts batch --filter "tag:process" --action "Archive"

# Get multiple drafts by UUID
drafts get UUID1 UUID2 UUID3 --json

# Tag multiple drafts
drafts tag --add urgent --filter "content:deadline"
```

---

### Documentation Improvements

#### 1. Man Page / Comprehensive Help

```bash
$ drafts help
# Opens comprehensive documentation

$ drafts help new
# Detailed help for 'new' command with examples
```

#### 2. Machine-Readable Documentation

```bash
$ drafts docs --format markdown > COMMANDS.md
$ drafts docs --format json > commands.json
```

---

## Implementation Plan

### Phase 1: Quick Wins (Action Support)

**Goal:** Expose existing functionality, add critical missing features

1. **Add `run-action` command** (function already exists!)
   - Files: `cmd/drafts/main.go`
   - Expose `drafts.RunAction()` as CLI command

2. **Add `--action` flag to `new`, `prepend`, `append`**
   - Files: `cmd/drafts/main.go`, `pkg/drafts/drafts.go`

3. **Add `--tag` flag to `prepend` and `append`**
   - Files: Same as above

### Phase 2: LLM-Friendliness Foundation

**Goal:** Make output machine-parseable

4. **Add global `--json` flag**
   - Files: `cmd/drafts/main.go`
   - JSON output for all commands

5. **Enhance `get` to return full draft metadata**
   - Files: `pkg/drafts/js/get.js`
   - Add timestamps, syntax, permalink, etc.

6. **Add structured error handling**
   - Files: Throughout

7. **Add `schema` command**
   - Files: New file `cmd/drafts/schema.go`

### Phase 3: New High-Value Commands

8. **Implement `open` command**
9. **Implement `search` command**
10. **Implement `workspace` command**
11. **Implement `info` command**

### Phase 4: Documentation & Polish

12. **Enhanced `--help` with examples**
13. **Add command aliases** (`create`, `show`, `find`)
14. **Add man page generation**
15. **Update README with comprehensive examples**

### Phase 5: Advanced Features

16. **Query DSL** for complex filtering
17. **Batch operations**
18. **Action group commands**
19. **Version management**
20. **Task management**

---

## File Change Summary

| File | Changes |
|------|---------|
| `cmd/drafts/main.go` | Add new commands, flags, JSON output |
| `cmd/drafts/schema.go` | NEW - schema introspection |
| `cmd/drafts/output.go` | NEW - unified output formatting |
| `pkg/drafts/drafts.go` | Add new URL scheme functions |
| `pkg/drafts/options.go` | Add new option structs |
| `pkg/drafts/js/get.js` | Return additional properties |
| `pkg/drafts/js/info.js` | NEW - environment info |
| `README.md` | Comprehensive documentation update |

---

## Example: Complete LLM Workflow

```bash
# 1. Discover available commands
$ drafts schema --json | jq '.commands | keys'
["new", "get", "list", "append", "prepend", ...]

# 2. Get details about a specific command
$ drafts schema new --json
{"arguments": [...], "options": [...], "examples": [...]}

# 3. Create a draft with JSON response
$ drafts new "LLM-generated content" --tag ai --json
{"uuid": "ABC123", "content": "LLM-generated content", ...}

# 4. Run an action on it
$ drafts run-action "Markdown to HTML" --uuid ABC123 --json
{"success": true, "output": "<p>LLM-generated content</p>"}

# 5. Query drafts with structured output
$ drafts list --filter inbox --tag ai --json
{"drafts": [...], "count": 5}
```

---

## Priority Matrix

| Feature | Effort | Impact | Priority |
|---------|--------|--------|----------|
| `run-action` command | Low | High | **P0** |
| `--json` flag | Medium | Very High | **P0** |
| `--action` on commands | Low | High | **P1** |
| `schema` command | Medium | High | **P1** |
| Enhanced `--help` | Low | Medium | **P1** |
| `open` command | Low | High | **P1** |
| `search` command | Low | Medium | **P2** |
| `workspace` command | Low | Medium | **P2** |
| `info` command | Medium | Medium | **P2** |
| Query DSL | High | Medium | **P3** |
| Batch operations | High | Medium | **P3** |

---

## References

- [Drafts URL Schemes Documentation](https://docs.getdrafts.com/docs/automation/urlschemes.html)
- [Drafts Scripting Reference](https://scripting.getdrafts.com/classes/Draft)
- [x-callback-url Specification](http://x-callback-url.com/specifications/)
- [Original Repository](https://github.com/ernstwi/drafts)
