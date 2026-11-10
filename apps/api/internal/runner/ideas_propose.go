package runner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

// ProposeToolName is the host tool an ideas batch records its cards with.
const ProposeToolName = "idea_propose"

const (
	sourceLabelLimit = 120
	sourceRefLimit   = 200
	fingerprintLimit = 200
)

const proposeDescription = "Record one idea for the user as a card. Call it once per idea, 3 to 7 times in all. " +
	"The card shows title and why; prompt is the first message the user sends, as the user, to start a chat where you do it; " +
	"draft is optional text you prepared (a reply, an agenda), shown read-only. " +
	"fingerprint is a stable key for what the idea is about, like courrier:5/email:812/reply or opus:task:44/due. " +
	"A refused call says why: fix it or pick something else."

type proposeInput struct {
	Kind        string             `json:"kind" jsonschema:"required,enum=mail,enum=task,enum=event,enum=follow_up,enum=research,enum=other"`
	Title       string             `json:"title" jsonschema:"required,description=Short card title of at most 120 characters"`
	Why         string             `json:"why" jsonschema:"required,description=One concrete line on why now grounded in something you read; at most 280 characters"`
	Prompt      string             `json:"prompt" jsonschema:"required,description=The user's first chat message that asks you to do it; at most 2000 characters"`
	Draft       string             `json:"draft,omitempty" jsonschema:"description=Optional prepared text of at most 4000 characters"`
	Sources     []store.IdeaSource `json:"sources,omitempty" jsonschema:"description=At most 5 items the idea came from; each with a label and a ref"`
	Fingerprint string             `json:"fingerprint" jsonschema:"required,description=Stable key for what the idea is about"`
	App         string             `json:"app,omitempty" jsonschema:"enum=courrier,enum=agenda,enum=opus,enum=sablier,enum=mycelium,enum=nuage,enum=glouton,enum=bulle,description=The app the user does it in; taken from the fingerprint prefix when omitted"`
}

func (r *Runner) proposeTool(run *ideaRun) (nacelle.Tool, error) {
	return nacelle.NewTool(ProposeToolName, proposeDescription, func(ctx context.Context, in proposeInput) (string, error) {
		return r.propose(ctx, run, in)
	})
}

func (r *Runner) propose(ctx context.Context, run *ideaRun, in proposeInput) (string, error) {
	idea, err := ideaFrom(in)
	if err != nil {
		return "", err
	}
	run.mu.Lock()
	defer run.mu.Unlock()
	if run.count >= store.IdeaBatchLimit {
		return "", fmt.Errorf("this batch already has %d ideas; stop proposing and end your turn", store.IdeaBatchLimit)
	}
	now := r.now()
	taken, err := r.cfg.Store.SuppressedFingerprints(ctx, run.batch.UserID, run.space.ID, now)
	if err != nil {
		return "", err
	}
	if slices.Contains(taken, idea.Fingerprint) {
		return "", errors.New("already suggested, pick something else")
	}
	idea.UserID, idea.SpaceID, idea.BatchID = run.batch.UserID, run.space.ID, run.batch.ID
	idea.App = ideaApp(run.batch.App, in.App, idea.Fingerprint)
	idea.ExpiresAt = now.Add(store.IdeaLifetime)
	if _, err := r.cfg.Store.CreateIdea(ctx, idea); err != nil {
		return "", err
	}
	run.count++
	return fmt.Sprintf("Saved idea %d of at most %d.", run.count, store.IdeaBatchLimit), nil
}

// ideaApp is the app an idea is filed under: the batch's app when the batch
// is scoped, else the app the model named, else the fingerprint's prefix.
func ideaApp(batchApp, named, fingerprint string) string {
	if batchApp != "" {
		return batchApp
	}
	if named = strings.ToLower(strings.TrimSpace(named)); slices.Contains(store.IdeaApps(), named) {
		return named
	}
	return store.IdeaAppOf(fingerprint)
}

func ideaFrom(in proposeInput) (store.Idea, error) {
	idea := store.Idea{
		Kind: strings.TrimSpace(in.Kind), Title: strings.TrimSpace(in.Title), Why: strings.TrimSpace(in.Why),
		Prompt: strings.TrimSpace(in.Prompt), Draft: strings.TrimSpace(in.Draft),
		Fingerprint: strings.TrimSpace(in.Fingerprint), Status: store.IdeaNew,
	}
	if !slices.Contains(store.IdeaKinds(), idea.Kind) {
		return store.Idea{}, fmt.Errorf("kind must be one of %s", strings.Join(store.IdeaKinds(), ", "))
	}
	err := errors.Join(
		textLimit("title", idea.Title, store.IdeaTitleLimit, true),
		textLimit("why", idea.Why, store.IdeaWhyLimit, true),
		textLimit("prompt", idea.Prompt, store.IdeaPromptLimit, true),
		textLimit("draft", idea.Draft, store.IdeaDraftLimit, false),
		textLimit("fingerprint", idea.Fingerprint, fingerprintLimit, true),
	)
	if err != nil {
		return store.Idea{}, err
	}
	idea.Sources, err = sourcesFrom(in.Sources)
	return idea, err
}

func sourcesFrom(in []store.IdeaSource) ([]store.IdeaSource, error) {
	if len(in) > store.IdeaSourcesLimit {
		return nil, fmt.Errorf("at most %d sources", store.IdeaSourcesLimit)
	}
	out := make([]store.IdeaSource, 0, len(in))
	for _, source := range in {
		source.Label, source.Ref = strings.TrimSpace(source.Label), strings.TrimSpace(source.Ref)
		err := errors.Join(
			textLimit("source label", source.Label, sourceLabelLimit, true),
			textLimit("source ref", source.Ref, sourceRefLimit, true),
		)
		if err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	return out, nil
}

func textLimit(field, value string, limit int, required bool) error {
	if required && value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if n := len([]rune(value)); n > limit {
		return fmt.Errorf("%s is %d characters, at most %d", field, n, limit)
	}
	return nil
}
