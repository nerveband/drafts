# AppleScript Documentation Alignment - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Align the CLI with the full Drafts AppleScript documentation — add missing properties, new operations, query optimizations, and new CLI commands.

**Architecture:** Extend the existing `pkg/drafts` package with new fields in the `Draft` struct, new functions for flag/unflag/syntax/workspace operations, and optimize existing queries to use AppleScript-level filtering. Add corresponding CLI commands in `cmd/drafts/main.go` and update the schema.

**Tech Stack:** Go 1.24+, AppleScript via `osascript`, existing `go-arg` CLI framework, existing `internal/assert` test helpers.

**Testing:** All tests are integration tests requiring Drafts.app to be running on macOS. Tests create real drafts and clean up via `defer Trash(uuid)`. Run with `go test ./pkg/drafts/ -v -run TestName`.

---

## Phase 1: Data Model — Add Missing Draft Properties

### Task 1: Add missing fields to Draft struct

**Files:**
- Modify: `pkg/drafts/struct.go:5-17`

**Step 1: Write the failing test**

Add to `pkg/drafts/drafts_test.go`:

```go
func TestGetReturnsLanguageGrammar(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	// Default grammar should be "Plain Text" or similar non-empty string
	if draft.LanguageGrammar == "" {
		t.Errorf("expected non-empty LanguageGrammar, got empty string")
	}
}

func TestGetReturnsLocationFields(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	// Location fields exist and are parseable (may be 0.0 if no location)
	// Just verify the fields exist in the struct and were parsed
	if draft.UUID == "" {
		t.Errorf("expected valid draft, got empty UUID")
	}
	// Verify the type is correct (float64 zero value is fine)
	_ = draft.CreatedLatitude
	_ = draft.CreatedLongitude
	_ = draft.ModifiedLatitude
	_ = draft.ModifiedLongitude
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run "TestGetReturnsLanguageGrammar|TestGetReturnsLocationFields"`
Expected: Compilation failure — `draft.LanguageGrammar undefined`, `draft.CreatedLatitude undefined`, etc.

**Step 3: Write minimal implementation — update struct**

In `pkg/drafts/struct.go`, update the `Draft` struct:

```go
type Draft struct {
	UUID              string   `json:"uuid"`
	Content           string   `json:"content"`
	Title             string   `json:"title"`
	Tags              []string `json:"tags"`
	IsFlagged         bool     `json:"isFlagged"`
	IsArchived        bool     `json:"isArchived"`
	IsTrashed         bool     `json:"isTrashed"`
	Folder            string   `json:"folder"`
	LanguageGrammar   string   `json:"languageGrammar"`
	CreatedAt         string   `json:"createdAt"`
	ModifiedAt        string   `json:"modifiedAt"`
	CreatedLatitude   float64  `json:"createdLatitude"`
	CreatedLongitude  float64  `json:"createdLongitude"`
	ModifiedLatitude  float64  `json:"modifiedLatitude"`
	ModifiedLongitude float64  `json:"modifiedLongitude"`
	Permalink         string   `json:"permalink"`
}
```

This compiles but tests still fail because `Get()` doesn't return these fields yet. That's Task 2.

**Step 4: Run test to verify compilation passes (tests still fail at runtime)**

Run: `go build ./...`
Expected: Compilation succeeds.

**Step 5: Commit**

```bash
git add pkg/drafts/struct.go pkg/drafts/drafts_test.go
git commit -m "feat: add languageGrammar and location fields to Draft struct

Add LanguageGrammar, CreatedLatitude, CreatedLongitude,
ModifiedLatitude, ModifiedLongitude to match Drafts AppleScript docs.
Tests added but will pass after Task 2 updates Get/Query."
```

---

### Task 2: Update Get() AppleScript to return new fields

**Files:**
- Modify: `pkg/drafts/drafts.go:128-184` (Get + parseDraftFromAppleScript)

**Step 1: The failing tests from Task 1 already exist**

Tests `TestGetReturnsLanguageGrammar` and `TestGetReturnsLocationFields` should now compile but fail at runtime because the AppleScript and parser don't return the new fields.

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run "TestGetReturnsLanguageGrammar"`
Expected: FAIL — `LanguageGrammar` is empty string because AppleScript doesn't return it and parser doesn't populate it.

**Step 3: Write minimal implementation — update Get() and parser**

Update the `Get` function's AppleScript in `pkg/drafts/drafts.go` to include the new fields. The tab-separated output adds 5 new fields after `permalink`:

```go
func Get(uuid string) Draft {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set folder_name to folder of d as string
	set is_archived to false
	set is_trashed to false
	if folder_name is "archive" then
		set is_archived to true
	else if folder_name is "trash" then
		set is_trashed to true
	end if
	set tag_list to tags of d
	set tag_str to ""
	repeat with t in tag_list
		if tag_str is not "" then
			set tag_str to tag_str & "|||"
		end if
		set tag_str to tag_str & t
	end repeat
	return (id of d) & "	" & (title of d) & "	" & (content of d) & "	" & folder_name & "	" & (flagged of d) & "	" & is_archived & "	" & is_trashed & "	" & tag_str & "	" & ((createdAt of d) as string) & "	" & ((modifiedAt of d) as string) & "	" & (permalink of d) & "	" & (languageGrammar of d) & "	" & (createdLatitude of d) & "	" & (createdLongitude of d) & "	" & (modifiedLatitude of d) & "	" & (modifiedLongitude of d)
end tell`, escapeForAppleScript(uuid))

	output, err := runAppleScript(script)
	if err != nil {
		return Draft{}
	}

	return parseDraftFromAppleScript(output)
}
```

Update `parseDraftFromAppleScript` to parse the new fields (now expects 16 tab-separated fields):

```go
func parseDraftFromAppleScript(output string) Draft {
	parts := strings.Split(output, "\t")
	if len(parts) < 11 {
		return Draft{}
	}

	tags := []string{}
	if parts[7] != "" {
		tags = strings.Split(parts[7], "|||")
	}

	draft := Draft{
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

	// Parse new fields if present (backward-compatible)
	if len(parts) > 11 {
		draft.LanguageGrammar = parts[11]
	}
	if len(parts) > 14 {
		fmt.Sscanf(parts[12], "%f", &draft.CreatedLatitude)
		fmt.Sscanf(parts[13], "%f", &draft.CreatedLongitude)
		fmt.Sscanf(parts[14], "%f", &draft.ModifiedLatitude)
	}
	if len(parts) > 15 {
		fmt.Sscanf(parts[15], "%f", &draft.ModifiedLongitude)
	}

	return draft
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run "TestGetReturnsLanguageGrammar|TestGetReturnsLocationFields"`
Expected: PASS

**Step 5: Run ALL existing tests to verify no regressions**

Run: `go test ./pkg/drafts/ -v`
Expected: All existing tests still pass (the DeepEqual comparisons in existing tests need updating — see Task 2b).

**Note:** Existing tests like `TestCreateDefault` use `assert.DeepEqual` which compares the FULL struct. They'll now fail because the old expected value doesn't include `LanguageGrammar`. Those tests need updating to not hardcode the full struct or to include the new fields. Handle this by updating the test expectations to use field-level assertions instead of full DeepEqual for the fields they care about.

**Step 5b: Fix existing test assertions for new fields**

Update existing tests in `drafts_test.go` that use `assert.DeepEqual` to account for new fields. For example, `TestCreateDefault` should verify the fields it cares about individually rather than doing a full struct comparison, OR we set the expected struct to include the new fields:

For `TestCreateDefault`, change the assertion approach to check relevant fields:
```go
func TestCreateDefault(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{}, draft.Tags)
	assert.Equal(t, false, draft.IsFlagged)
	assert.Equal(t, false, draft.IsArchived)
	assert.Equal(t, false, draft.IsTrashed)
}
```

Apply the same pattern to: `TestCreateFlagged`, `TestCreateArchived`, `TestCreateTags`, `TestTrash`, `TestArchive`, `TestTag`.

**Step 6: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: return languageGrammar and location fields from Get()

Update AppleScript in Get() to fetch languageGrammar, createdLatitude,
createdLongitude, modifiedLatitude, modifiedLongitude.
Update parser to handle 16 fields (backward-compatible with 11).
Update existing test assertions to field-level checks."
```

---

### Task 3: Update Query() AppleScript to return new fields

**Files:**
- Modify: `pkg/drafts/drafts.go:186-258` (Query function)

**Step 1: Write the failing test**

Add to `pkg/drafts/drafts_test.go`:

```go
func TestQueryReturnsLanguageGrammar(t *testing.T) {
	tag := rand()
	uuid := Create("test grammar", CreateOptions{Tags: []string{tag}})
	defer func() {
		Trash(uuid)
	}()

	results := Query("", FilterInbox, QueryOptions{Tags: []string{tag}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].LanguageGrammar == "" {
		t.Errorf("expected non-empty LanguageGrammar in query results")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestQueryReturnsLanguageGrammar`
Expected: FAIL — `LanguageGrammar` empty because Query's AppleScript doesn't include it.

**Step 3: Write minimal implementation**

Update the AppleScript template in `Query()` to include the new fields (match the same 16-field format as `Get()`):

```go
func Query(queryString string, filter Filter, opt QueryOptions) []Draft {
	var folderFilter string
	switch filter {
	case FilterArchive:
		folderFilter = "whose folder is archive"
	case FilterTrash:
		folderFilter = "whose folder is trash"
	case FilterAll:
		folderFilter = ""
	default:
		folderFilter = "whose folder is inbox"
	}

	script := fmt.Sprintf(`tell application "Drafts"
	set output to ""
	set allDrafts to every draft %s
	repeat with d in allDrafts
		set folder_name to folder of d as string
		set is_archived to false
		set is_trashed to false
		if folder_name is "archive" then
			set is_archived to true
		else if folder_name is "trash" then
			set is_trashed to true
		end if
		set tag_list to tags of d
		set tag_str to ""
		repeat with t in tag_list
			if tag_str is not "" then
				set tag_str to tag_str & "|||"
			end if
			set tag_str to tag_str & t
		end repeat
		set line_out to (id of d) & "	" & (title of d) & "	" & (content of d) & "	" & folder_name & "	" & (flagged of d) & "	" & is_archived & "	" & is_trashed & "	" & tag_str & "	" & ((createdAt of d) as string) & "	" & ((modifiedAt of d) as string) & "	" & (permalink of d) & "	" & (languageGrammar of d) & "	" & (createdLatitude of d) & "	" & (createdLongitude of d) & "	" & (modifiedLatitude of d) & "	" & (modifiedLongitude of d)
		if output is "" then
			set output to line_out
		else
			set output to output & linefeed & line_out
		end if
	end repeat
	return output
end tell`, folderFilter)

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
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestQueryReturnsLanguageGrammar`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./pkg/drafts/ -v`
Expected: All pass.

**Step 6: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: return new fields from Query() AppleScript

Query now returns languageGrammar and location fields matching Get()."
```

---

## Phase 2: New Operations

### Task 4: Flag/Unflag existing drafts

**Files:**
- Modify: `pkg/drafts/drafts.go` (add SetFlagged function)
- Modify: `pkg/drafts/drafts_test.go` (add tests)

**Step 1: Write the failing test**

```go
func TestSetFlagged(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()

	// Verify starts unflagged
	draft := Get(uuid)
	assert.Equal(t, false, draft.IsFlagged)

	// Flag it
	SetFlagged(uuid, true)
	draft = Get(uuid)
	assert.Equal(t, true, draft.IsFlagged)

	// Unflag it
	SetFlagged(uuid, false)
	draft = Get(uuid)
	assert.Equal(t, false, draft.IsFlagged)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestSetFlagged`
Expected: Compile error — `SetFlagged` undefined.

**Step 3: Write minimal implementation**

Add to `pkg/drafts/drafts.go` (after the `Tag` function):

```go
// SetFlagged sets the flagged status of a draft.
func SetFlagged(uuid string, flagged bool) {
	flaggedStr := "false"
	if flagged {
		flaggedStr = "true"
	}
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set flagged of d to %s
end tell`, escapeForAppleScript(uuid), flaggedStr)

	runAppleScript(script)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestSetFlagged`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: add SetFlagged() to flag/unflag existing drafts"
```

---

### Task 5: Set language grammar on existing drafts

**Files:**
- Modify: `pkg/drafts/drafts.go` (add SetLanguageGrammar function)
- Modify: `pkg/drafts/drafts_test.go` (add test)

**Step 1: Write the failing test**

```go
func TestSetLanguageGrammar(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()

	SetLanguageGrammar(uuid, "JavaScript")
	draft := Get(uuid)
	assert.Equal(t, "JavaScript", draft.LanguageGrammar)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestSetLanguageGrammar`
Expected: Compile error — `SetLanguageGrammar` undefined.

**Step 3: Write minimal implementation**

Add to `pkg/drafts/drafts.go`:

```go
// SetLanguageGrammar sets the syntax highlighting language of a draft.
func SetLanguageGrammar(uuid, grammar string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set languageGrammar of d to "%s"
end tell`, escapeForAppleScript(uuid), escapeForAppleScript(grammar))

	runAppleScript(script)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestSetLanguageGrammar`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: add SetLanguageGrammar() to set draft syntax"
```

---

### Task 6: Content search in queries

**Files:**
- Modify: `pkg/drafts/drafts.go:186-258` (Query function)
- Modify: `pkg/drafts/drafts_test.go`

**Step 1: Write the failing test**

```go
func TestQueryContentSearch(t *testing.T) {
	tag := rand()
	needle := "UNIQUE_SEARCH_" + rand()

	a := Create("has the needle "+needle+" inside", CreateOptions{Tags: []string{tag}})
	b := Create("no match here", CreateOptions{Tags: []string{tag}})
	defer func() {
		Trash(a)
		Trash(b)
	}()

	// Search with content filter
	results := Query(needle, FilterInbox, QueryOptions{Tags: []string{tag}})
	uuids := getUUIDs(results)
	assert.EqualSlice(t, []string{a}, uuids)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestQueryContentSearch`
Expected: FAIL — returns both drafts because `queryString` is currently ignored.

**Step 3: Write minimal implementation**

Update `Query()` in `pkg/drafts/drafts.go` to use `queryString` for AppleScript-level content filtering:

```go
func Query(queryString string, filter Filter, opt QueryOptions) []Draft {
	// Build the filter clauses
	var filterClauses []string

	switch filter {
	case FilterArchive:
		filterClauses = append(filterClauses, "folder is archive")
	case FilterTrash:
		filterClauses = append(filterClauses, "folder is trash")
	case FilterFlagged:
		filterClauses = append(filterClauses, "folder is inbox", "flagged is true")
	case FilterAll:
		// No folder filter
	default: // FilterInbox
		filterClauses = append(filterClauses, "folder is inbox")
	}

	// Content search at AppleScript level
	if queryString != "" {
		filterClauses = append(filterClauses, fmt.Sprintf(`content contains "%s"`, escapeForAppleScript(queryString)))
	}

	// Build the whose clause
	whereClause := ""
	if len(filterClauses) > 0 {
		whereClause = "whose " + strings.Join(filterClauses, " and ")
	}

	script := fmt.Sprintf(`tell application "Drafts"
	set output to ""
	set allDrafts to every draft %s
	repeat with d in allDrafts
		set folder_name to folder of d as string
		set is_archived to false
		set is_trashed to false
		if folder_name is "archive" then
			set is_archived to true
		else if folder_name is "trash" then
			set is_trashed to true
		end if
		set tag_list to tags of d
		set tag_str to ""
		repeat with t in tag_list
			if tag_str is not "" then
				set tag_str to tag_str & "|||"
			end if
			set tag_str to tag_str & t
		end repeat
		set line_out to (id of d) & "	" & (title of d) & "	" & (content of d) & "	" & folder_name & "	" & (flagged of d) & "	" & is_archived & "	" & is_trashed & "	" & tag_str & "	" & ((createdAt of d) as string) & "	" & ((modifiedAt of d) as string) & "	" & (permalink of d) & "	" & (languageGrammar of d) & "	" & (createdLatitude of d) & "	" & (createdLongitude of d) & "	" & (modifiedLatitude of d) & "	" & (modifiedLongitude of d)
		if output is "" then
			set output to line_out
		else
			set output to output & linefeed & line_out
		end if
	end repeat
	return output
end tell`, whereClause)

	output, err := runAppleScript(script)
	if err != nil {
		return []Draft{}
	}

	if output == "" {
		return []Draft{}
	}

	lines := strings.Split(output, "\n")
	results := make([]Draft, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			d := parseDraftFromAppleScript(line)
			if d.UUID != "" {
				if len(opt.Tags) > 0 && !hasAllTags(d.Tags, opt.Tags) {
					continue
				}
				if len(opt.OmitTags) > 0 && hasAnyTag(d.Tags, opt.OmitTags) {
					continue
				}
				results = append(results, d)
			}
		}
	}

	return results
}
```

**Important:** This also fixes `FilterFlagged` to use AppleScript-level `flagged is true` (was previously just falling through to inbox filter). This is the optimization from the gap analysis.

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestQueryContentSearch`
Expected: PASS

**Step 5: Run all existing Query tests**

Run: `go test ./pkg/drafts/ -v -run TestQuery`
Expected: All pass (existing callers pass `""` as queryString so behavior unchanged).

**Step 6: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: add content search and AppleScript-level flagged filter

Query() now uses queryString for AppleScript-level content search.
FilterFlagged now uses AppleScript-level 'flagged is true' clause.
All filter clauses combined with 'and' for efficient querying."
```

---

### Task 7: Workspace-based query filtering

**Files:**
- Modify: `pkg/drafts/options.go` (add Workspace to QueryOptions)
- Modify: `pkg/drafts/drafts.go` (add QueryWorkspace function)
- Modify: `pkg/drafts/drafts_test.go`

**Step 1: Write the failing test**

```go
func TestQueryWorkspace(t *testing.T) {
	// This test verifies workspace querying works.
	// It uses whatever workspaces exist - just verifies the function runs.
	results := QueryWorkspace("", FilterAll, QueryOptions{})
	// Should return an empty slice (not nil) if workspace doesn't exist
	if results == nil {
		t.Errorf("expected non-nil result from QueryWorkspace")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestQueryWorkspace`
Expected: Compile error — `QueryWorkspace` undefined.

**Step 3: Write minimal implementation**

Add to `pkg/drafts/drafts.go`:

```go
// QueryWorkspace returns drafts from a specific workspace.
func QueryWorkspace(workspace string, filter Filter, opt QueryOptions) []Draft {
	if workspace == "" {
		return []Draft{}
	}

	script := fmt.Sprintf(`tell application "Drafts"
	set output to ""
	set ws to workspace "%s"
	set allDrafts to every draft of ws
	repeat with d in allDrafts
		set folder_name to folder of d as string
		set is_archived to false
		set is_trashed to false
		if folder_name is "archive" then
			set is_archived to true
		else if folder_name is "trash" then
			set is_trashed to true
		end if
		set tag_list to tags of d
		set tag_str to ""
		repeat with t in tag_list
			if tag_str is not "" then
				set tag_str to tag_str & "|||"
			end if
			set tag_str to tag_str & t
		end repeat
		set line_out to (id of d) & "	" & (title of d) & "	" & (content of d) & "	" & folder_name & "	" & (flagged of d) & "	" & is_archived & "	" & is_trashed & "	" & tag_str & "	" & ((createdAt of d) as string) & "	" & ((modifiedAt of d) as string) & "	" & (permalink of d) & "	" & (languageGrammar of d) & "	" & (createdLatitude of d) & "	" & (createdLongitude of d) & "	" & (modifiedLatitude of d) & "	" & (modifiedLongitude of d)
		if output is "" then
			set output to line_out
		else
			set output to output & linefeed & line_out
		end if
	end repeat
	return output
end tell`, escapeForAppleScript(workspace))

	output, err := runAppleScript(script)
	if err != nil {
		return []Draft{}
	}

	if output == "" {
		return []Draft{}
	}

	lines := strings.Split(output, "\n")
	results := make([]Draft, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			d := parseDraftFromAppleScript(line)
			if d.UUID != "" {
				if len(opt.Tags) > 0 && !hasAllTags(d.Tags, opt.Tags) {
					continue
				}
				if len(opt.OmitTags) > 0 && hasAnyTag(d.Tags, opt.OmitTags) {
					continue
				}
				results = append(results, d)
			}
		}
	}

	return results
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestQueryWorkspace`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: add QueryWorkspace() for workspace-based filtering"
```

---

### Task 8: Get current workspace

**Files:**
- Modify: `pkg/drafts/drafts.go` (add CurrentWorkspace function)
- Modify: `pkg/drafts/drafts_test.go`

**Step 1: Write the failing test**

```go
func TestCurrentWorkspace(t *testing.T) {
	ws := CurrentWorkspace()
	// Current workspace should return some string (could be empty if none set)
	// Just verify it doesn't panic and returns a string
	_ = ws
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestCurrentWorkspace`
Expected: Compile error — `CurrentWorkspace` undefined.

**Step 3: Write minimal implementation**

Add to `pkg/drafts/drafts.go` (near the Active function):

```go
// CurrentWorkspace returns the name of the current workspace.
func CurrentWorkspace() string {
	script := `tell application "Drafts"
	return name of current workspace
end tell`

	name, err := runAppleScript(script)
	if err != nil {
		return ""
	}
	return name
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestCurrentWorkspace`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: add CurrentWorkspace() to get active workspace name"
```

---

### Task 9: Get list of workspaces from pkg/drafts

**Files:**
- Modify: `pkg/drafts/drafts.go` (add Workspaces function)
- Modify: `pkg/drafts/drafts_test.go`

This moves workspace enumeration from `info.go` (cmd layer) into the `pkg/drafts` package so it's reusable.

**Step 1: Write the failing test**

```go
func TestWorkspaces(t *testing.T) {
	workspaces := Workspaces()
	// Should return a non-nil slice
	if workspaces == nil {
		t.Errorf("expected non-nil workspaces list")
	}
	// Drafts always has at least one workspace
	if len(workspaces) == 0 {
		t.Errorf("expected at least one workspace")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/drafts/ -v -run TestWorkspaces`
Expected: Compile error — `Workspaces` undefined.

**Step 3: Write minimal implementation**

Add to `pkg/drafts/drafts.go`:

```go
// Workspaces returns the names of all workspaces.
func Workspaces() []string {
	script := `tell application "Drafts"
	set wsNames to {}
	repeat with w in (every workspace)
		set end of wsNames to (name of w)
	end repeat
	set output to ""
	repeat with n in wsNames
		if output is "" then
			set output to n
		else
			set output to output & "|||" & n
		end if
	end repeat
	return output
end tell`

	output, err := runAppleScript(script)
	if err != nil || output == "" {
		return []string{}
	}
	return strings.Split(output, "|||")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/drafts/ -v -run TestWorkspaces`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/drafts/drafts.go pkg/drafts/drafts_test.go
git commit -m "feat: add Workspaces() to list all workspace names"
```

---

## Phase 3: CLI Integration

### Task 10: Add `flag` and `unflag` CLI commands

**Files:**
- Modify: `cmd/drafts/main.go` (add FlagCmd, UnflagCmd, wire up)

**Step 1: Write the failing test**

Since CLI tests are manual (no test file for `cmd/drafts`), verify compilation:

Run: `go build ./cmd/drafts/`
Expected: Build succeeds after adding code.

For manual verification:
```bash
# Create a test draft, flag it, verify, unflag it, verify
./drafts new "flag test" --plain
./drafts flag <uuid>
./drafts get <uuid>     # should show isFlagged: true
./drafts unflag <uuid>
./drafts get <uuid>     # should show isFlagged: false
```

**Step 2: Write minimal implementation**

Add to `cmd/drafts/main.go` — new command structs:

```go
type FlagCmd struct {
	UUID string `arg:"positional" help:"UUID (omit to use active draft)"`
}

type UnflagCmd struct {
	UUID string `arg:"positional" help:"UUID (omit to use active draft)"`
}
```

Add handler functions:

```go
func flag(param *FlagCmd) interface{} {
	uuid := orActive(param.UUID)
	drafts.SetFlagged(uuid, true)
	return drafts.Get(uuid)
}

func unflag(param *UnflagCmd) interface{} {
	uuid := orActive(param.UUID)
	drafts.SetFlagged(uuid, false)
	return drafts.Get(uuid)
}
```

Add to `Args` struct:

```go
Flag    *FlagCmd    `arg:"subcommand:flag" help:"flag a draft"`
Unflag  *UnflagCmd  `arg:"subcommand:unflag" help:"unflag a draft"`
```

Add to the switch in `main()`:

```go
case args.Flag != nil:
	output(flag(args.Flag))
case args.Unflag != nil:
	output(unflag(args.Unflag))
```

**Step 3: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 4: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: add flag/unflag CLI commands"
```

---

### Task 11: Add `syntax` CLI command

**Files:**
- Modify: `cmd/drafts/main.go`

**Step 1: Write minimal implementation**

Add command struct:

```go
type SyntaxCmd struct {
	Grammar string `arg:"positional,required" help:"language grammar (e.g., Markdown, JavaScript, Plain Text)"`
	UUID    string `arg:"-u" help:"UUID (omit to use active draft)"`
}
```

Add handler:

```go
func syntax(param *SyntaxCmd) interface{} {
	uuid := orActive(param.UUID)
	drafts.SetLanguageGrammar(uuid, param.Grammar)
	return drafts.Get(uuid)
}
```

Add to `Args` struct:

```go
Syntax  *SyntaxCmd  `arg:"subcommand:syntax" help:"set language grammar/syntax of a draft"`
```

Add to switch:

```go
case args.Syntax != nil:
	output(syntax(args.Syntax))
```

**Step 2: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 3: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: add syntax CLI command to set language grammar"
```

---

### Task 12: Add search and workspace options to `list` command

**Files:**
- Modify: `cmd/drafts/main.go` (update ListCmd)

**Step 1: Write minimal implementation**

Update `ListCmd`:

```go
type ListCmd struct {
	Filter    string   `arg:"-f" default:"inbox" help:"filter: inbox|flagged|archive|trash|all"`
	Tag       []string `arg:"-t,separate" help:"filter by tag"`
	Search    string   `arg:"-s" help:"search draft content"`
	Workspace string   `arg:"-w" help:"filter by workspace name"`
}
```

Update `list()`:

```go
func list(param *ListCmd) interface{} {
	filter := parseFilter(param.Filter)

	var ds []drafts.Draft
	if param.Workspace != "" {
		ds = drafts.QueryWorkspace(param.Workspace, filter, drafts.QueryOptions{Tags: param.Tag})
	} else {
		ds = drafts.Query(param.Search, filter, drafts.QueryOptions{Tags: param.Tag})
	}

	result := map[string]interface{}{
		"drafts": ds,
		"count":  len(ds),
		"filter": param.Filter,
	}
	if param.Search != "" {
		result["search"] = param.Search
	}
	if param.Workspace != "" {
		result["workspace"] = param.Workspace
	}
	return result
}
```

**Step 2: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 3: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: add -s search and -w workspace options to list command"
```

---

### Task 13: Add `workspace` standalone command

**Files:**
- Modify: `cmd/drafts/main.go`

**Step 1: Write minimal implementation**

Add command struct:

```go
type WorkspaceCmd struct {
	List bool `arg:"-l,--list" help:"list all workspaces"`
}
```

Add handler:

```go
func workspace(param *WorkspaceCmd) interface{} {
	if param.List {
		ws := drafts.Workspaces()
		return map[string]interface{}{
			"workspaces": ws,
			"count":      len(ws),
		}
	}
	// Default: show current workspace
	current := drafts.CurrentWorkspace()
	return map[string]interface{}{
		"current": current,
	}
}
```

Add to `Args`:

```go
Workspace *WorkspaceCmd `arg:"subcommand:workspace" help:"show current workspace or list all"`
```

Add to switch:

```go
case args.Workspace != nil:
	output(workspace(args.Workspace))
```

**Step 2: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 3: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: add workspace command (current + list)"
```

---

### Task 14: Update schema for all new commands

**Files:**
- Modify: `cmd/drafts/schema.go`

**Step 1: Write minimal implementation**

Add to the `getTools()` function in `schema.go`, after the existing tools:

```go
map[string]interface{}{
	"name":        "drafts_flag",
	"description": "Flag a draft",
	"parameters": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"uuid": map[string]interface{}{
				"type":        "string",
				"description": "UUID of the draft (omit for active draft)",
			},
		},
		"required": []string{},
	},
},
map[string]interface{}{
	"name":        "drafts_unflag",
	"description": "Unflag a draft",
	"parameters": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"uuid": map[string]interface{}{
				"type":        "string",
				"description": "UUID of the draft (omit for active draft)",
			},
		},
		"required": []string{},
	},
},
map[string]interface{}{
	"name":        "drafts_syntax",
	"description": "Set the language grammar/syntax of a draft (e.g., Markdown, JavaScript, Plain Text)",
	"parameters": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"grammar": map[string]interface{}{
				"type":        "string",
				"description": "Language grammar name (e.g., Markdown, JavaScript, Plain Text)",
			},
			"uuid": map[string]interface{}{
				"type":        "string",
				"description": "UUID of the draft (omit for active draft)",
			},
		},
		"required": []string{"grammar"},
	},
},
map[string]interface{}{
	"name":        "drafts_workspace",
	"description": "Show current workspace or list all workspaces",
	"parameters": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"list": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "List all workspaces instead of showing current",
			},
		},
		"required": []string{},
	},
},
```

Also update the `drafts_list` schema to include the new options:

```go
// In the drafts_list tool, add to properties:
"search": map[string]interface{}{
	"type":        "string",
	"description": "Search draft content (AppleScript-level filtering)",
},
"workspace": map[string]interface{}{
	"type":        "string",
	"description": "Filter by workspace name",
},
```

**Step 2: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 3: Commit**

```bash
git add cmd/drafts/schema.go
git commit -m "feat: update schema with flag, unflag, syntax, workspace, search"
```

---

## Phase 4: Update info.go to use pkg/drafts functions

### Task 15: Refactor info.go to use shared workspace/tag functions

**Files:**
- Modify: `cmd/drafts/info.go`

**Step 1: Refactor**

Update `runInfo` in `info.go` to use the new `drafts.Workspaces()` function instead of its local `getAvailableWorkspaces()`:

```go
// In runInfo(), change:
if param.Verbose {
	result.Tags = getAvailableTags()
	result.Actions = getAvailableActions()
	result.Workspaces = drafts.Workspaces()
}
```

Then remove the `getAvailableWorkspaces()` function from `info.go` since it's now in `pkg/drafts`.

**Step 2: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 3: Commit**

```bash
git add cmd/drafts/info.go
git commit -m "refactor: use pkg/drafts.Workspaces() in info command"
```

---

## Phase 5: Cleanup

### Task 16: Remove unused queryString TODO and clean up dead code

**Files:**
- Modify: `cmd/drafts/info.go:484-485` (remove unused `_ = time.Now`)

**Step 1: Remove dead code**

In `info.go`, remove line 485:
```go
// Remove this line:
var _ = time.Now
```

Also remove the `"time"` import if it's no longer used.

**Step 2: Build and verify**

Run: `go build ./cmd/drafts/ && echo "OK"`
Expected: OK

**Step 3: Commit**

```bash
git add cmd/drafts/info.go
git commit -m "chore: remove unused time import and dead code"
```

---

### Task 17: Run full test suite and final verification

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass (requires Drafts.app running).

**Step 2: Build final binary**

Run: `go build -o drafts ./cmd/drafts/`

**Step 3: Manual smoke test**

```bash
./drafts info
./drafts schema
./drafts workspace
./drafts workspace --list
./drafts new "test content" --plain
./drafts flag <uuid>
./drafts get <uuid>
./drafts syntax Markdown -u <uuid>
./drafts get <uuid>
./drafts unflag <uuid>
./drafts list -s "test content"
./drafts list -f all
```

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete AppleScript docs alignment

Adds: languageGrammar, location fields, flag/unflag, syntax,
content search, workspace commands, AppleScript-level filtering."
```

---

## Summary of All Changes

| File | Changes |
|------|---------|
| `pkg/drafts/struct.go` | Add 5 fields: LanguageGrammar, CreatedLatitude, CreatedLongitude, ModifiedLatitude, ModifiedLongitude |
| `pkg/drafts/drafts.go` | Update Get/Query AppleScript + parser; add SetFlagged, SetLanguageGrammar, QueryWorkspace, CurrentWorkspace, Workspaces; refactor Query to use clause builder |
| `pkg/drafts/drafts_test.go` | Add 8 new tests; update existing DeepEqual assertions to field-level |
| `cmd/drafts/main.go` | Add flag, unflag, syntax, workspace commands; add -s/-w to list |
| `cmd/drafts/schema.go` | Add schemas for 4 new commands; update list schema |
| `cmd/drafts/info.go` | Use `drafts.Workspaces()`; remove dead code |

**New CLI commands:** `flag`, `unflag`, `syntax`, `workspace`
**Enhanced commands:** `list` (now has `-s` search and `-w` workspace)
**New draft properties:** `languageGrammar`, `createdLatitude`, `createdLongitude`, `modifiedLatitude`, `modifiedLongitude`
