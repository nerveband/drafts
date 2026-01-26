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
	DocsURL string `json:"docs_url,omitempty"`
}

// Error codes and their documentation URLs
var errorDocs = map[string]string{
	"DRAFT_NOT_FOUND":    "https://github.com/nerveband/drafts-applescript-cli#get",
	"UNKNOWN_COMMAND":    "https://github.com/nerveband/drafts-applescript-cli#usage",
	"DRAFTS_NOT_RUNNING": "https://github.com/nerveband/drafts-applescript-cli#troubleshooting",
	"PERMISSION_DENIED":  "https://github.com/nerveband/drafts-applescript-cli#troubleshooting",
	"ACTION_NOT_FOUND":   "https://github.com/nerveband/drafts-applescript-cli#run",
	"PRO_REQUIRED":       "https://github.com/nerveband/drafts-applescript-cli#requirements",
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
	docsURL := errorDocs[code]

	if plainOutput {
		fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", code, message)
		if hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
		}
		if docsURL != "" {
			fmt.Fprintf(os.Stderr, "Docs: %s\n", docsURL)
		}
		fmt.Fprintf(os.Stderr, "\nRun 'drafts info' to check environment status.\n")
		os.Exit(1)
	}

	errInfo := &ErrorInfo{
		Code:    code,
		Message: message,
		Hint:    hint,
	}
	if docsURL != "" {
		errInfo.DocsURL = docsURL
	}

	resp := Response{
		Success: false,
		Error:   errInfo,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(resp)
	os.Exit(1)
}

// OutputErrorWithContext writes an error with additional context
func outputErrorWithContext(code, message, hint string, context map[string]interface{}) {
	docsURL := errorDocs[code]

	if plainOutput {
		fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", code, message)
		if hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
		}
		for k, v := range context {
			fmt.Fprintf(os.Stderr, "%s: %v\n", k, v)
		}
		if docsURL != "" {
			fmt.Fprintf(os.Stderr, "Docs: %s\n", docsURL)
		}
		os.Exit(1)
	}

	errInfo := &ErrorInfo{
		Code:    code,
		Message: message,
		Hint:    hint,
	}
	if docsURL != "" {
		errInfo.DocsURL = docsURL
	}

	resp := Response{
		Success: false,
		Error:   errInfo,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(resp)
	os.Exit(1)
}
