package runner

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

const (
	profileLimit   = 4 << 10
	profileTimeout = 3 * time.Second
	profileHeading = "What you know about the user (the memory page " + suite.ProfilePage + "; correct it with memory_write when the user corrects you):"
	profileClipped = "\n[" + suite.ProfilePage + " is longer; read it with memory_read for the rest]"
)

func (r *Runner) myceliumEnabled() bool {
	return r.cfg.Suite != nil && r.cfg.Suite.Config().MyceliumEnabled()
}

// memoryProfile is the section of a chat's system prompt that holds the
// space's profile.md, capped at 4 KB. Only web and matrix chats get it. A
// missing page, a slow or failing Mycelium, or no Mycelium at all answers an
// empty section: the profile never fails or holds up a run.
func (r *Runner) memoryProfile(ctx context.Context, userID int64, channel string, space store.Space) string {
	if !store.Conversational(channel) || !r.myceliumEnabled() {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, profileTimeout)
	defer cancel()
	raw, err := suite.NewMycelium(r.cfg.Suite).ReadPage(ctx, userID, suite.SpaceDefaults(space), r.suiteRuntime(userID), suite.ProfilePage)
	if err != nil {
		r.cfg.Logger.Debug("reading the memory profile", "user", userID, "error", err)
		return ""
	}
	text := capProfile(string(raw))
	if text == "" {
		return ""
	}
	return profileHeading + "\n\n" + text
}

func capProfile(text string) string {
	text = strings.TrimSpace(strings.ToValidUTF8(text, ""))
	if len(text) <= profileLimit {
		return text
	}
	cut := profileLimit - len(profileClipped)
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	head := text[:cut]
	if line := strings.LastIndexByte(head, '\n'); line > cut/2 {
		head = head[:line]
	}
	return strings.TrimSpace(head) + profileClipped
}
