package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/drafts-applescript-cli/pkg/drafts"
)

func readTextInput(text string) (string, error) {
	if text != "" {
		return text, nil
	}
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(stdin), "\n"), nil
}

func resolveActiveUUID(uuid string) (string, error) {
	if uuid != "" {
		return uuid, nil
	}
	return drafts.Active()
}

// Run FZF on input, return UUID.
func fzfUUID(input string) (string, error) {
	line, err := fzf(input)
	if err != nil {
		return "", err
	}
	return strings.Split(line, fmt.Sprintf(" %c ", drafts.Separator))[0], nil
}

// Run FZF on input, return line.
func fzf(input string) (string, error) {
	var result strings.Builder
	cmd := exec.Command("fzf", "--delimiter", "\\|", "--with-nth", "2")
	cmd.Stdout = &result
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader(input)

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 {
			return "", fmt.Errorf("selection cancelled")
		}
		return "", err
	}

	return strings.TrimSpace(result.String()), nil
}

func editor(input string) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name()) // clean up

	if _, err := f.Write([]byte(input)); err != nil {
		return "", err
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(data), "\n"), nil
}
