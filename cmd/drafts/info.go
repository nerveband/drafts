package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ernstwi/drafts/pkg/drafts"
)

// InfoResult holds all diagnostic information
type InfoResult struct {
	CLI       CLIInfo       `json:"cli"`
	DraftsApp DraftsAppInfo `json:"drafts_app"`
	Counts    DraftCounts   `json:"counts"`
	Actions   []string      `json:"available_actions,omitempty"`
	Tags      []string      `json:"available_tags,omitempty"`
	Workspaces []string     `json:"workspaces,omitempty"`
	Recent    []RecentDraft `json:"recent_drafts,omitempty"`
	Permissions *PermissionTest `json:"permissions,omitempty"`
}

type CLIInfo struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

type DraftsAppInfo struct {
	Running bool   `json:"running"`
	Version string `json:"version,omitempty"`
	Build   string `json:"build,omitempty"`
	Pro     bool   `json:"pro"`
}

type DraftCounts struct {
	Inbox    int `json:"inbox"`
	Flagged  int `json:"flagged"`
	Archive  int `json:"archive"`
	Trash    int `json:"trash"`
	All      int `json:"all"`
}

type RecentDraft struct {
	UUID     string `json:"uuid"`
	Title    string `json:"title"`
	Modified string `json:"modified"`
}

type PermissionTest struct {
	Read   PermissionResult `json:"read"`
	Write  PermissionResult `json:"write"`
	Action PermissionResult `json:"action"`
}

type PermissionResult struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message,omitempty"`
}

func runInfo(param *InfoCmd) interface{} {
	result := InfoResult{
		CLI: CLIInfo{
			Version: version,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
	}

	// Check if Drafts is running
	result.DraftsApp.Running = isDraftsRunning()

	if !result.DraftsApp.Running {
		if plainOutput {
			return formatInfoPlain(result, param.Verbose)
		}
		return result
	}

	// Get Drafts app info
	result.DraftsApp.Version, result.DraftsApp.Build = getDraftsVersion()
	result.DraftsApp.Pro = checkDraftsPro()

	// Get draft counts (archive/trash only with --verbose, as they can be slow)
	result.Counts = getDraftCounts(param.Verbose)

	// Get verbose info if requested (tags, actions, workspaces can be slow)
	if param.Verbose {
		result.Tags = getAvailableTags()
		result.Actions = getAvailableActions()
		result.Workspaces = getAvailableWorkspaces()
	}

	// Get recent drafts (last 5)
	result.Recent = getRecentDrafts(5)

	// Test permissions if requested
	if param.TestPermissions {
		result.Permissions = testPermissions()
	}

	if plainOutput {
		return formatInfoPlain(result, param.Verbose)
	}
	return result
}

func isDraftsRunning() bool {
	script := `tell application "System Events" to (name of processes) contains "Drafts"`
	output, err := runSimpleAppleScript(script)
	if err != nil {
		return false
	}
	return output == "true"
}

func getDraftsVersion() (string, string) {
	script := `tell application "Drafts" to return version`
	output, err := runSimpleAppleScript(script)
	if err != nil {
		return "", ""
	}
	return output, ""
}

func checkDraftsPro() bool {
	// Try to access an action - this requires Pro
	script := `tell application "Drafts"
	try
		set a to first action
		return "true"
	on error
		return "false"
	end try
end tell`
	output, err := runSimpleAppleScript(script)
	if err != nil {
		return false
	}
	return output == "true"
}

func getDraftCounts(includeArchive bool) DraftCounts {
	counts := DraftCounts{}

	// Fast query for inbox and flagged only
	script := `tell application "Drafts"
	set inboxCount to count of (every draft whose folder is inbox)
	set flaggedCount to count of (every draft whose folder is inbox and flagged is true)
	return (inboxCount as string) & "	" & (flaggedCount as string)
end tell`

	output, err := runSimpleAppleScript(script)
	if err != nil {
		return counts
	}

	parts := strings.Split(output, "\t")
	if len(parts) >= 2 {
		fmt.Sscanf(parts[0], "%d", &counts.Inbox)
		fmt.Sscanf(parts[1], "%d", &counts.Flagged)
	}

	// Only get archive/trash counts if requested (slow with large archives)
	if includeArchive {
		archiveScript := `tell application "Drafts"
	set archiveCount to count of (every draft whose folder is archive)
	set trashCount to count of (every draft whose folder is trash)
	return (archiveCount as string) & "	" & (trashCount as string)
end tell`
		archiveOutput, err := runSimpleAppleScript(archiveScript)
		if err == nil {
			archiveParts := strings.Split(archiveOutput, "\t")
			if len(archiveParts) >= 2 {
				fmt.Sscanf(archiveParts[0], "%d", &counts.Archive)
				fmt.Sscanf(archiveParts[1], "%d", &counts.Trash)
			}
		}
		counts.All = counts.Inbox + counts.Archive + counts.Trash
	} else {
		counts.All = counts.Inbox // Only show inbox total when not verbose
	}

	return counts
}

func getAvailableTags() []string {
	// Get tags efficiently - limit to first 50 drafts for performance
	script := `tell application "Drafts"
	set allTags to {}
	set draftList to (every draft whose folder is inbox)
	set maxDrafts to 50
	if (count of draftList) < maxDrafts then
		set maxDrafts to (count of draftList)
	end if
	repeat with i from 1 to maxDrafts
		set d to item i of draftList
		repeat with t in (tags of d)
			if t is not in allTags then
				set end of allTags to (t as string)
			end if
		end repeat
	end repeat
	set output to ""
	repeat with t in allTags
		if output is "" then
			set output to t
		else
			set output to output & "|||" & t
		end if
	end repeat
	return output
end tell`
	output, err := runSimpleAppleScript(script)
	if err != nil || output == "" {
		return []string{}
	}
	tags := strings.Split(output, "|||")
	sort.Strings(tags)
	return tags
}

func getAvailableActions() []string {
	script := `tell application "Drafts"
	set actionNames to {}
	repeat with a in (every action)
		set end of actionNames to (name of a)
	end repeat
	set output to ""
	repeat with n in actionNames
		if output is "" then
			set output to n
		else
			set output to output & "|||" & n
		end if
	end repeat
	return output
end tell`
	output, err := runSimpleAppleScript(script)
	if err != nil || output == "" {
		return []string{}
	}
	actions := strings.Split(output, "|||")
	sort.Strings(actions)
	return actions
}

func getAvailableWorkspaces() []string {
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
	output, err := runSimpleAppleScript(script)
	if err != nil || output == "" {
		return []string{}
	}
	workspaces := strings.Split(output, "|||")
	return workspaces
}

func getRecentDrafts(limit int) []RecentDraft {
	// Get recent drafts efficiently using AppleScript with limit
	script := fmt.Sprintf(`tell application "Drafts"
	set output to ""
	set draftList to (every draft whose folder is inbox)
	set maxDrafts to %d
	if (count of draftList) < maxDrafts then
		set maxDrafts to (count of draftList)
	end if
	repeat with i from 1 to maxDrafts
		set d to item i of draftList
		set line_out to (id of d) & "	" & (title of d) & "	" & ((modifiedAt of d) as string)
		if output is "" then
			set output to line_out
		else
			set output to output & linefeed & line_out
		end if
	end repeat
	return output
end tell`, limit)

	output, err := runSimpleAppleScript(script)
	if err != nil || output == "" {
		return []RecentDraft{}
	}

	lines := strings.Split(output, "\n")
	// First pass: count valid lines
	validCount := 0
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				validCount++
			}
		}
	}

	// Second pass: fill the slice
	recent := make([]RecentDraft, validCount)
	idx := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			title := parts[1]
			if title == "" {
				title = "(untitled)"
			}
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			recent[idx] = RecentDraft{
				UUID:     parts[0],
				Title:    title,
				Modified: parts[2],
			}
			idx++
		}
	}
	return recent
}

func testPermissions() *PermissionTest {
	result := &PermissionTest{}

	// Test read permission
	testDrafts := drafts.Query("", drafts.FilterInbox, drafts.QueryOptions{})
	if len(testDrafts) > 0 {
		result.Read = PermissionResult{Allowed: true, Message: "Can read drafts"}
	} else {
		// Could be empty or error - try to get active
		active := drafts.Active()
		if active != "" {
			result.Read = PermissionResult{Allowed: true, Message: "Can read drafts (inbox empty)"}
		} else {
			result.Read = PermissionResult{Allowed: false, Message: "Cannot read drafts - check Drafts Pro status"}
		}
	}

	// Test write permission - create and immediately trash a test draft
	testUUID := drafts.Create("__drafts_cli_permission_test__", drafts.CreateOptions{})
	if testUUID != "" {
		result.Write = PermissionResult{Allowed: true, Message: "Can create drafts"}
		// Clean up - trash the test draft
		drafts.Trash(testUUID)
	} else {
		result.Write = PermissionResult{Allowed: false, Message: "Cannot create drafts"}
	}

	// Test action permission
	script := `tell application "Drafts"
	try
		set a to first action
		return name of a
	on error
		return ""
	end try
end tell`
	actionName, err := runSimpleAppleScript(script)
	if err == nil && actionName != "" {
		result.Action = PermissionResult{Allowed: true, Message: fmt.Sprintf("Can access actions (e.g., '%s')", actionName)}
	} else {
		result.Action = PermissionResult{Allowed: false, Message: "Cannot access actions - requires Drafts Pro"}
	}

	return result
}

func formatInfoPlain(info InfoResult, verbose bool) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Drafts CLI v%s (%s/%s)\n\n", info.CLI.Version, info.CLI.OS, info.CLI.Arch))

	b.WriteString("Drafts.app Status:\n")
	if info.DraftsApp.Running {
		b.WriteString("  Running:     ✓ Yes\n")
		if info.DraftsApp.Version != "" {
			b.WriteString(fmt.Sprintf("  Version:     %s", info.DraftsApp.Version))
			if info.DraftsApp.Build != "" {
				b.WriteString(fmt.Sprintf(" (build %s)", info.DraftsApp.Build))
			}
			b.WriteString("\n")
		}
		if info.DraftsApp.Pro {
			b.WriteString("  Pro:         ✓ Active\n")
		} else {
			b.WriteString("  Pro:         ✗ Not detected\n")
		}
	} else {
		b.WriteString("  Running:     ✗ No\n")
		b.WriteString("\n⚠️  Drafts must be running for commands to work.\n")
		b.WriteString("   Start it with: open -a Drafts\n")
		return b.String()
	}

	b.WriteString("\nDraft Counts:\n")
	b.WriteString(fmt.Sprintf("  Inbox:       %d\n", info.Counts.Inbox))
	b.WriteString(fmt.Sprintf("  Flagged:     %d\n", info.Counts.Flagged))
	b.WriteString(fmt.Sprintf("  Archive:     %d\n", info.Counts.Archive))
	b.WriteString(fmt.Sprintf("  Trash:       %d\n", info.Counts.Trash))
	b.WriteString(fmt.Sprintf("  Total:       %d\n", info.Counts.All))

	if len(info.Tags) > 0 {
		b.WriteString(fmt.Sprintf("\nAvailable Tags: %d tags\n", len(info.Tags)))
		if verbose {
			for _, tag := range info.Tags {
				b.WriteString(fmt.Sprintf("  - %s\n", tag))
			}
		}
	}

	if verbose && len(info.Actions) > 0 {
		b.WriteString(fmt.Sprintf("\nAvailable Actions: %d actions\n", len(info.Actions)))
		for _, action := range info.Actions {
			b.WriteString(fmt.Sprintf("  - %s\n", action))
		}
	}

	if verbose && len(info.Workspaces) > 0 {
		b.WriteString(fmt.Sprintf("\nWorkspaces: %d\n", len(info.Workspaces)))
		for _, ws := range info.Workspaces {
			b.WriteString(fmt.Sprintf("  - %s\n", ws))
		}
	}

	if len(info.Recent) > 0 {
		b.WriteString("\nRecent Drafts:\n")
		for _, d := range info.Recent {
			b.WriteString(fmt.Sprintf("  - %s (%s)\n", d.Title, d.UUID[:8]))
		}
	}

	if info.Permissions != nil {
		b.WriteString("\nPermission Tests:\n")
		if info.Permissions.Read.Allowed {
			b.WriteString(fmt.Sprintf("  ✓ Read:   %s\n", info.Permissions.Read.Message))
		} else {
			b.WriteString(fmt.Sprintf("  ✗ Read:   %s\n", info.Permissions.Read.Message))
		}
		if info.Permissions.Write.Allowed {
			b.WriteString(fmt.Sprintf("  ✓ Write:  %s\n", info.Permissions.Write.Message))
		} else {
			b.WriteString(fmt.Sprintf("  ✗ Write:  %s\n", info.Permissions.Write.Message))
		}
		if info.Permissions.Action.Allowed {
			b.WriteString(fmt.Sprintf("  ✓ Action: %s\n", info.Permissions.Action.Message))
		} else {
			b.WriteString(fmt.Sprintf("  ✗ Action: %s\n", info.Permissions.Action.Message))
		}
	}

	b.WriteString(fmt.Sprintf("\nDocumentation: %s\n", repoURL))

	return b.String()
}

// runSimpleAppleScript is a helper for info-specific scripts
func runSimpleAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Unused but keeping for potential future use
var _ = time.Now
