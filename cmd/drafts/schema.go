package main

import (
	"reflect"
	"strconv"
	"strings"
)

type SchemaDocument struct {
	Name    string           `json:"name"`
	Version string           `json:"version"`
	Tools   []ToolDefinition `json:"tools"`
}

type ToolDefinition struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Parameters   map[string]interface{} `json:"parameters"`
	Aliases      []string               `json:"aliases,omitempty"`
	RawInputFlag string                 `json:"raw_input_flag,omitempty"`
	Interactive  bool                   `json:"interactive,omitempty"`
}

// Schema returns the tool-use formatted schema for all commands.
func getSchema(command string) interface{} {
	tools := getTools()
	schema := SchemaDocument{
		Name:    "drafts",
		Version: version,
		Tools:   tools,
	}

	if command == "" {
		return schema
	}

	command = strings.TrimSpace(command)
	command = strings.TrimPrefix(command, "drafts_")

	for _, tool := range tools {
		if strings.TrimPrefix(tool.Name, "drafts_") == command {
			return tool
		}
		for _, alias := range tool.Aliases {
			if alias == command {
				return tool
			}
		}
	}

	outputError("UNKNOWN_COMMAND",
		"Unknown command: "+command,
		"Use 'drafts schema' to see all available commands")
	return nil
}

func getTools() []ToolDefinition {
	return []ToolDefinition{
		newTool("drafts_new", "Create a new draft in Drafts.app", CreateRequest{}, []string{"create"}, "--input", false),
		newTool("drafts_get", "Get a draft by UUID, returns full draft metadata", GetRequest{}, nil, "", false),
		newTool("drafts_list", "List drafts with filtering, workspace scoping, and token-aware summaries by default", ListRequest{}, nil, "", false),
		newTool("drafts_append", "Append text to an existing draft", ModifyRequest{}, nil, "--input", false),
		newTool("drafts_prepend", "Prepend text to an existing draft", ModifyRequest{}, nil, "--input", false),
		newTool("drafts_replace", "Replace the content of an existing draft", ReplaceRequest{}, nil, "--input", false),
		newTool("drafts_edit", "Open a draft in $EDITOR and replace its content with the edited result", GetRequest{}, nil, "", true),
		newTool("drafts_select", "Interactively select the active draft with fzf", struct{}{}, nil, "", true),
		newTool("drafts_flag", "Flag a draft", UUIDRequest{}, nil, "--input", false),
		newTool("drafts_unflag", "Unflag a draft", UUIDRequest{}, nil, "--input", false),
		newTool("drafts_workspace", "Show current workspace, list all workspaces, or open a workspace by name", WorkspaceRequest{}, nil, "", false),
		newTool("drafts_actions", "List available Drafts actions, optionally filtered by substring", ActionsRequest{}, nil, "", false),
		newTool("drafts_run", "Run a Drafts action on text or an existing draft", RunRequest{}, nil, "--input", false),
		newTool("drafts_info", "Get environment information and diagnostics. Use verbose mode for actions, tags, and workspaces; use test_permissions to verify automation access.", InfoRequest{}, nil, "", false),
		newTool("drafts_schema", "Return the machine-readable command contract for the full CLI or a single command alias.", SchemaRequest{}, nil, "", false),
		newTool("drafts_upgrade", "Upgrade to the latest version from GitHub releases.", struct{}{}, nil, "", false),
		newTool("drafts_version", "Show current CLI version information including OS and architecture.", struct{}{}, nil, "", false),
	}
}

func newTool(name, description string, request interface{}, aliases []string, rawInputFlag string, interactive bool) ToolDefinition {
	return ToolDefinition{
		Name:         name,
		Description:  description,
		Parameters:   buildParametersSchema(request),
		Aliases:      aliases,
		RawInputFlag: rawInputFlag,
		Interactive:  interactive,
	}
}

func buildParametersSchema(request interface{}) map[string]interface{} {
	t := reflect.TypeOf(request)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	properties := map[string]interface{}{}
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name, omitempty := parseJSONField(field.Tag.Get("json"))
		if name == "" {
			continue
		}

		property := map[string]interface{}{
			"type": fieldJSONType(field.Type),
		}

		if field.Type.Kind() == reflect.Slice {
			property["items"] = map[string]interface{}{"type": fieldJSONType(field.Type.Elem())}
		}

		if desc := field.Tag.Get("desc"); desc != "" {
			property["description"] = desc
		}

		if enumTag := field.Tag.Get("enum"); enumTag != "" {
			property["enum"] = strings.Split(enumTag, ",")
		}

		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			property["default"] = parseDefaultValue(field.Type, defaultTag)
		}

		properties[name] = property
		if !omitempty {
			required = append(required, name)
		}
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func parseJSONField(tag string) (string, bool) {
	if tag == "" || tag == "-" {
		return "", false
	}

	parts := strings.Split(tag, ",")
	name := parts[0]
	omitempty := false
	for _, option := range parts[1:] {
		if option == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty
}

func fieldJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice:
		return "array"
	default:
		return "string"
	}
}

func parseDefaultValue(t reflect.Type, value string) interface{} {
	switch t.Kind() {
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return value
}
