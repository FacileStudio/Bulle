package suite

import (
	"context"

	"github.com/FacileStudio/nacelle"
)

// CapsuleCreateName is Bulle's own tool that seals text into a new capsule.
const CapsuleCreateName = "capsule_create"

const capsuleCreateUsage = "Seal text into a new Capsule: an end-to-end encrypted link that destroys itself once read " +
	"(burn_after_read, true by default) or when it expires (expires_in: 1h, 24h, 7d or 30d; 24h by default). " +
	"Bulle encrypts the text itself, so Capsule never sees it. Answers the link to hand to the user, the capsule id " +
	"and the delete token capsule_revoke takes."

const capsuleCreateNote = "To share something privately, call capsule_create: it seals the text and returns a link " +
	"whose part after # is the only key. Give that link to the user alone, in your answer. Never write it to memory, " +
	"an artifact, a task or any other tool unless the user asked you to send it there, and never repeat the text you sealed."

type capsuleCreateInput struct {
	Text          string `json:"text" jsonschema:"required,description=The text to seal. It is encrypted before it leaves Bulle."`
	ExpiresIn     string `json:"expires_in,omitempty" jsonschema:"enum=1h,enum=24h,enum=7d,enum=30d,description=When the capsule expires. Defaults to 24h."`
	BurnAfterRead *bool  `json:"burn_after_read,omitempty" jsonschema:"description=Destroy the capsule on its first read. Defaults to true."`
}

func capsuleTools(set *toolSet, sealer *Sealer) {
	tool, err := nacelle.NewTool(CapsuleCreateName, capsuleCreateUsage, func(ctx context.Context, in capsuleCreateInput) (string, error) {
		expires := in.ExpiresIn
		if expires == "" {
			expires = "24h"
		}
		burn := in.BurnAfterRead == nil || *in.BurnAfterRead
		return jsonResult(sealer.Seal(ctx, in.Text, expires, burn))
	})
	set.add(tool, err)
}
