# Phase 2: AppleScript Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the URL scheme + callback handler backend with pure AppleScript for zero-dependency operation.

**Architecture:** Replace `helpers.go` with AppleScript execution via `osascript`. Rewrite all functions in `drafts.go` to construct and execute AppleScript commands. Remove all JavaScript files and the callback handler dependency.

**Tech Stack:** Go, os/exec for osascript, AppleScript for Drafts communication

---

## Task 1: Create AppleScript Helper Functions

**Files:**
- Create: `pkg/drafts/applescript.go`
- Delete: `pkg/drafts/helpers.go` (after Task 6)

**Step 1: Create applescript.go with core helpers**

Create `pkg/drafts/applescript.go`:

```go
package drafts

import (
	"fmt"
	"os/exec"
	"strings"
)

// runAppleScript executes an AppleScript and returns the output
func runAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("applescript error: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// escapeForAppleScript escapes a string for use in AppleScript
func escapeForAppleScript(s string) string {
	// Escape backslashes first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// tagsToAppleScript converts a slice of tags to AppleScript list format
func tagsToAppleScript(tags []string) string {
	if len(tags) == 0 {
		return "{}"
	}
	escaped := make([]string, len(tags))
	for i, t := range tags {
		escaped[i] = fmt.Sprintf("\"%s\"", escapeForAppleScript(t))
	}
	return "{" + strings.Join(escaped, ", ") + "}"
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./pkg/drafts`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/drafts/applescript.go
git commit -m "feat: add AppleScript helper functions"
```

---

## Task 2: Implement Create via AppleScript

**Files:**
- Modify: `pkg/drafts/drafts.go`

**Step 1: Rewrite Create function**

Replace the Create function in `pkg/drafts/drafts.go`:

```go
// Create a new draft. Return new draft's UUID.
func Create(text string, opt CreateOptions) string {
	folder := "inbox"
	if opt.Folder == FolderArchive {
		folder = "archive"
	}

	flaggedStr := "false"
	if opt.Flagged {
		flaggedStr = "true"
	}

	script := fmt.Sprintf(`tell application "Drafts"
	set d to make new draft with properties {content:"%s", flagged:%s, tags:%s}
	set folder of d to %s
	return id of d
end tell`, escapeForAppleScript(text), flaggedStr, tagsToAppleScript(opt.Tags), folder)

	uuid, err := runAppleScript(script)
	if err != nil {
		return ""
	}

	// Run action if specified
	if opt.Action != "" {
		RunActionOnDraft(opt.Action, uuid)
	}

	return uuid
}
```

**Step 2: Test create works**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts && ./drafts create "Test from new AppleScript backend" --plain`
Expected: Returns UUID

**Step 3: Commit**

```bash
git add pkg/drafts/drafts.go
git commit -m "feat: implement Create via AppleScript"
```

---

## Task 3: Implement Get via AppleScript

**Files:**
- Modify: `pkg/drafts/drafts.go`

**Step 1: Rewrite Get function**

Replace the Get function:

```go
// Get content of draft.
func Get(uuid string) Draft {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set folder_name to "inbox"
	if isTrashed of d then
		set folder_name to "trash"
	else if isArchived of d then
		set folder_name to "archive"
	end if
	set tag_list to tags of d
	set tag_str to ""
	repeat with t in tag_list
		if tag_str is not "" then
			set tag_str to tag_str & "|||"
		end if
		set tag_str to tag_str & t
	end repeat
	return (id of d) & "	" & (title of d) & "	" & (content of d) & "	" & folder_name & "	" & (flagged of d) & "	" & (isArchived of d) & "	" & (isTrashed of d) & "	" & tag_str & "	" & ((createdAt of d) as string) & "	" & ((modifiedAt of d) as string) & "	" & (permalink of d)
end tell`, escapeForAppleScript(uuid))

	output, err := runAppleScript(script)
	if err != nil {
		return Draft{}
	}

	return parseDraftFromAppleScript(output)
}

// parseDraftFromAppleScript parses tab-separated AppleScript output into a Draft
func parseDraftFromAppleScript(output string) Draft {
	parts := strings.Split(output, "\t")
	if len(parts) < 11 {
		return Draft{}
	}

	tags := []string{}
	if parts[7] != "" {
		tags = strings.Split(parts[7], "|||")
	}

	return Draft{
		UUID:       parts[0],
		Title:      parts[1],
		Content:    parts[2],
		Folder:     parts[3],
		IsFlagged:  parts[4] == "true",
		IsArchived: parts[5] == "true",
		IsTrashed:  parts[6] == "true",
		Tags:       tags,
		CreatedAt:  parts[8],
		ModifiedAt: parts[9],
		Permalink:  parts[10],
	}
}
```

**Step 2: Test get works**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts && ./drafts get --plain`
Expected: Returns draft content (or error if no active draft)

**Step 3: Commit**

```bash
git add pkg/drafts/drafts.go
git commit -m "feat: implement Get via AppleScript"
```

---

## Task 4: Implement Query (List) via AppleScript

**Files:**
- Modify: `pkg/drafts/drafts.go`

**Step 1: Rewrite Query function**

Replace the Query function:

```go
// Query for drafts.
func Query(queryString string, filter Filter, opt QueryOptions) []Draft {
	filterStr := filter.String()

	script := fmt.Sprintf(`tell application "Drafts"
	set output to ""
	set allDrafts to (every draft whose folder is %s)
	repeat with d in allDrafts
		set folder_name to "inbox"
		if isTrashed of d then
			set folder_name to "trash"
		else if isArchived of d then
			set folder_name to "archive"
		end if
		set tag_list to tags of d
		set tag_str to ""
		repeat with t in tag_list
			if tag_str is not "" then
				set tag_str to tag_str & "|||"
			end if
			set tag_str to tag_str & t
		end repeat
		set line_out to (id of d) & "	" & (title of d) & "	" & (content of d) & "	" & folder_name & "	" & (flagged of d) & "	" & (isArchived of d) & "	" & (isTrashed of d) & "	" & tag_str & "	" & ((createdAt of d) as string) & "	" & ((modifiedAt of d) as string) & "	" & (permalink of d)
		if output is "" then
			set output to line_out
		else
			set output to output & linefeed & line_out
		end if
	end repeat
	return output
end tell`, filterStr)

	output, err := runAppleScript(script)
	if err != nil {
		return []Draft{}
	}

	if output == "" {
		return []Draft{}
	}

	lines := strings.Split(output, "\n")
	drafts := make([]Draft, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			d := parseDraftFromAppleScript(line)
			if d.UUID != "" {
				// Apply tag filters if specified
				if len(opt.Tags) > 0 && !hasAllTags(d.Tags, opt.Tags) {
					continue
				}
				if len(opt.OmitTags) > 0 && hasAnyTag(d.Tags, opt.OmitTags) {
					continue
				}
				drafts = append(drafts, d)
			}
		}
	}

	return drafts
}

func hasAllTags(draftTags, requiredTags []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range draftTags {
		tagSet[t] = true
	}
	for _, t := range requiredTags {
		if !tagSet[t] {
			return false
		}
	}
	return true
}

func hasAnyTag(draftTags, excludeTags []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range draftTags {
		tagSet[t] = true
	}
	for _, t := range excludeTags {
		if tagSet[t] {
			return true
		}
	}
	return false
}
```

**Step 2: Test list works**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts && ./drafts list --plain`
Expected: Lists drafts from inbox

**Step 3: Commit**

```bash
git add pkg/drafts/drafts.go
git commit -m "feat: implement Query via AppleScript"
```

---

## Task 5: Implement Modify Operations via AppleScript

**Files:**
- Modify: `pkg/drafts/drafts.go`

**Step 1: Rewrite Prepend, Append, Replace functions**

```go
// Prepend to an existing draft.
func Prepend(uuid, text string, opt ModifyOptions) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set content of d to "%s" & linefeed & (content of d)
end tell`, escapeForAppleScript(uuid), escapeForAppleScript(text))

	runAppleScript(script)

	// Add tags if specified
	if len(opt.Tags) > 0 {
		Tag(uuid, opt.Tags...)
	}

	// Run action if specified
	if opt.Action != "" {
		RunActionOnDraft(opt.Action, uuid)
	}
}

// Append to an existing draft.
func Append(uuid, text string, opt ModifyOptions) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set content of d to (content of d) & linefeed & "%s"
end tell`, escapeForAppleScript(uuid), escapeForAppleScript(text))

	runAppleScript(script)

	// Add tags if specified
	if len(opt.Tags) > 0 {
		Tag(uuid, opt.Tags...)
	}

	// Run action if specified
	if opt.Action != "" {
		RunActionOnDraft(opt.Action, uuid)
	}
}

// Replace content of an existing draft.
func Replace(uuid, text string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set content of d to "%s"
end tell`, escapeForAppleScript(uuid), escapeForAppleScript(text))

	runAppleScript(script)
}
```

**Step 2: Test append works**

Run: `./drafts create "Initial content" --plain` then `./drafts append "Appended text" -u <UUID> --plain`
Expected: Content is appended

**Step 3: Commit**

```bash
git add pkg/drafts/drafts.go
git commit -m "feat: implement Prepend/Append/Replace via AppleScript"
```

---

## Task 6: Implement Draft Status and Tag Operations

**Files:**
- Modify: `pkg/drafts/drafts.go`

**Step 1: Rewrite Trash, Archive, Tag, Select, Active functions**

```go
// Trash a draft.
func Trash(uuid string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set folder of d to trash
end tell`, escapeForAppleScript(uuid))

	runAppleScript(script)
}

// Archive a draft.
func Archive(uuid string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set folder of d to archive
end tell`, escapeForAppleScript(uuid))

	runAppleScript(script)
}

// Tag adds tags to a draft.
func Tag(uuid string, tags ...string) {
	if len(tags) == 0 {
		return
	}

	// Get existing tags and merge
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set existingTags to tags of d
	set newTags to %s
	repeat with t in newTags
		if t is not in existingTags then
			set end of existingTags to t
		end if
	end repeat
	set tags of d to existingTags
end tell`, escapeForAppleScript(uuid), tagsToAppleScript(tags))

	runAppleScript(script)
}

// Select sets the active draft.
func Select(uuid string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	open d
end tell`, escapeForAppleScript(uuid))

	runAppleScript(script)
}

// Active returns the UUID of the active draft.
func Active() string {
	script := `tell application "Drafts"
	return id of current draft
end tell`

	uuid, err := runAppleScript(script)
	if err != nil {
		return ""
	}
	return uuid
}
```

**Step 2: Test Active works**

Run: `./drafts get --plain`
Expected: Returns current draft (uses Active internally)

**Step 3: Commit**

```bash
git add pkg/drafts/drafts.go
git commit -m "feat: implement Trash/Archive/Tag/Select/Active via AppleScript"
```

---

## Task 7: Implement RunAction via AppleScript

**Files:**
- Modify: `pkg/drafts/drafts.go`

**Step 1: Rewrite RunAction and add RunActionOnDraft**

```go
// RunAction runs an action with text (creates temp draft, runs action, returns result).
// Note: For running actions on existing drafts, use RunActionOnDraft.
func RunAction(action, text string) url.Values {
	// Create temp draft, run action, get result
	// This is a simplified version - actions may have various outputs
	script := fmt.Sprintf(`tell application "Drafts"
	set d to make new draft with properties {content:"%s"}
	set actionToRun to missing value
	repeat with a in (every action)
		if name of a is "%s" then
			set actionToRun to a
			exit repeat
		end if
	end repeat
	if actionToRun is not missing value then
		perform action actionToRun on draft d
	end if
	return id of d
end tell`, escapeForAppleScript(text), escapeForAppleScript(action))

	uuid, _ := runAppleScript(script)

	result := url.Values{}
	result.Set("uuid", uuid)
	return result
}

// RunActionOnDraft runs an action on an existing draft.
func RunActionOnDraft(action, uuid string) error {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set actionToRun to missing value
	repeat with a in (every action)
		if name of a is "%s" then
			set actionToRun to a
			exit repeat
		end if
	end repeat
	if actionToRun is not missing value then
		perform action actionToRun on draft d
		return "success"
	else
		return "action not found"
	end if
end tell`, escapeForAppleScript(uuid), escapeForAppleScript(action))

	result, err := runAppleScript(script)
	if err != nil {
		return err
	}
	if result == "action not found" {
		return fmt.Errorf("action not found: %s", action)
	}
	return nil
}
```

**Step 2: Test run command works**

Run: `./drafts run "Copy" "Test text" --plain`
Expected: Runs the Copy action

**Step 3: Commit**

```bash
git add pkg/drafts/drafts.go
git commit -m "feat: implement RunAction via AppleScript"
```

---

## Task 8: Remove Old URL Scheme Code and JS Files

**Files:**
- Delete: `pkg/drafts/helpers.go`
- Delete: `pkg/drafts/js.go`
- Delete: `pkg/drafts/js/` directory

**Step 1: Remove old files**

```bash
cd /Users/ashrafali/git-projects/drafts
rm pkg/drafts/helpers.go
rm pkg/drafts/js.go
rm -rf pkg/drafts/js/
```

**Step 2: Update imports in drafts.go**

Remove unused imports from `pkg/drafts/drafts.go`. The new file should only need:

```go
package drafts

import (
	"fmt"
	"net/url"
	"strings"
)
```

**Step 3: Verify build**

Run: `go build ./cmd/drafts`
Expected: No errors

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove URL scheme and JS code, AppleScript backend complete"
```

---

## Task 9: Update README and Test Full CLI

**Files:**
- Modify: `README.md`

**Step 1: Update README with new install instructions**

Replace the install section in README.md:

```markdown
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

No additional dependencies required! The CLI communicates directly with Drafts via AppleScript.

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

## LLM Integration

This CLI is designed for LLM tool use. Get the schema:

```bash
drafts schema
```

Output is JSON by default. Use `--plain` for human-readable output.
```

**Step 2: Test all commands**

Run these tests:
```bash
# Create
./drafts create "Test draft" -t test --plain

# List
./drafts list --plain

# Get (use UUID from create)
./drafts get <UUID> --plain

# Append
./drafts append "More content" -u <UUID> --plain

# Schema
./drafts schema | head -20

# Run action
./drafts run "Copy" "Clipboard test" --plain
```

**Step 3: Commit and push**

```bash
git add README.md
git commit -m "docs: update README for AppleScript backend"
git push origin main
```

---

## Summary

Phase 2 implements:

1. ✅ AppleScript helper functions
2. ✅ Create via AppleScript
3. ✅ Get via AppleScript
4. ✅ Query/List via AppleScript
5. ✅ Prepend/Append/Replace via AppleScript
6. ✅ Trash/Archive/Tag/Select/Active via AppleScript
7. ✅ RunAction via AppleScript
8. ✅ Remove old URL scheme + JS code
9. ✅ Updated README

**Result:** Zero-dependency CLI that works out of the box on any Mac with Drafts.
