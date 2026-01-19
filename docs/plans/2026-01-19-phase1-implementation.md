# Phase 1: LLM-First Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform the Drafts CLI from human-first to LLM-first by making JSON the default output, adding structured errors, exposing the `run` command, and adding a `schema` command.

**Architecture:** Add an output layer that wraps all command results in a consistent JSON envelope. Commands return structured data, the output layer serializes it. A global `--plain` flag bypasses JSON for human readability.

**Tech Stack:** Go, go-arg for CLI parsing, encoding/json for output

---

## Task 1: Add Output Types and JSON Envelope

**Files:**
- Create: `cmd/drafts/output.go`
- Modify: `pkg/drafts/struct.go`

**Step 1: Create the output types file**

Create `cmd/drafts/output.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Response envelope for all JSON output
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// Global flag for plain output
var plainOutput bool

// Output writes the response as JSON or plain text
func output(data interface{}) {
	if plainOutput {
		fmt.Println(data)
		return
	}
	resp := Response{Success: true, Data: data}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(resp)
}

// OutputError writes an error response
func outputError(code, message, hint string) {
	if plainOutput {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
		if hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
		}
		os.Exit(1)
	}
	resp := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Hint:    hint,
		},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(resp)
	os.Exit(1)
}
```

**Step 2: Run go fmt to verify syntax**

Run: `cd /Users/ashrafali/git-projects/drafts && go fmt ./cmd/drafts/output.go`
Expected: No errors, file formatted

**Step 3: Commit**

```bash
git add cmd/drafts/output.go
git commit -m "feat: add JSON output envelope types"
```

---

## Task 2: Add --plain Flag and Wire Up Output Layer

**Files:**
- Modify: `cmd/drafts/main.go:151-184`

**Step 1: Add --plain flag to args struct**

In `cmd/drafts/main.go`, modify the `main()` function args struct (around line 152):

```go
func main() {
	var args struct {
		Plain   bool        `arg:"--plain" help:"output plain text instead of JSON"`
		New     *NewCmd     `arg:"subcommand:new" help:"create new draft"`
		Prepend *PrependCmd `arg:"subcommand:prepend" help:"prepend to draft"`
		Append  *AppendCmd  `arg:"subcommand:append" help:"append to draft"`
		Replace *ReplaceCmd `arg:"subcommand:replace" help:"replace content of draft"`
		Edit    *EditCmd    `arg:"subcommand:edit" help:"edit draft in $EDITOR"`
		Get     *GetCmd     `arg:"subcommand:get" help:"get content of draft"`
		Select  *SelectCmd  `arg:"subcommand:select" help:"select active draft using fzf"`
		List    *ListCmd    `arg:"subcommand:list" help:"list drafts"`
	}
	p := arg.MustParse(&args)

	// Set global plain output flag
	plainOutput = args.Plain

	if p.Subcommand() == nil {
		p.Fail("missing subcommand")
	}
	switch {
	case args.New != nil:
		output(new(args.New))
	case args.Prepend != nil:
		output(prepend(args.Prepend))
	case args.Append != nil:
		output(append(args.Append))
	case args.Replace != nil:
		output(replace(args.Replace))
	case args.Edit != nil:
		output(edit(args.Edit))
	case args.Get != nil:
		output(get(args.Get))
	case args.Select != nil:
		output(_select())
	case args.List != nil:
		output(list(args.List))
	}
}
```

**Step 2: Build and verify it compiles**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 3: Test plain flag works**

Run: `cd /Users/ashrafali/git-projects/drafts && ./drafts --plain list 2>/dev/null || echo "OK - expected to fail without Drafts"`
Expected: Either list output or error (confirms flag is parsed)

**Step 4: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: add --plain flag, wire up JSON output"
```

---

## Task 3: Update Draft Struct with Full Metadata

**Files:**
- Modify: `pkg/drafts/struct.go`
- Modify: `pkg/drafts/js/get.js`

**Step 1: Expand Draft struct with additional fields**

Replace the Draft struct in `pkg/drafts/struct.go`:

```go
package drafts

const Separator = '|'

type Draft struct {
	UUID       string   `json:"uuid"`
	Content    string   `json:"content"`
	Title      string   `json:"title"`
	Tags       []string `json:"tags"`
	IsFlagged  bool     `json:"isFlagged"`
	IsArchived bool     `json:"isArchived"`
	IsTrashed  bool     `json:"isTrashed"`
	Folder     string   `json:"folder"`
	CreatedAt  string   `json:"createdAt"`
	ModifiedAt string   `json:"modifiedAt"`
	Permalink  string   `json:"permalink"`
}

// ---- Enums ------------------------------------------------------------------

type Folder int

const (
	FolderInbox Folder = iota
	FolderArchive
)

func (f Folder) String() string {
	return [...]string{"inbox", "archive"}[f]
}

type Filter int

const (
	FilterInbox Filter = iota
	FilterFlagged
	FilterArchive
	FilterTrash
	FilterAll
)

func (f Filter) String() string {
	return [...]string{"inbox", "flagged", "archive", "trash", "all"}[f]
}

type Sort int

const (
	SortCreated Sort = iota
	SortModified
	SortAccessed
)

func (s Sort) String() string {
	return [...]string{"created", "modified", "accessed"}[s]
}
```

**Step 2: Update get.js to return full metadata**

Replace `pkg/drafts/js/get.js`:

```javascript
let d = Draft.find(input[0]);
if (d) {
  let folder = d.isTrashed ? "trash" : (d.isArchived ? "archive" : "inbox");
  let res = {
    uuid: d.uuid,
    content: d.content,
    title: d.displayTitle,
    tags: d.tags,
    isFlagged: d.isFlagged,
    isArchived: d.isArchived,
    isTrashed: d.isTrashed,
    folder: folder,
    createdAt: d.createdAt.toISOString(),
    modifiedAt: d.modifiedAt.toISOString(),
    permalink: d.permalink
  };
  context.addSuccessParameter("result", JSON.stringify(res));
}
```

**Step 3: Verify build**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 4: Commit**

```bash
git add pkg/drafts/struct.go pkg/drafts/js/get.js
git commit -m "feat: expand Draft struct with full metadata"
```

---

## Task 4: Update Commands to Return Structured Data

**Files:**
- Modify: `cmd/drafts/main.go`

**Step 1: Update new() to return Draft object**

Change the `new` function to return a Draft instead of just UUID:

```go
func new(param *NewCmd) interface{} {
	// Input
	text := orStdin(param.Message)

	// Params -> Options
	opt := drafts.CreateOptions{
		Tags:    param.Tag,
		Flagged: param.Flagged,
	}

	if param.Archive {
		opt.Folder = drafts.FolderArchive
	}

	uuid := drafts.Create(text, opt)
	return drafts.Get(uuid)
}
```

**Step 2: Update get() to return Draft object**

```go
func get(param *GetCmd) interface{} {
	uuid := orActive(param.UUID)
	d := drafts.Get(uuid)
	if d.UUID == "" {
		outputError("DRAFT_NOT_FOUND",
			fmt.Sprintf("No draft found with UUID: %s", uuid),
			"Use 'drafts list' to see available drafts")
	}
	return d
}
```

**Step 3: Update prepend(), append(), replace() to return Draft**

```go
func prepend(param *PrependCmd) interface{} {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	drafts.Prepend(uuid, text)
	return drafts.Get(uuid)
}

func append(param *AppendCmd) interface{} {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	drafts.Append(uuid, text)
	return drafts.Get(uuid)
}

func replace(param *ReplaceCmd) interface{} {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	drafts.Replace(uuid, text)
	return drafts.Get(uuid)
}
```

**Step 4: Update edit() to return Draft**

```go
func edit(param *EditCmd) interface{} {
	uuid := orActive(param.UUID)
	new := editor(drafts.Get(uuid).Content)
	drafts.Replace(uuid, new)
	return drafts.Get(uuid)
}
```

**Step 5: Update list() to return []Draft**

```go
func list(param *ListCmd) interface{} {
	filter := parseFilter(param.Filter)
	ds := drafts.Query("", filter, drafts.QueryOptions{Tags: param.Tag})
	return map[string]interface{}{
		"drafts": ds,
		"count":  len(ds),
		"filter": param.Filter,
	}
}
```

**Step 6: Update _select() to return Draft**

```go
func _select() interface{} {
	ds := drafts.Query("", drafts.FilterInbox, drafts.QueryOptions{})
	var b strings.Builder
	linebreakRegex := regexp.MustCompile(`\n+`)
	for _, d := range ds {
		fmt.Fprintf(&b, "%s %c %s\n", d.UUID, drafts.Separator, linebreakRegex.ReplaceAllString(d.Content, linebreak))
	}
	uuid := fzfUUID(b.String())
	drafts.Select(uuid)
	return drafts.Get(uuid)
}
```

**Step 7: Verify build**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 8: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: update all commands to return structured data"
```

---

## Task 5: Add `run` Command

**Files:**
- Modify: `cmd/drafts/main.go`

**Step 1: Add RunCmd struct**

Add after the ListCmd struct (around line 115):

```go
type RunCmd struct {
	Action  string `arg:"positional,required" help:"action name to run"`
	Text    string `arg:"positional" help:"text to process (omit to use stdin)"`
	UUID    string `arg:"-u" help:"run action on existing draft by UUID"`
}
```

**Step 2: Add run() function**

Add after the list() function:

```go
func run(param *RunCmd) interface{} {
	var text string

	if param.UUID != "" {
		// Run action on existing draft
		d := drafts.Get(param.UUID)
		if d.UUID == "" {
			outputError("DRAFT_NOT_FOUND",
				fmt.Sprintf("No draft found with UUID: %s", param.UUID),
				"Use 'drafts list' to see available drafts")
		}
		text = d.Content
	} else {
		// Run action on provided text or stdin
		text = orStdin(param.Text)
	}

	result := drafts.RunAction(param.Action, text)

	return map[string]interface{}{
		"action": param.Action,
		"result": result.Encode(),
	}
}
```

**Step 3: Register run command in main()**

Add to the args struct:

```go
Run     *RunCmd     `arg:"subcommand:run" help:"run a Drafts action"`
```

Add to the switch statement:

```go
case args.Run != nil:
	output(run(args.Run))
```

**Step 4: Verify build**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 5: Test help shows run command**

Run: `./drafts --help`
Expected: Shows `run` in subcommands list

**Step 6: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: expose run command for action execution"
```

---

## Task 6: Add --action Flag to create/prepend/append

**Files:**
- Modify: `cmd/drafts/main.go`
- Modify: `pkg/drafts/drafts.go`
- Modify: `pkg/drafts/options.go`

**Step 1: Update CreateOptions with Action field**

In `pkg/drafts/options.go`:

```go
package drafts

type CreateOptions struct {
	Tags    []string
	Folder  Folder
	Flagged bool
	Action  string
}

type QueryOptions struct {
	Tags             []string
	OmitTags         []string
	Sort             Sort
	SortDescending   bool
	SortFlaggedToTop bool
}

type ModifyOptions struct {
	Tags   []string
	Action string
}
```

**Step 2: Update Create() to handle action parameter**

In `pkg/drafts/drafts.go`, update the Create function:

```go
func Create(text string, opt CreateOptions) string {
	v := url.Values{
		"text":    []string{text},
		"folder":  []string{opt.Folder.String()},
		"flagged": []string{mustJSON(opt.Flagged)},
	}
	if len(opt.Tags) > 0 {
		v["tag"] = opt.Tags
	}
	if opt.Action != "" {
		v["action"] = []string{opt.Action}
	}
	res := open("create", v)
	return res.Get("uuid")
}
```

**Step 3: Update Prepend/Append to accept options**

In `pkg/drafts/drafts.go`:

```go
func Prepend(uuid, text string, opt ModifyOptions) {
	v := url.Values{
		"uuid": []string{uuid},
		"text": []string{text},
	}
	if len(opt.Tags) > 0 {
		v["tag"] = opt.Tags
	}
	if opt.Action != "" {
		v["action"] = []string{opt.Action}
	}
	open("prepend", v)
}

func Append(uuid, text string, opt ModifyOptions) {
	v := url.Values{
		"uuid": []string{uuid},
		"text": []string{text},
	}
	if len(opt.Tags) > 0 {
		v["tag"] = opt.Tags
	}
	if opt.Action != "" {
		v["action"] = []string{opt.Action}
	}
	open("append", v)
}
```

**Step 4: Update CLI command structs**

In `cmd/drafts/main.go`:

```go
type NewCmd struct {
	Message string   `arg:"positional" help:"draft content (omit to use stdin)"`
	Tag     []string `arg:"-t,separate" help:"tag"`
	Archive bool     `arg:"-a" help:"create draft in archive"`
	Flagged bool     `arg:"-f" help:"create flagged draft"`
	Action  string   `arg:"--action" help:"action to run after creation"`
}

type PrependCmd struct {
	Message string   `arg:"positional" help:"text to prepend (omit to use stdin)"`
	UUID    string   `arg:"-u" help:"UUID (omit to use active draft)"`
	Tag     []string `arg:"-t,separate" help:"tag to add"`
	Action  string   `arg:"--action" help:"action to run after prepend"`
}

type AppendCmd struct {
	Message string   `arg:"positional" help:"text to append (omit to use stdin)"`
	UUID    string   `arg:"-u" help:"UUID (omit to use active draft)"`
	Tag     []string `arg:"-t,separate" help:"tag to add"`
	Action  string   `arg:"--action" help:"action to run after append"`
}
```

**Step 5: Update command implementations**

```go
func new(param *NewCmd) interface{} {
	text := orStdin(param.Message)
	opt := drafts.CreateOptions{
		Tags:    param.Tag,
		Flagged: param.Flagged,
		Action:  param.Action,
	}
	if param.Archive {
		opt.Folder = drafts.FolderArchive
	}
	uuid := drafts.Create(text, opt)
	return drafts.Get(uuid)
}

func prepend(param *PrependCmd) interface{} {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	opt := drafts.ModifyOptions{
		Tags:   param.Tag,
		Action: param.Action,
	}
	drafts.Prepend(uuid, text, opt)
	return drafts.Get(uuid)
}

func append(param *AppendCmd) interface{} {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	opt := drafts.ModifyOptions{
		Tags:   param.Tag,
		Action: param.Action,
	}
	drafts.Append(uuid, text, opt)
	return drafts.Get(uuid)
}
```

**Step 6: Verify build**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 7: Test help shows new flags**

Run: `./drafts new --help`
Expected: Shows `--action` and `-t` flags

**Step 8: Commit**

```bash
git add cmd/drafts/main.go pkg/drafts/drafts.go pkg/drafts/options.go
git commit -m "feat: add --action and --tag flags to create/prepend/append"
```

---

## Task 7: Add `schema` Command

**Files:**
- Create: `cmd/drafts/schema.go`
- Modify: `cmd/drafts/main.go`

**Step 1: Create schema.go with tool definitions**

Create `cmd/drafts/schema.go`:

```go
package main

// Schema returns the tool-use formatted schema for all commands
func getSchema(command string) interface{} {
	schema := map[string]interface{}{
		"name":    "drafts",
		"version": "0.2.0",
		"tools":   getTools(),
	}

	if command != "" {
		// Return schema for single command
		tools := getTools()
		for _, tool := range tools {
			t := tool.(map[string]interface{})
			if t["name"] == "drafts_"+command {
				return t
			}
		}
		outputError("UNKNOWN_COMMAND",
			"Unknown command: "+command,
			"Use 'drafts schema' to see all available commands")
	}

	return schema
}

func getTools() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"name":        "drafts_new",
			"description": "Create a new draft in Drafts.app",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The draft content",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to apply to the draft",
					},
					"folder": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"inbox", "archive"},
						"default":     "inbox",
						"description": "Folder to create draft in",
					},
					"flagged": map[string]interface{}{
						"type":        "boolean",
						"default":     false,
						"description": "Whether to flag the draft",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Action name to run after creation",
					},
				},
				"required": []string{},
			},
		},
		map[string]interface{}{
			"name":        "drafts_get",
			"description": "Get a draft by UUID, returns full draft metadata",
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
			"name":        "drafts_list",
			"description": "List drafts with optional filtering",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"filter": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"inbox", "flagged", "archive", "trash", "all"},
						"default":     "inbox",
						"description": "Filter drafts by folder",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
				},
				"required": []string{},
			},
		},
		map[string]interface{}{
			"name":        "drafts_append",
			"description": "Append text to an existing draft",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the draft (omit for active draft)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text to append",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to add",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Action to run after appending",
					},
				},
				"required": []string{"content"},
			},
		},
		map[string]interface{}{
			"name":        "drafts_prepend",
			"description": "Prepend text to an existing draft",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the draft (omit for active draft)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text to prepend",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to add",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Action to run after prepending",
					},
				},
				"required": []string{"content"},
			},
		},
		map[string]interface{}{
			"name":        "drafts_replace",
			"description": "Replace the content of an existing draft",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the draft (omit for active draft)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "New content for the draft",
					},
				},
				"required": []string{"content"},
			},
		},
		map[string]interface{}{
			"name":        "drafts_run",
			"description": "Run a Drafts action on text or an existing draft",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Name of the action to run",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text to process (ignored if uuid provided)",
					},
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "UUID of draft to run action on",
					},
				},
				"required": []string{"action"},
			},
		},
	}
}
```

**Step 2: Add SchemaCmd and wire up in main.go**

Add the struct:

```go
type SchemaCmd struct {
	Command string `arg:"positional" help:"command name (omit for full schema)"`
}
```

Add to args struct:

```go
Schema  *SchemaCmd  `arg:"subcommand:schema" help:"output tool-use schema for LLM integration"`
```

Add schema function:

```go
func schema(param *SchemaCmd) interface{} {
	return getSchema(param.Command)
}
```

Add to switch:

```go
case args.Schema != nil:
	output(schema(args.Schema))
```

**Step 3: Verify build**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 4: Test schema output**

Run: `./drafts schema`
Expected: JSON schema output with tools array

**Step 5: Commit**

```bash
git add cmd/drafts/schema.go cmd/drafts/main.go
git commit -m "feat: add schema command for LLM tool discovery"
```

---

## Task 8: Add Alias `create` for `new`

**Files:**
- Modify: `cmd/drafts/main.go`

**Step 1: Add CreateCmd as alias**

Add after NewCmd:

```go
type CreateCmd = NewCmd // Alias for 'new' using Drafts terminology
```

Add to args struct:

```go
Create  *CreateCmd  `arg:"subcommand:create" help:"create new draft (alias for 'new')"`
```

Add to switch:

```go
case args.Create != nil:
	output(new(args.Create))
```

**Step 2: Verify build and test**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts && ./drafts create --help`
Expected: Shows same help as `new`

**Step 3: Commit**

```bash
git add cmd/drafts/main.go
git commit -m "feat: add 'create' alias for 'new' command"
```

---

## Task 9: Final Integration Test and Polish

**Files:**
- Modify: `cmd/drafts/main.go` (minor cleanup)

**Step 1: Run full build**

Run: `cd /Users/ashrafali/git-projects/drafts && go build ./cmd/drafts`
Expected: No errors

**Step 2: Test JSON output**

Run: `./drafts schema | head -20`
Expected: Properly formatted JSON

**Step 3: Test plain output**

Run: `./drafts --plain schema | head -5`
Expected: Plain text output

**Step 4: Verify all commands in help**

Run: `./drafts --help`
Expected: Shows new, create, prepend, append, replace, edit, get, select, list, run, schema

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: Phase 1 complete - LLM-first foundation"
```

**Step 6: Push to fork**

```bash
git push origin main
```

---

## Summary

Phase 1 implements:

1. ✅ JSON output by default with `--plain` flag for humans
2. ✅ Structured error responses with codes and hints
3. ✅ Full draft metadata in responses (timestamps, title, permalink)
4. ✅ `run` command to execute Drafts actions
5. ✅ `--action` flag on new/prepend/append
6. ✅ `--tag` flag on prepend/append
7. ✅ `schema` command for LLM tool discovery
8. ✅ `create` alias for `new`

**Next Phase:** Add `info` command for environment discovery, then complete URL scheme coverage.
