package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxArtifactTitle   = 200
	maxArtifactContent = 1 << 20
)

type artifactCreateInput struct {
	Title   string `json:"title" jsonschema:"required,description=Short title, at most 200 characters"`
	Content string `json:"content" jsonschema:"required,description=The whole artifact in markdown, at most 1 MB"`
}

type artifactUpdateInput struct {
	ArtifactID string `json:"artifact_id" jsonschema:"required,description=Artifact id as artifact_create or artifact_list returns it"`
	Title      string `json:"title,omitempty" jsonschema:"description=New title; omit to keep the current one"`
	Content    string `json:"content" jsonschema:"required,description=The whole artifact in markdown; it replaces the old content"`
}

type artifactListInput struct {
	Query string `json:"query,omitempty" jsonschema:"description=Only artifacts whose title or id contains this text"`
}

type artifactReadInput struct {
	ArtifactID string `json:"artifact_id" jsonschema:"required,description=Artifact id as artifact_list returns it"`
}

type artifactRef struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	SpaceID string `json:"space_id"`
}

const artifactNote = "You can also write artifacts (artifact_create, artifact_update, artifact_list, artifact_read): " +
	"markdown documents saved in Mycelium that the user reopens from the Artifacts section. " +
	"Make one only for significant, self-contained content worth coming back to, usually over about 15 lines, " +
	"such as a report, a plan, a comparison or a guide; answer everything else in the chat. " +
	"To change one, read it, then call artifact_update with its artifact_id instead of creating a copy. " +
	"Use Mycelium's rich markdown: tables, `> [!NOTE]` and `> [!WARNING]` alerts, " +
	"```chart:bar fenced charts and ```mermaid diagrams. " +
	"Bulle shows the saved artifact as a card, so say in one sentence what it holds and do not paste it again."

func artifactTools(set *toolSet, m Mycelium, space MemorySpace, bulleSpace string) {
	resolve := space.resolve
	readTool(set, "artifact_list", "List the artifacts of this space, newest first, optionally filtered by title.",
		func(ctx context.Context, in artifactListInput) (string, error) {
			id, err := resolve(ctx)
			if err != nil {
				return "", err
			}
			items, err := m.ListArtifacts(ctx, set.userID, id)
			return jsonResult(filterArtifacts(items, in.Query), err)
		})
	readTool(set, "artifact_read", "Read one artifact's markdown. Read it before artifact_update, which replaces the whole content. artifact_id comes from artifact_list.",
		func(ctx context.Context, in artifactReadInput) (string, error) {
			id, err := resolve(ctx)
			if err != nil {
				return "", err
			}
			return jsonResult(m.GetArtifact(ctx, set.userID, id, in.ArtifactID))
		})
	artifactWriteTools(set, m, resolve, bulleSpace)
}

func artifactWriteTools(set *toolSet, m Mycelium, resolve MemorySpace, bulleSpace string) {
	selfTool(set, "artifact_create", "Save a new markdown artifact the user can reopen from the Artifacts section. Answers its artifact id. To change an artifact that exists, use artifact_update instead of a copy.",
		func(ctx context.Context, in artifactCreateInput) (string, error) {
			title := strings.TrimSpace(in.Title)
			id, err := checkArtifact(ctx, resolve, title, in.Content, true)
			if err != nil {
				return "", err
			}
			saved, err := m.CreateArtifact(ctx, set.userID, id, ArtifactWrite{Title: title, Content: in.Content})
			return artifactResult(saved, bulleSpace, err)
		})
	selfTool(set, "artifact_update", "Replace the content, and optionally the title, of an existing artifact. Use it instead of artifact_create when the user wants a saved artifact changed. Pass the whole new markdown; artifact_id comes from artifact_list or artifact_create.",
		func(ctx context.Context, in artifactUpdateInput) (string, error) {
			title := strings.TrimSpace(in.Title)
			id, err := checkArtifact(ctx, resolve, title, in.Content, false)
			if err != nil {
				return "", err
			}
			saved, err := m.UpdateArtifact(ctx, set.userID, id, in.ArtifactID, ArtifactWrite{Title: title, Content: in.Content})
			return artifactResult(saved, bulleSpace, err)
		})
}

// checkArtifact refuses what Mycelium would refuse before calling it, and a
// missing title on create, so the card's title is always one the model chose.
// It answers the Mycelium space to write to.
func checkArtifact(ctx context.Context, resolve MemorySpace, title, content string, create bool) (string, error) {
	switch {
	case create && title == "":
		return "", fmt.Errorf("title is required")
	case utf8.RuneCountInString(title) > maxArtifactTitle:
		return "", fmt.Errorf("title must be at most %d characters", maxArtifactTitle)
	case strings.TrimSpace(content) == "" || len(content) > maxArtifactContent:
		return "", fmt.Errorf("content must be between 1 byte and %d MB", maxArtifactContent>>20)
	}
	return resolve(ctx)
}

// artifactResult answers the compact JSON the client draws the artifact card
// from. The title is cut to the title limit so the JSON always fits whole in
// the runner's tool output preview.
func artifactResult(saved ArtifactDetail, bulleSpace string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	title := saved.Title
	if utf8.RuneCountInString(title) > maxArtifactTitle {
		title = string([]rune(title)[:maxArtifactTitle])
	}
	data, err := json.Marshal(map[string]artifactRef{"artifact": {ID: saved.ID, Title: title, SpaceID: bulleSpace}})
	return string(data), err
}

func filterArtifacts(items []ArtifactSummary, query string) []ArtifactSummary {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return items
	}
	out := []ArtifactSummary{}
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Title), query) || strings.Contains(strings.ToLower(item.ID), query) {
			out = append(out, item)
		}
	}
	return out
}
