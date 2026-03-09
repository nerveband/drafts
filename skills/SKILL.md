---
name: drafts
description: Manage Drafts app notes via CLI on macOS. Create, view, list, edit, append, prepend, flag/unflag, manage workspaces, inspect actions, and run actions on drafts. Supports raw JSON payloads on mutating commands. Use when a user asks to create a note, list drafts, search drafts, flag/unflag drafts, run Drafts actions, or manage their Drafts inbox. IMPORTANT - Drafts app must be running on macOS for this to work.
homepage: https://github.com/nerveband/drafts-applescript-cli
metadata: {"clawdbot":{"emoji":"📋","os":["darwin"],"requires":{"bins":["drafts"]}}}
---

# Drafts CLI

Manage [Drafts](https://getdrafts.com) notes from the terminal on macOS.

## IMPORTANT REQUIREMENTS

> **This CLI ONLY works on macOS with Drafts app running.**

- **macOS only** - Uses AppleScript, will not work on Linux/Windows
- **Drafts must be RUNNING** - The app must be open for any command to work
- **Drafts Pro required** - Automation features require Pro subscription

If commands fail or hang, first check: `open -a Drafts`

## Setup

Install via Go:
```bash
go install github.com/nerveband/drafts-applescript-cli/cmd/drafts@latest
```

Or build from source:
```bash
git clone https://github.com/nerveband/drafts-applescript-cli
cd drafts-applescript-cli && go build ./cmd/drafts
```

## Commands

### Create a Draft

```bash
# Simple draft
drafts create "Meeting notes for Monday"

# With tags
drafts create "Shopping list" -t groceries -t todo

# Flagged draft
drafts create "Urgent reminder" -f

# Create in archive
drafts create "Reference note" -a

# Raw JSON payload
drafts create --input '{"content":"Agent-safe input","tags":["ai","cli"]}'
```

### List Drafts

```bash
# List inbox (default)
drafts list

# List archived drafts
drafts list -f archive

# List trashed drafts
drafts list -f trash

# List all drafts
drafts list -f all

# Filter by tag
drafts list -t mytag

# Search by content
drafts list -s "search term"

# Filter by workspace
drafts list -w "My Workspace"

# Combine filters
drafts list -f inbox -t work -s "meeting"

# Reduce response size (default is already 20)
drafts list --limit 5

# Include full content and coordinates
drafts list --full -t work
```

### Get a Draft

```bash
# Get specific draft
drafts get <uuid>

# Get active draft (currently open in Drafts)
drafts get
```

Returns full draft metadata including: uuid, content, title, tags, folder, flagged status, dates, location coordinates, and permalink.

### Modify Drafts

```bash
# Prepend text
drafts prepend "New first line" -u <uuid>

# Append text
drafts append "Added at the end" -u <uuid>

# Replace entire content
drafts replace "Completely new content" -u <uuid>
```

### Flag / Unflag Drafts

```bash
# Flag a draft
drafts flag <uuid>

# Flag active draft
drafts flag

# Unflag a draft
drafts unflag <uuid>

# Unflag active draft
drafts unflag
```

### Workspaces

```bash
# Show current workspace
drafts workspace

# List all workspaces
drafts workspace --list

# Open a workspace
drafts workspace --open Ideas
```

### Actions

```bash
# List all actions
drafts actions

# Filter actions by substring
drafts actions -s Copy
```

### Edit in Editor

```bash
drafts edit <uuid>
```

### Run Actions

```bash
# Run action on text
drafts run "Copy" "Text to copy to clipboard"

# Run action on existing draft
drafts run "Copy" -u <uuid>

# Raw JSON payload
drafts run --input '{"action":"Copy","uuid":"<uuid>"}'
```

### Get Schema

```bash
# Full schema for LLM integration
drafts schema

# Schema for specific command
drafts schema create
```

### Environment Info

```bash
# Basic info (version, app status, counts)
drafts info

# Verbose (includes actions, tags, workspaces)
drafts info --verbose

# Test permissions
drafts info --test-permissions
```

### Self-Upgrade

```bash
drafts upgrade
```

## Output Format

**JSON (default)** - All commands return structured JSON:
```json
{
  "success": true,
  "data": {
    "uuid": "ABC123",
    "content": "Note content",
    "title": "Note title",
    "tags": ["tag1", "tag2"],
    "folder": "inbox",
    "isFlagged": false,
    "isArchived": false,
    "isTrashed": false,
    "createdAt": "2026-01-29 10:00:00",
    "modifiedAt": "2026-01-29 10:30:00",
    "createdLatitude": 37.7749,
    "createdLongitude": -122.4194,
    "modifiedLatitude": 37.7749,
    "modifiedLongitude": -122.4194,
    "permalink": "drafts://open?uuid=ABC123"
  }
}
```

For `drafts list`, the default JSON omits `content` and location fields unless you pass `--full`.

**Plain text** - Human-readable output:
```bash
drafts list --plain
```

## Common Workflows

### Quick Capture
```bash
drafts create "Remember to call dentist tomorrow" -t reminder
```

### Daily Journal
```bash
drafts append "$(date): Completed project review" -u <journal-uuid>
```

### Search and Review
```bash
# Search draft content
drafts list -s "project alpha"

# List all drafts with a specific tag
drafts list -t work

# Get full content of a draft
drafts get <uuid>
```

### Flag Important Items
```bash
# Flag a draft for attention
drafts flag <uuid>

# List flagged drafts
drafts list -f flagged
```

### Workspace Filtering
```bash
# See current workspace
drafts workspace

# List drafts from a specific workspace
drafts list -w "Work Projects"
```

## Troubleshooting

**Commands fail or return empty:**
1. Is Drafts running? → `open -a Drafts`
2. Is Drafts Pro active? → Automation requires Pro
3. Permissions granted? → System Settings > Privacy > Automation

**Commands hang:**
- Check if Drafts is showing a dialog

## Notes

- macOS ONLY (AppleScript-based)
- Drafts app MUST be running
- Requires Drafts Pro subscription
- All UUIDs are Drafts-generated identifiers
- Tags are case-sensitive
- Drafts AppleScript does not expose syntax/language grammar, so this CLI does not surface a `syntax` command

## Version

v2.1.0
