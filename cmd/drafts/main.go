package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	arg "github.com/alexflint/go-arg"

	"github.com/nerveband/drafts-applescript-cli/pkg/drafts"
)

// Documentation URLs
const (
	repoURL   = "https://github.com/nerveband/drafts-applescript-cli"
	issuesURL = "https://github.com/nerveband/drafts-applescript-cli/issues"
	docsURL   = "https://docs.getdrafts.com"
)

const linebreak = " ¶ "

// ---- Commands ---------------------------------------------------------------

type NewCmd struct {
	Message string   `arg:"positional" help:"draft content (omit to use stdin)"`
	Tag     []string `arg:"-t,separate" help:"tag"`
	Archive bool     `arg:"-a" help:"create draft in archive"`
	Flagged bool     `arg:"-f" help:"create flagged draft"`
	Action  string   `arg:"--action" help:"action to run after creation"`
	Input   string   `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

type CreateCmd = NewCmd // Alias for 'new' using Drafts terminology

func runNew(param *NewCmd) interface{} {
	request, err := resolveCreateRequest(param)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or use flags/positionals, not both")
	}

	opt := drafts.CreateOptions{
		Tags:    request.Tags,
		Flagged: request.Flagged,
		Action:  request.Action,
	}
	if request.Folder == "archive" {
		opt.Folder = drafts.FolderArchive
	}

	uuid, err := drafts.Create(request.Content, opt)
	handleDraftsError(err)

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type PrependCmd struct {
	Message string   `arg:"positional" help:"text to prepend (omit to use stdin)"`
	UUID    string   `arg:"-u" help:"UUID (omit to use active draft)"`
	Tag     []string `arg:"-t,separate" help:"tag to add"`
	Action  string   `arg:"--action" help:"action to run after prepend"`
	Input   string   `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

func runPrepend(param *PrependCmd) interface{} {
	request, err := resolveModifyRequest(param.Input, param.Message, param.Tag, param.Action, param.UUID)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or use flags/positionals, not both")
	}

	uuid, err := resolveCommandUUID(request.UUID)
	handleDraftsError(err)

	err = drafts.Prepend(uuid, request.Content, drafts.ModifyOptions{
		Tags:   request.Tags,
		Action: request.Action,
	})
	handleDraftsError(err)

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type AppendCmd struct {
	Message string   `arg:"positional" help:"text to append (omit to use stdin)"`
	UUID    string   `arg:"-u" help:"UUID (omit to use active draft)"`
	Tag     []string `arg:"-t,separate" help:"tag to add"`
	Action  string   `arg:"--action" help:"action to run after append"`
	Input   string   `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

func runAppend(param *AppendCmd) interface{} {
	request, err := resolveModifyRequest(param.Input, param.Message, param.Tag, param.Action, param.UUID)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or use flags/positionals, not both")
	}

	uuid, err := resolveCommandUUID(request.UUID)
	handleDraftsError(err)

	err = drafts.Append(uuid, request.Content, drafts.ModifyOptions{
		Tags:   request.Tags,
		Action: request.Action,
	})
	handleDraftsError(err)

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type ReplaceCmd struct {
	Message string `arg:"positional" help:"text to replace draft content with (omit to use stdin)"`
	UUID    string `arg:"-u" help:"UUID (omit to use active draft)"`
	Input   string `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

func runReplace(param *ReplaceCmd) interface{} {
	request, err := resolveReplaceRequest(param)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or use flags/positionals, not both")
	}

	uuid, err := resolveCommandUUID(request.UUID)
	handleDraftsError(err)

	err = drafts.Replace(uuid, request.Content)
	handleDraftsError(err)

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type EditCmd struct {
	UUID string `arg:"positional" help:"UUID (omit to use active draft)"`
}

func runEdit(param *EditCmd) interface{} {
	uuid, err := resolveCommandUUID(param.UUID)
	handleDraftsError(err)

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)

	updatedContent, err := editor(draft.Content)
	if err != nil {
		outputError("PERMISSION_DENIED", err.Error(), "Set $EDITOR to a working editor and ensure it can launch from this shell")
	}

	err = drafts.Replace(uuid, updatedContent)
	handleDraftsError(err)

	updatedDraft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(updatedDraft, true)
}

type GetCmd struct {
	UUID string `arg:"positional" help:"UUID (omit to use active draft)"`
}

func runGet(param *GetCmd) interface{} {
	uuid, err := resolveCommandUUID(param.UUID)
	handleDraftsError(err)

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type SelectCmd struct{}

func runSelect() interface{} {
	ds, err := drafts.Query("", drafts.FilterInbox, drafts.QueryOptions{})
	handleDraftsError(err)

	var b strings.Builder
	linebreakRegex := regexp.MustCompile(`\n+`)
	for _, d := range ds {
		fmt.Fprintf(&b, "%s %c %s\n", d.UUID, drafts.Separator, linebreakRegex.ReplaceAllString(d.Content, linebreak))
	}

	uuid, err := fzfUUID(b.String())
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Selection is interactive; use 'drafts get <uuid>' for non-interactive access")
	}

	handleDraftsError(drafts.Select(uuid))

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type FlagCmd struct {
	UUID  string `arg:"positional" help:"UUID (omit to use active draft)"`
	Input string `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

type UnflagCmd struct {
	UUID  string `arg:"positional" help:"UUID (omit to use active draft)"`
	Input string `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

func runFlag(param *FlagCmd) interface{} {
	request, err := resolveUUIDRequest(param.Input, param.UUID)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or use a positional UUID")
	}

	uuid, err := resolveCommandUUID(request.UUID)
	handleDraftsError(err)

	handleDraftsError(drafts.SetFlagged(uuid, true))

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

func runUnflag(param *UnflagCmd) interface{} {
	request, err := resolveUUIDRequest(param.Input, param.UUID)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or use a positional UUID")
	}

	uuid, err := resolveCommandUUID(request.UUID)
	handleDraftsError(err)

	handleDraftsError(drafts.SetFlagged(uuid, false))

	draft, err := drafts.Get(uuid)
	handleDraftsError(err)
	return toDraftView(draft, true)
}

type WorkspaceCmd struct {
	List bool   `arg:"-l,--list" help:"list all workspaces"`
	Open string `arg:"-o,--open" help:"open a workspace by name"`
}

func runWorkspace(param *WorkspaceCmd) interface{} {
	if param.List && param.Open != "" {
		outputError("INVALID_INPUT", "cannot combine --list and --open", "Use either 'drafts workspace --list' or 'drafts workspace --open <name>'")
	}

	if param.List {
		workspaces, err := drafts.Workspaces()
		handleDraftsError(err)
		return WorkspaceResult{
			Workspaces: workspaces,
			Count:      len(workspaces),
		}
	}

	if param.Open != "" {
		handleDraftsError(drafts.OpenWorkspace(param.Open))
		current, err := drafts.CurrentWorkspace()
		handleDraftsError(err)
		return WorkspaceResult{
			Opened:  param.Open,
			Current: current,
		}
	}

	current, err := drafts.CurrentWorkspace()
	handleDraftsError(err)
	return WorkspaceResult{
		Current: current,
	}
}

type ActionsCmd struct {
	Search string `arg:"-s,--search" help:"filter action names by substring"`
}

func runActions(param *ActionsCmd) interface{} {
	filtered := filterActions(getAvailableActions(), param.Search)
	return ActionsResult{
		Actions: filtered,
		Count:   len(filtered),
		Search:  param.Search,
	}
}

type ListCmd struct {
	Filter    string   `arg:"-f" default:"inbox" help:"filter: inbox|flagged|archive|trash|all"`
	Tag       []string `arg:"-t,separate" help:"filter by tag"`
	Search    string   `arg:"-s" help:"search draft content"`
	Workspace string   `arg:"-w" help:"filter by workspace name"`
	Limit     int      `arg:"--limit" default:"20" help:"maximum drafts to return (0 for all)"`
	Full      bool     `arg:"--full" help:"include full content and location fields in list results"`
}

func parseFilter(s string) (drafts.Filter, error) {
	switch s {
	case "inbox":
		return drafts.FilterInbox, nil
	case "flagged":
		return drafts.FilterFlagged, nil
	case "archive":
		return drafts.FilterArchive, nil
	case "trash":
		return drafts.FilterTrash, nil
	case "all":
		return drafts.FilterAll, nil
	default:
		return drafts.FilterInbox, fmt.Errorf("invalid filter: %s", s)
	}
}

func runList(param *ListCmd) interface{} {
	if param.Limit < 0 {
		outputError("INVALID_INPUT", "limit must be 0 or greater", "Use --limit 0 for all results, or a positive number to cap output")
	}

	filter, err := parseFilter(param.Filter)
	if err != nil {
		outputError("INVALID_FILTER", err.Error(), "Valid filters: inbox, flagged, archive, trash, all")
	}

	var ds []drafts.Draft
	if param.Workspace != "" {
		ds, err = drafts.QueryWorkspace(param.Workspace, param.Search, filter, drafts.QueryOptions{Tags: param.Tag})
	} else {
		ds, err = drafts.Query(param.Search, filter, drafts.QueryOptions{Tags: param.Tag})
	}
	handleDraftsError(err)

	if param.Limit > 0 && len(ds) > param.Limit {
		ds = ds[:param.Limit]
	}

	views := make([]DraftView, len(ds))
	for i, draft := range ds {
		views[i] = toDraftView(draft, param.Full)
	}

	return ListResult{
		Drafts:    views,
		Count:     len(views),
		Filter:    param.Filter,
		Limit:     param.Limit,
		Full:      param.Full,
		Search:    param.Search,
		Workspace: param.Workspace,
	}
}

type RunCmd struct {
	Action string `arg:"positional" help:"action name to run"`
	Text   string `arg:"positional" help:"text to process (omit to use stdin)"`
	UUID   string `arg:"-u" help:"run action on existing draft by UUID"`
	Input  string `arg:"--input" help:"raw JSON payload; use '-' to read from stdin"`
}

type SchemaCmd struct {
	Command string `arg:"positional" help:"command name or alias (omit for full schema)"`
}

type InfoCmd struct {
	Verbose         bool `arg:"-v,--verbose" help:"show full lists of actions, tags, workspaces"`
	TestPermissions bool `arg:"--test-permissions" help:"test what operations work with current setup"`
}

type UpgradeCmd struct{}

type VersionCmd struct{}

func runAction(param *RunCmd) interface{} {
	request, err := resolveRunRequest(param)
	if err != nil {
		outputError("INVALID_INPUT", err.Error(), "Use --input with one JSON object or provide a positional action plus uuid/content")
	}

	if request.Action == "" {
		outputError("INVALID_INPUT", "run requires an action name", "Pass an action positionally or via --input")
	}
	if request.UUID != "" && request.Content != "" {
		outputError("INVALID_INPUT", "run accepts either uuid or content, not both", "Provide uuid to target an existing draft, or content to create a transient draft")
	}

	if request.UUID != "" {
		uuid, err := resolveCommandUUID(request.UUID)
		handleDraftsError(err)

		handleDraftsError(drafts.RunActionOnDraft(request.Action, uuid))

		draft, err := drafts.Get(uuid)
		handleDraftsError(err)
		view := toDraftView(draft, true)
		return RunResult{
			Action:       request.Action,
			UUID:         uuid,
			CreatedDraft: false,
			Draft:        &view,
		}
	}

	result, err := drafts.RunAction(request.Action, request.Content)
	handleDraftsError(err)

	draft, err := drafts.Get(result.UUID)
	handleDraftsError(err)
	view := toDraftView(draft, true)
	return RunResult{
		Action:       request.Action,
		UUID:         result.UUID,
		CreatedDraft: result.CreatedDraft,
		Draft:        &view,
	}
}

func runSchema(param *SchemaCmd) interface{} {
	return getSchema(param.Command)
}

// ---- Main -------------------------------------------------------------------

// Args holds command-line arguments
type Args struct {
	Plain     bool          `arg:"--plain" help:"output plain text instead of JSON"`
	New       *NewCmd       `arg:"subcommand:new" help:"create new draft"`
	Create    *CreateCmd    `arg:"subcommand:create" help:"create new draft (alias for 'new')"`
	Prepend   *PrependCmd   `arg:"subcommand:prepend" help:"prepend to draft"`
	Append    *AppendCmd    `arg:"subcommand:append" help:"append to draft"`
	Replace   *ReplaceCmd   `arg:"subcommand:replace" help:"replace content of draft"`
	Edit      *EditCmd      `arg:"subcommand:edit" help:"edit draft in $EDITOR"`
	Get       *GetCmd       `arg:"subcommand:get" help:"get content of draft"`
	Select    *SelectCmd    `arg:"subcommand:select" help:"select active draft using fzf"`
	List      *ListCmd      `arg:"subcommand:list" help:"list drafts"`
	Flag      *FlagCmd      `arg:"subcommand:flag" help:"flag a draft"`
	Unflag    *UnflagCmd    `arg:"subcommand:unflag" help:"unflag a draft"`
	Workspace *WorkspaceCmd `arg:"subcommand:workspace" help:"show, list, or open workspaces"`
	Actions   *ActionsCmd   `arg:"subcommand:actions" help:"list available actions"`
	Run       *RunCmd       `arg:"subcommand:run" help:"run a Drafts action"`
	Info      *InfoCmd      `arg:"subcommand:info" help:"show environment info and diagnostics"`
	Schema    *SchemaCmd    `arg:"subcommand:schema" help:"output tool-use schema for LLM integration"`
	Upgrade   *UpgradeCmd   `arg:"subcommand:upgrade" help:"upgrade to the latest version"`
	Version   *VersionCmd   `arg:"subcommand:version" help:"show version information"`
}

// Description returns the program description for help
func (Args) Description() string {
	return "Drafts CLI - Interact with Drafts.app from the command line\n\nRequires: macOS, Drafts.app running, Drafts Pro subscription"
}

// Epilogue returns the footer for help output
func (Args) Epilogue() string {
	return fmt.Sprintf(`Documentation:
  Repository:     %s
  Report issues:  %s
  Drafts docs:    %s

For agents:
  Run 'drafts schema' for the machine-readable contract.
  Use '--input' on mutating commands for raw JSON payloads.
  Run 'drafts info' to verify Drafts automation access.`, repoURL, issuesURL, docsURL)
}

func main() {
	var args Args
	p := arg.MustParse(&args)

	// Set global plain output flag
	plainOutput = args.Plain

	if p.Subcommand() == nil {
		printBanner()
		p.WriteHelp(os.Stdout)
		fmt.Println()
		fmt.Println(args.Epilogue())
		fmt.Println()
		checkAndNotifyUpdate()
		return
	}

	if args.Info == nil && args.Schema == nil && args.Upgrade == nil && args.Version == nil && !isDraftsRunning() {
		outputError("DRAFTS_NOT_RUNNING",
			"Drafts.app must be running for this command",
			"Open Drafts and try again, or run 'drafts info' for diagnostics")
	}

	switch {
	case args.New != nil:
		output(runNew(args.New))
	case args.Create != nil:
		output(runNew(args.Create))
	case args.Prepend != nil:
		output(runPrepend(args.Prepend))
	case args.Append != nil:
		output(runAppend(args.Append))
	case args.Replace != nil:
		output(runReplace(args.Replace))
	case args.Edit != nil:
		output(runEdit(args.Edit))
	case args.Get != nil:
		output(runGet(args.Get))
	case args.Select != nil:
		output(runSelect())
	case args.List != nil:
		output(runList(args.List))
	case args.Flag != nil:
		output(runFlag(args.Flag))
	case args.Unflag != nil:
		output(runUnflag(args.Unflag))
	case args.Workspace != nil:
		output(runWorkspace(args.Workspace))
	case args.Actions != nil:
		output(runActions(args.Actions))
	case args.Run != nil:
		output(runAction(args.Run))
	case args.Info != nil:
		output(runInfo(args.Info))
		checkAndNotifyUpdate()
	case args.Schema != nil:
		output(runSchema(args.Schema))
	case args.Upgrade != nil:
		output(runUpgrade())
	case args.Version != nil:
		output(runVersion())
		checkAndNotifyUpdate()
	}
}

func resolveCommandUUID(uuid string) (string, error) {
	if uuid != "" {
		return uuid, nil
	}

	active, err := drafts.Active()
	if err != nil || active == "" {
		return "", fmt.Errorf("%w: active draft", drafts.ErrDraftNotFound)
	}
	return active, nil
}

func handleDraftsError(err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, drafts.ErrDraftNotFound):
		outputError("DRAFT_NOT_FOUND", err.Error(), "Use 'drafts list' to see available drafts")
	case errors.Is(err, drafts.ErrActionNotFound):
		outputError("ACTION_NOT_FOUND", err.Error(), "Use 'drafts actions' to inspect available actions")
	case errors.Is(err, drafts.ErrWorkspaceNotFound):
		outputError("WORKSPACE_NOT_FOUND", err.Error(), "Use 'drafts workspace --list' to inspect available workspaces")
	default:
		outputError("PERMISSION_DENIED", err.Error(), "Run 'drafts info --test-permissions' to diagnose automation access")
	}
}
