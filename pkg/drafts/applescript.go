package drafts

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrDraftNotFound     = errors.New("draft not found")
	ErrActionNotFound    = errors.New("action not found")
	ErrWorkspaceNotFound = errors.New("workspace not found")
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
	// Escape backslashes first, then quotes, then newlines
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
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

func objectExists(objectSpecifier string) (bool, error) {
	script := fmt.Sprintf(`tell application "Drafts"
	return exists %s
end tell`, objectSpecifier)

	output, err := runAppleScript(script)
	if err != nil {
		return false, err
	}
	return output == "true", nil
}

// DraftExists reports whether a Drafts draft exists for the given UUID.
func DraftExists(uuid string) (bool, error) {
	return objectExists(fmt.Sprintf(`draft id "%s"`, escapeForAppleScript(uuid)))
}

// ActionExists reports whether a Drafts action exists by name.
func ActionExists(name string) (bool, error) {
	return objectExists(fmt.Sprintf(`action "%s"`, escapeForAppleScript(name)))
}

// WorkspaceExists reports whether a Drafts workspace exists by name.
func WorkspaceExists(name string) (bool, error) {
	return objectExists(fmt.Sprintf(`workspace "%s"`, escapeForAppleScript(name)))
}

// RunActionOnDraft runs an action on an existing draft.
func RunActionOnDraft(action, uuid string) error {
	exists, err := DraftExists(uuid)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: %s", ErrDraftNotFound, uuid)
	}

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
		return fmt.Errorf("%w: %s", ErrActionNotFound, action)
	}
	return nil
}
