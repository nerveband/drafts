# Drafts CLI

Command line interface for [Drafts](https://getdrafts.com). Requires Drafts Pro and macOS.

## Install

```bash
go install github.com/nerveband/drafts/cmd/drafts@latest
```

Or build from source:

```bash
git clone https://github.com/nerveband/drafts
cd drafts
go build ./cmd/drafts
```

**No additional dependencies required!** The CLI communicates directly with Drafts via AppleScript.

## Usage

```
$ drafts --help
Usage: drafts [--plain] <command> [<args>]

Options:
  --plain              output plain text instead of JSON
  --help, -h           display this help and exit

Commands:
  new, create          create new draft
  prepend              prepend to draft
  append               append to draft
  replace              replace content of draft
  edit                 edit draft in $EDITOR
  get                  get content of draft
  select               select active draft using fzf
  list                 list drafts
  run                  run a Drafts action
  schema               output tool-use schema for LLM integration
```

## Commands

### create / new

Create a new draft.

```bash
drafts create "Content here" [options]

Options:
  -t, --tag TAG        Add tag (can be used multiple times)
  -a, --archive        Create in archive folder
  -f, --flagged        Create as flagged
  --action ACTION      Run action after creation
```

### get

Get a draft by UUID.

```bash
drafts get [UUID]      # Omit UUID to get active draft
```

### list

List drafts with optional filtering.

```bash
drafts list [options]

Options:
  -f, --filter FILTER  Filter: inbox|archive|trash|all (default: inbox)
  -t, --tag TAG        Filter by tag (can be used multiple times)
```

### prepend / append

Add content to an existing draft.

```bash
drafts prepend "Text" -u UUID [options]
drafts append "Text" -u UUID [options]

Options:
  -u, --uuid UUID      Target draft UUID (omit to use active draft)
  -t, --tag TAG        Add tag
  --action ACTION      Run action after modification
```

### replace

Replace content of a draft.

```bash
drafts replace "New content" -u UUID
```

### edit

Open draft in your $EDITOR.

```bash
drafts edit [UUID]     # Omit UUID to edit active draft
```

### run

Run a Drafts action.

```bash
drafts run "Action Name" "Text to process"
drafts run "Action Name" -u UUID    # Run on existing draft
```

### schema

Output tool-use schema for LLM integration.

```bash
drafts schema          # Full schema
drafts schema create   # Schema for specific command
```

## Output Formats

**JSON (default)** - Structured output for programmatic use:
```bash
drafts list
```

**Plain text** - Human-readable output:
```bash
drafts list --plain
```

## LLM Integration

This CLI is designed for LLM tool use. Features:

- **JSON output by default** - Easy to parse
- **Structured errors** - Error code, message, and recovery hints
- **Tool-use schema** - Get schema with `drafts schema`
- **Full metadata** - All draft properties returned

### Example LLM Workflow

```bash
# 1. Get available commands schema
drafts schema

# 2. Create a draft
drafts create "Meeting notes for project X"

# 3. List recent drafts
drafts list

# 4. Get specific draft
drafts get <uuid>
```

## Implementation

The CLI communicates directly with Drafts via AppleScript (`osascript`).

**Architecture:**
- No helper apps required
- No Drafts actions to install
- Pure AppleScript communication
- Works on any Mac with Drafts Pro

## Development

```bash
go build ./cmd/drafts    # Build
go test ./...            # Run tests
go vet ./...             # Lint
```

## License

MIT
