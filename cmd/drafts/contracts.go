package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nerveband/drafts-applescript-cli/pkg/drafts"
)

type DraftView struct {
	UUID              string   `json:"uuid"`
	Title             string   `json:"title"`
	Tags              []string `json:"tags"`
	IsFlagged         bool     `json:"isFlagged"`
	IsArchived        bool     `json:"isArchived"`
	IsTrashed         bool     `json:"isTrashed"`
	Folder            string   `json:"folder"`
	CreatedAt         string   `json:"createdAt"`
	ModifiedAt        string   `json:"modifiedAt"`
	Permalink         string   `json:"permalink"`
	Content           *string  `json:"content,omitempty"`
	CreatedLatitude   *float64 `json:"createdLatitude,omitempty"`
	CreatedLongitude  *float64 `json:"createdLongitude,omitempty"`
	ModifiedLatitude  *float64 `json:"modifiedLatitude,omitempty"`
	ModifiedLongitude *float64 `json:"modifiedLongitude,omitempty"`
}

type ListResult struct {
	Drafts    []DraftView `json:"drafts"`
	Count     int         `json:"count"`
	Filter    string      `json:"filter"`
	Limit     int         `json:"limit"`
	Full      bool        `json:"full"`
	Search    string      `json:"search,omitempty"`
	Workspace string      `json:"workspace,omitempty"`
}

type RunResult struct {
	Action       string     `json:"action"`
	UUID         string     `json:"uuid"`
	CreatedDraft bool       `json:"createdDraft"`
	Draft        *DraftView `json:"draft,omitempty"`
}

type ActionsResult struct {
	Actions []string `json:"actions"`
	Count   int      `json:"count"`
	Search  string   `json:"search,omitempty"`
}

type WorkspaceResult struct {
	Current    string   `json:"current,omitempty"`
	Opened     string   `json:"opened,omitempty"`
	Workspaces []string `json:"workspaces,omitempty"`
	Count      int      `json:"count,omitempty"`
}

type CreateRequest struct {
	Content string   `json:"content" desc:"The draft content"`
	Tags    []string `json:"tags,omitempty" desc:"Tags to apply to the draft"`
	Folder  string   `json:"folder,omitempty" enum:"inbox,archive" default:"inbox" desc:"Folder to create draft in"`
	Flagged bool     `json:"flagged,omitempty" default:"false" desc:"Whether to flag the draft"`
	Action  string   `json:"action,omitempty" desc:"Action name to run after creation"`
}

type ModifyRequest struct {
	UUID    string   `json:"uuid,omitempty" desc:"UUID of the draft (omit for active draft)"`
	Content string   `json:"content" desc:"Text to append or prepend"`
	Tags    []string `json:"tags,omitempty" desc:"Tags to add"`
	Action  string   `json:"action,omitempty" desc:"Action name to run after modification"`
}

type ReplaceRequest struct {
	UUID    string `json:"uuid,omitempty" desc:"UUID of the draft (omit for active draft)"`
	Content string `json:"content" desc:"New content for the draft"`
}

type GetRequest struct {
	UUID string `json:"uuid,omitempty" desc:"UUID of the draft (omit for active draft)"`
}

type ListRequest struct {
	Filter    string   `json:"filter,omitempty" enum:"inbox,flagged,archive,trash,all" default:"inbox" desc:"Filter drafts by folder"`
	Tags      []string `json:"tags,omitempty" desc:"Filter by tags"`
	Search    string   `json:"search,omitempty" desc:"Search draft content"`
	Workspace string   `json:"workspace,omitempty" desc:"Filter by workspace name"`
	Limit     int      `json:"limit,omitempty" default:"20" desc:"Maximum drafts to return (0 for all)"`
	Full      bool     `json:"full,omitempty" default:"false" desc:"Include full draft content and location fields"`
}

type RunRequest struct {
	Action  string `json:"action" desc:"Name of the action to run"`
	Content string `json:"content,omitempty" desc:"Text to process when not targeting an existing draft"`
	UUID    string `json:"uuid,omitempty" desc:"UUID of the draft to run the action on"`
}

type UUIDRequest struct {
	UUID string `json:"uuid,omitempty" desc:"UUID of the draft (omit for active draft)"`
}

type WorkspaceRequest struct {
	List bool   `json:"list,omitempty" default:"false" desc:"List all workspaces instead of showing current"`
	Open string `json:"open,omitempty" desc:"Open a workspace by name"`
}

type ActionsRequest struct {
	Search string `json:"search,omitempty" desc:"Filter action names by substring"`
}

type InfoRequest struct {
	Verbose         bool `json:"verbose,omitempty" default:"false" desc:"Show full lists of actions, tags, and workspaces"`
	TestPermissions bool `json:"test_permissions,omitempty" default:"false" desc:"Test read, write, and action permissions"`
}

type SchemaRequest struct {
	Command string `json:"command,omitempty" desc:"Command name or alias (omit for the full schema)"`
}

func toDraftView(d drafts.Draft, full bool) DraftView {
	view := DraftView{
		UUID:       d.UUID,
		Title:      d.Title,
		Tags:       d.Tags,
		IsFlagged:  d.IsFlagged,
		IsArchived: d.IsArchived,
		IsTrashed:  d.IsTrashed,
		Folder:     d.Folder,
		CreatedAt:  d.CreatedAt,
		ModifiedAt: d.ModifiedAt,
		Permalink:  d.Permalink,
	}

	if full {
		content := d.Content
		createdLatitude := d.CreatedLatitude
		createdLongitude := d.CreatedLongitude
		modifiedLatitude := d.ModifiedLatitude
		modifiedLongitude := d.ModifiedLongitude
		view.Content = &content
		view.CreatedLatitude = &createdLatitude
		view.CreatedLongitude = &createdLongitude
		view.ModifiedLatitude = &modifiedLatitude
		view.ModifiedLongitude = &modifiedLongitude
	}

	return view
}

func decodeJSONInput(raw string, dest interface{}) error {
	var data []byte
	var err error

	if raw == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		data = []byte(raw)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}

	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("input must contain a single JSON object")
	}

	return nil
}

func resolveCreateRequest(param *NewCmd) (CreateRequest, error) {
	if param.Input != "" {
		if param.Message != "" || len(param.Tag) > 0 || param.Archive || param.Flagged || param.Action != "" {
			return CreateRequest{}, fmt.Errorf("cannot combine --input with positional or flag arguments")
		}
		var request CreateRequest
		if err := decodeJSONInput(param.Input, &request); err != nil {
			return CreateRequest{}, err
		}
		if request.Folder == "" {
			request.Folder = "inbox"
		}
		if request.Folder != "inbox" && request.Folder != "archive" {
			return CreateRequest{}, fmt.Errorf("invalid folder: %s", request.Folder)
		}
		return request, nil
	}

	content, err := readTextInput(param.Message)
	if err != nil {
		return CreateRequest{}, err
	}

	request := CreateRequest{
		Content: content,
		Tags:    param.Tag,
		Folder:  "inbox",
		Flagged: param.Flagged,
		Action:  param.Action,
	}
	if param.Archive {
		request.Folder = "archive"
	}
	return request, nil
}

func resolveModifyRequest(input string, content string, tags []string, action string, uuid string) (ModifyRequest, error) {
	if input != "" {
		if content != "" || len(tags) > 0 || action != "" || uuid != "" {
			return ModifyRequest{}, fmt.Errorf("cannot combine --input with positional or flag arguments")
		}
		var request ModifyRequest
		if err := decodeJSONInput(input, &request); err != nil {
			return ModifyRequest{}, err
		}
		return request, nil
	}

	text, err := readTextInput(content)
	if err != nil {
		return ModifyRequest{}, err
	}

	return ModifyRequest{
		UUID:    uuid,
		Content: text,
		Tags:    tags,
		Action:  action,
	}, nil
}

func resolveReplaceRequest(param *ReplaceCmd) (ReplaceRequest, error) {
	if param.Input != "" {
		if param.Message != "" || param.UUID != "" {
			return ReplaceRequest{}, fmt.Errorf("cannot combine --input with positional or flag arguments")
		}
		var request ReplaceRequest
		if err := decodeJSONInput(param.Input, &request); err != nil {
			return ReplaceRequest{}, err
		}
		return request, nil
	}

	content, err := readTextInput(param.Message)
	if err != nil {
		return ReplaceRequest{}, err
	}

	return ReplaceRequest{
		UUID:    param.UUID,
		Content: content,
	}, nil
}

func resolveRunRequest(param *RunCmd) (RunRequest, error) {
	if param.Input != "" {
		if param.Action != "" || param.Text != "" || param.UUID != "" {
			return RunRequest{}, fmt.Errorf("cannot combine --input with positional or flag arguments")
		}
		var request RunRequest
		if err := decodeJSONInput(param.Input, &request); err != nil {
			return RunRequest{}, err
		}
		return request, nil
	}

	request := RunRequest{
		Action: param.Action,
		UUID:   param.UUID,
	}
	if param.UUID == "" {
		text, err := readTextInput(param.Text)
		if err != nil {
			return RunRequest{}, err
		}
		request.Content = text
	}
	return request, nil
}

func resolveUUIDRequest(input, uuid string) (UUIDRequest, error) {
	if input != "" {
		if uuid != "" {
			return UUIDRequest{}, fmt.Errorf("cannot combine --input with positional or flag arguments")
		}
		var request UUIDRequest
		if err := decodeJSONInput(input, &request); err != nil {
			return UUIDRequest{}, err
		}
		return request, nil
	}

	return UUIDRequest{UUID: uuid}, nil
}

func filterActions(actions []string, query string) []string {
	if query == "" {
		return actions
	}

	filtered := make([]string, 0, len(actions))
	needle := strings.ToLower(query)
	for _, action := range actions {
		if strings.Contains(strings.ToLower(action), needle) {
			filtered = append(filtered, action)
		}
	}
	return filtered
}
