package sandbox

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/FacileStudio/nacelle"
)

type editInput struct {
	Path       string `json:"path" jsonschema:"required,description=Path to the file relative to the working directory"`
	OldString  string `json:"old_string" jsonschema:"required,description=The exact text to replace. Include enough surrounding lines to make it unique in the file unless replace_all is set"`
	NewString  string `json:"new_string" jsonschema:"required,description=The text to put in its place"`
	ReplaceAll bool   `json:"replace_all,omitempty" jsonschema:"description=Replace every occurrence instead of requiring exactly one"`
}

func buildEditTool(s *remoteSession) (nacelle.Tool, error) {
	return nacelle.NewTool("edit_file",
		"Replace an exact piece of text in a file. The old text must appear exactly once unless replace_all is set: include the lines around it until it does. Read the file first.",
		func(ctx context.Context, in editInput) (string, error) {
			return runEditFile(ctx, s, in)
		})
}

func squeezeWhitespace(text string) string {
	return string(bytes.Join(bytes.Fields([]byte(text)), []byte(" ")))
}

func nearMissText(content, oldText string) string {
	if strings.Contains(squeezeWhitespace(content), squeezeWhitespace(oldText)) {
		return " - the text is there but the whitespace differs; copy it exactly as read_file showed it, without the line-number column"
	}
	return ""
}

func replaceText(content string, in editInput) (string, error) {
	oldText, newText := in.OldString, in.NewString
	if oldText == "" {
		return "", fmt.Errorf("the text to replace is empty; use write_file to create a file")
	}
	if oldText == newText {
		return "", fmt.Errorf("the replacement is identical to the original")
	}
	count := strings.Count(content, oldText)
	if in.ReplaceAll && count > 0 {
		return strings.ReplaceAll(content, oldText, newText), nil
	}
	if count == 1 {
		return strings.Replace(content, oldText, newText, 1), nil
	}
	if count == 0 {
		return "", fmt.Errorf("that exact text is not in the file%s", nearMissText(content, oldText))
	}
	return "", fmt.Errorf("that text appears %d times; include more surrounding lines so it matches only the one you mean", count)
}

func runEditFile(ctx context.Context, s *remoteSession, in editInput) (string, error) {
	targetPath := resolveRemotePath(s.opts.WorkDir, in.Path)
	readCmd := fmt.Sprintf("cat -- %s", quoteArg(targetPath))
	raw, err := s.runSSH(ctx, readCmd)
	if err != nil {
		return "", fmt.Errorf("edit_file: %w (%s)", err, strings.TrimSpace(string(raw)))
	}
	edited, err := replaceText(string(raw), in)
	if err != nil {
		return "", fmt.Errorf("%s: %w", in.Path, err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(edited))
	writeCmd := fmt.Sprintf("printf '%%s' %s | base64 -d > %s", quoteArg(encoded), quoteArg(targetPath))
	out, err := s.runSSH(ctx, writeCmd)
	if err != nil {
		return "", fmt.Errorf("edit_file: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return fmt.Sprintf("edited %s", in.Path), nil
}
