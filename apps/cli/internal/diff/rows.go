package diff

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FacileStudio/Bulle/apps/cli/internal/theme"
	"github.com/FacileStudio/Bulle/apps/cli/internal/toolview"
)

// recap is the box's footer: one margin space, then "+x -y", additions counted
// in green and removals in red, sharing the block background with the header.
// Each figure fragment ends in a full reset, so the space joining them must
// carry the block background itself — a literal space would render bare on the
// terminal's default background, a visibly different patch in the pane.
func recap(change EditChange, muted lipgloss.Style, content int, transparent bool) string {
	added, removed := CountChange(change)
	var parts []string
	if added > 0 {
		parts = append(parts, toolview.MatchBackground(theme.Fore(theme.Ink.Success), transparent).Render("+"+strconv.Itoa(added)))
	}
	if removed > 0 {
		parts = append(parts, toolview.MatchBackground(theme.Fore(theme.Ink.Danger), transparent).Render("-"+strconv.Itoa(removed)))
	}
	gap := block(muted, transparent).Render(" ")
	return cell(block(muted, transparent), " "+strings.Join(parts, gap), content)
}

// renderBlock appends one hunk's rows to the box, stopping at the display cap
// with a marker saying the diff was cut rather than pretending it wasn't. It
// returns the grown slice, how many rows are through, and whether the cap was
// reached.
func renderBlock(rows []string, blk []diffOp, nums []int, pane diffPane, shown int) ([]string, int, bool) {
	for i, op := range blk {
		if shown >= shownDiffLines {
			rows = append(rows, cell(pane.Backdrop, strings.Repeat(" ", pane.Gutter+3)+"… more", pane.Content))
			return rows, shown, true
		}
		var row string
		switch op.kind {
		case '-':
			row = cell(removedStyle(), gutter(nums[i], pane.Gutter, "-")+truncate(op.text, pane.Content-pane.Gutter-3), pane.Content)
		case '+':
			row = cell(addedStyle(), gutter(nums[i], pane.Gutter, "+")+truncate(op.text, pane.Content-pane.Gutter-3), pane.Content)
		default:
			row = cell(pane.Backdrop, gutter(nums[i], pane.Gutter, " ")+truncate(op.text, pane.Content-pane.Gutter-3), pane.Content)
		}
		rows = append(rows, row)
		shown++
	}
	return rows, shown, false
}
