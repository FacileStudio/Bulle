package sessions

// lastText is the turn's final text part: the answer a model gives after its
// last tool call, without the narration it wrote between steps.
func lastText(texts []string) string {
	if len(texts) == 0 {
		return ""
	}
	return texts[len(texts)-1]
}
