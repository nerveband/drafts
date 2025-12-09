package main

import (
	"fmt"
	"regexp"
	"strings"

	arg "github.com/alexflint/go-arg"

	"github.com/ernstwi/drafts/pkg/drafts"
)

const linebreak = " ¶ "

// ---- Commands ---------------------------------------------------------------

type NewCmd struct {
	Message string   `arg:"positional" help:"draft content (omit to use stdin)"`
	Tag     []string `arg:"-t,separate" help:"tag"`
	Archive bool     `arg:"-a" help:"create draft in archive"`
	Flagged bool     `arg:"-f" help:"create flagged draft"`
}

func new(param *NewCmd) string {
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
	return uuid
}

type PrependCmd struct {
	Message string `arg:"positional" help:"text to prepend (omit to use stdin)"`
	UUID    string `arg:"-u" help:"UUID (omit to use active draft)"`
}

func prepend(param *PrependCmd) string {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	drafts.Prepend(uuid, text)
	return drafts.Get(uuid).Content
}

type AppendCmd struct {
	Message string `arg:"positional" help:"text to append (omit to use stdin)"`
	UUID    string `arg:"-u" help:"UUID (omit to use active draft)"`
}

func append(param *AppendCmd) string {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	drafts.Append(uuid, text)
	return drafts.Get(uuid).Content
}

type ReplaceCmd struct {
	Message string `arg:"positional" help:"text to append (omit to use stdin)"`
	UUID    string `arg:"-u" help:"UUID (omit to use active draft)"`
}

func replace(param *ReplaceCmd) string {
	text := orStdin(param.Message)
	uuid := orActive(param.UUID)
	drafts.Replace(uuid, text)
	return drafts.Get(uuid).Content
}

type EditCmd struct {
	UUID string `arg:"positional" help:"UUID (omit to use active draft)"`
}

func edit(param *EditCmd) string {
	uuid := orActive(param.UUID)
	new := editor(drafts.Get(uuid).Content)
	drafts.Replace(uuid, new)
	return new
}

type GetCmd struct {
	UUID string `arg:"positional" help:"UUID (omit to use active draft)"`
}

func get(param *GetCmd) string {
	uuid := orActive(param.UUID)
	return drafts.Get(uuid).Content
}

type SelectCmd struct{}

func _select() string {
	ds := drafts.Query("", drafts.FilterInbox, drafts.QueryOptions{})
	var b strings.Builder
	linebreakRegex := regexp.MustCompile(`\n+`)
	for _, d := range ds {
		fmt.Fprintf(&b, "%s %c %s\n", d.UUID, drafts.Separator, linebreakRegex.ReplaceAllString(d.Content, linebreak))
	}
	uuid := fzfUUID(b.String())
	drafts.Select(uuid)
	return drafts.Get(uuid).Content
}

type ListCmd struct {
	Filter string   `arg:"-f" default:"inbox" help:"filter: inbox|flagged|archive|trash|all"`
	Tag    []string `arg:"-t,separate" help:"filter by tag"`
}

func parseFilter(s string) drafts.Filter {
	switch s {
	case "inbox":
		return drafts.FilterInbox
	case "flagged":
		return drafts.FilterFlagged
	case "archive":
		return drafts.FilterArchive
	case "trash":
		return drafts.FilterTrash
	case "all":
		return drafts.FilterAll
	default:
		return drafts.FilterInbox
	}
}

func list(param *ListCmd) string {
	filter := parseFilter(param.Filter)
	ds := drafts.Query("", filter, drafts.QueryOptions{Tags: param.Tag})
	var b strings.Builder
	linebreakRegex := regexp.MustCompile(`\n+`)
	for _, d := range ds {
		firstLine := linebreakRegex.Split(d.Content, 2)[0]
		if len(firstLine) > 80 {
			firstLine = firstLine[:77] + "..."
		}
		fmt.Fprintf(&b, "%s\t%s\n", d.UUID, firstLine)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// ---- Main -------------------------------------------------------------------

func main() {
	var args struct {
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
	if p.Subcommand() == nil {
		p.Fail("missing subcommand")
	}
	switch {
	case args.New != nil:
		fmt.Println(new(args.New))
	case args.Prepend != nil:
		fmt.Println(prepend(args.Prepend))
	case args.Append != nil:
		fmt.Println(append(args.Append))
	case args.Replace != nil:
		fmt.Println(replace(args.Replace))
	case args.Edit != nil:
		fmt.Println(edit(args.Edit))
	case args.Get != nil:
		fmt.Println(get(args.Get))
	case args.Select != nil:
		fmt.Println(_select())
	case args.List != nil:
		fmt.Println(list(args.List))
	}
}
