package drafts

import (
	"fmt"
	"net/url"
	"strings"
)

// ---- Writing drafts ---------------------------------------------------------

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

	if opt.Action != "" {
		RunActionOnDraft(opt.Action, uuid)
	}

	return uuid
}

// Prepend to an existing draft.
func Prepend(uuid, text string, opt ModifyOptions) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set content of d to "%s" & linefeed & (content of d)
end tell`, escapeForAppleScript(uuid), escapeForAppleScript(text))

	runAppleScript(script)

	if len(opt.Tags) > 0 {
		Tag(uuid, opt.Tags...)
	}
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

	if len(opt.Tags) > 0 {
		Tag(uuid, opt.Tags...)
	}
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

// Trash a draft.
func Trash(uuid string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set isTrashed of d to true
end tell`, escapeForAppleScript(uuid))

	runAppleScript(script)
}

// Archive a draft.
func Archive(uuid string) {
	script := fmt.Sprintf(`tell application "Drafts"
	set d to draft id "%s"
	set isArchived of d to true
end tell`, escapeForAppleScript(uuid))

	runAppleScript(script)
}

// Tag adds tags to a draft.
func Tag(uuid string, tags ...string) {
	if len(tags) == 0 {
		return
	}

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

// ---- Reading drafts ---------------------------------------------------------

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

// Query for drafts.
func Query(queryString string, filter Filter, opt QueryOptions) []Draft {
	filterClause := "inbox"
	if filter == FilterArchive {
		filterClause = "archive"
	} else if filter == FilterTrash {
		filterClause = "trash"
	} else if filter == FilterAll {
		filterClause = "all"
	}

	// For "all", we need to get from all folders
	var script string
	if filterClause == "all" {
		script = `tell application "Drafts"
	set output to ""
	set allDrafts to every draft
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
end tell`
	} else {
		script = fmt.Sprintf(`tell application "Drafts"
	set output to ""
	set allDrafts to every draft whose isArchived is %t and isTrashed is %t
	repeat with d in allDrafts
		set folder_name to "%s"
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
end tell`,
			filterClause == "archive",
			filterClause == "trash",
			filterClause)
	}

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

// ---- App state --------------------------------------------------------------

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

// ---- Actions ----------------------------------------------------------------

// RunAction runs an action with text (creates temp draft, runs action).
func RunAction(action, text string) url.Values {
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
