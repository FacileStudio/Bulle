package tui

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
	"github.com/FacileStudio/Bulle/apps/cli/internal/diff"
	"github.com/FacileStudio/Bulle/apps/cli/internal/toolview"
)

// approvalCap is the most input rows the approval block shows before it says
// how many more there are.
const approvalCap = 12

// approvalPaint caches the drawn approval block for one request at one width,
// so a spinner tick does not re-diff a large write on every frame.
type approvalPaint struct {
	req   *approval.Request
	width int
	text  string
}

// approvalView is the pending approval drawn above the prompt: the tool line
// as its header, then the edit's diff, or the full input in a box. The status
// line only says which keys answer it.
func (m *Model) approvalView() string {
	req := m.run.pending
	if req == nil {
		return ""
	}
	if m.run.paint.req != req || m.run.paint.width != m.width {
		m.run.paint = approvalPaint{req: req, width: m.width, text: m.paintApproval(*req)}
	}
	return m.run.paint.text
}

// approvalRows is how many rows the approval block takes, with the blank row
// under it.
func (m *Model) approvalRows() int {
	view := m.approvalView()
	if view == "" {
		return 0
	}
	return lipgloss.Height(view) + 1
}

// paintApproval draws one request at the current width.
func (m *Model) paintApproval(req approval.Request) string {
	width := max(m.width-2, 12)
	header := toolview.ToolLineRunning(toolview.ToolLine(req.Name, string(req.Input), width))
	if req.Name != "run_command" {
		if change, ok := diff.CaptureEdit(m.run.root, req.Name, string(req.Input)); ok {
			if d := diff.RenderDiff(change, width, toolview.ToolBorder(), m.theme.Muted, m.transparent); d != "" {
				return header + "\n" + strings.TrimSuffix(d, "\n")
			}
		}
	}
	return header + "\n" + strings.TrimSuffix(m.inputBox(approvalText(req.Input), width), "\n")
}

// inputBox wraps text into a boxed pane at width, capped at approvalCap rows.
func (m *Model) inputBox(text string, width int) string {
	content := width - 2
	base := toolview.MatchBackground(m.theme.Plain, m.transparent).Width(content)
	var rows []string
	for _, line := range strings.Split(text, "\n") {
		wrapped := ansi.Wrap(strings.ReplaceAll(line, "\t", "    "), content-1, "")
		for _, piece := range strings.Split(wrapped, "\n") {
			rows = append(rows, base.Render(piece))
		}
	}
	if over := len(rows) - approvalCap; over > 0 {
		rows = append(rows[:approvalCap], base.Render("… "+strconv.Itoa(over)+" more rows"))
	}
	return toolview.Box(rows, toolview.ToolBorder(), m.transparent, width)
}

// approvalText is what the box shows: a command verbatim, and any other input
// as indented JSON.
func approvalText(input json.RawMessage) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal(input, &fields) == nil {
		var command string
		if json.Unmarshal(fields["command"], &command) == nil && command != "" {
			return command
		}
	}
	var pretty bytes.Buffer
	if json.Indent(&pretty, input, "", "  ") != nil {
		return string(input)
	}
	return pretty.String()
}
