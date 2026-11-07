package suite

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const opusColumnsPath = "/api/column/%d"

var opusVirtualStatuses = []string{"planned", "archived"}

var opusFinalAliases = []string{"completed", "complete", "closed", "finished"}

// OpusColumn is one board column of an Opus project. Its slug is the status a
// task in that column carries.
type OpusColumn struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	IsFinal bool   `json:"isFinal"`
}

func (o Opus) columns(ctx context.Context, userID, projectID int64) ([]OpusColumn, error) {
	raw, err := o.gw.Get(ctx, userID, AppOpus, fmt.Sprintf(opusColumnsPath, projectID), nil)
	if err != nil {
		return nil, err
	}
	var columns []OpusColumn
	if err := decodeInto(raw, &columns); err != nil {
		return nil, fmt.Errorf("opus: reading the columns of project %d: %w", projectID, err)
	}
	return columns, nil
}

func (o Opus) resolveStatus(ctx context.Context, userID, projectID int64, status string) (string, error) {
	columns, err := o.columns(ctx, userID, projectID)
	if err != nil {
		return "", err
	}
	if slug, ok := matchStatus(columns, status); ok {
		return slug, nil
	}
	return "", statusError(status, opusStatuses(columns))
}

// matchStatus finds the column a status names. Case, underscores, spaces and
// hyphens are ignored, the slug wins over the name, and completed or closed
// mean the project's final column.
func matchStatus(columns []OpusColumn, status string) (string, bool) {
	key := statusKey(status)
	if key == "" {
		return "", false
	}
	for _, column := range columns {
		if statusKey(column.Slug) == key || (column.IsFinal && slices.Contains(opusFinalAliases, key)) {
			return column.Slug, true
		}
	}
	for _, column := range columns {
		if statusKey(column.Name) == key {
			return column.Slug, true
		}
	}
	if slices.Contains(opusVirtualStatuses, key) {
		return key, true
	}
	return "", false
}

func statusKey(value string) string {
	return strings.NewReplacer("_", "", " ", "", "-", "").Replace(strings.ToLower(strings.TrimSpace(value)))
}

func opusStatuses(columns []OpusColumn) []string {
	out := make([]string, 0, len(columns)+len(opusVirtualStatuses))
	for _, column := range columns {
		out = append(out, column.Slug)
	}
	return append(out, opusVirtualStatuses...)
}

func statusError(status string, valid []string) error {
	return fmt.Errorf("opus: status %q matches no column of this project; valid statuses are %s (opus_get_project lists each project's columns)",
		status, strings.Join(valid, ", "))
}

func filterNamed(items []opusNamed, kind, ref string) ([]opusNamed, error) {
	ref = strings.TrimSpace(ref)
	var out []opusNamed
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
		if strconv.FormatInt(item.ID, 10) == ref || strings.EqualFold(item.Name, ref) || strings.EqualFold(item.Slug, ref) {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("opus: no %s is named %q; the %ss are %s", kind, ref, kind, strings.Join(names, ", "))
	}
	return out, nil
}
