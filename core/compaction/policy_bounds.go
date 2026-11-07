package compaction

// Over reports whether size has crossed the policy's trigger. It is the one
// boundary comparison every caller makes: a size exactly on the trigger is not
// over it, and a policy with no trigger is never over.
func (p Policy) Over(size int64) bool {
	trigger := p.Trigger()
	return trigger > 0 && size > trigger
}

// Resolved is the policy with the shipped defaults filling in every field a
// caller left at zero: the ratio ladder, the tail floor and budget, and the
// pinned head. Window, Reserve and Ceiling stay as given, since zero is a real
// answer for each of them.
func (p Policy) Resolved() Policy {
	if p.Ratios == (Ratios{}) {
		p.Ratios = Ratios{Soft: DefaultSoftRatio, Smart: DefaultSmartRatio}
	}
	if p.KeepTurns <= 0 {
		p.KeepTurns = DefaultKeepTurns
	}
	if p.KeepTokens <= 0 {
		p.KeepTokens = DefaultKeepTokens
	}
	if p.AnchorMessages <= 0 {
		p.AnchorMessages = DefaultAnchorMessages
	}
	return p
}
