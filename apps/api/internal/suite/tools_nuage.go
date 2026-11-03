package suite

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const maxSavedText = 1 << 20

type filesListInput struct {
	Folder int64 `json:"folder,omitempty" jsonschema:"description=Folder id to list; omit for the root of the user's files"`
}

type filesSearchInput struct {
	Query string `json:"query" jsonschema:"required,description=Text to look for in file and folder names"`
}

type filesReadInput struct {
	ID int64 `json:"id" jsonschema:"required,description=File id"`
}

type filesSaveInput struct {
	Folder       int64  `json:"folder,omitempty" jsonschema:"description=Folder id to save into; omit for the root"`
	Name         string `json:"name,omitempty" jsonschema:"description=File name with its extension; defaults to the attachment's name"`
	Content      string `json:"content,omitempty" jsonschema:"description=Text content of a new file"`
	AttachmentID string `json:"attachment_id,omitempty" jsonschema:"description=Id of a file the user attached in this chat, saved instead of content"`
}

type attachmentReader func(ctx context.Context, id string) (File, error)

func nuageTools(set *toolSet, n Nuage, attachment attachmentReader) {
	readTool(set, "files_list", "List the folders and files in the user's Nuage storage, at the root or inside a folder.",
		func(ctx context.Context, in filesListInput) (string, error) {
			return jsonResult(n.List(ctx, set.userID, in.Folder))
		})
	readTool(set, "files_search", "Search the user's Nuage files and folders by name. Answers each hit with its path.",
		func(ctx context.Context, in filesSearchInput) (string, error) {
			return jsonResult(n.Search(ctx, set.userID, in.Query))
		})
	readTool(set, "files_read", "Read a file from Nuage: its metadata, and its text when it is a text file (first 100 KB).",
		func(ctx context.Context, in filesReadInput) (string, error) {
			return jsonResult(n.Read(ctx, set.userID, in.ID))
		})
	writeTool(set, "files_save", "Save a new file to the user's Nuage storage, from text content or from a file the user attached.",
		func(ctx context.Context, in filesSaveInput) (string, error) {
			file, err := saveFile(ctx, in, attachment)
			if err != nil {
				return "", err
			}
			return jsonResult(n.Save(ctx, set.userID, in.Folder, file))
		})
}

func saveFile(ctx context.Context, in filesSaveInput, attachment attachmentReader) (File, error) {
	if (in.Content == "") == (in.AttachmentID == "") {
		return File{}, errors.New("pass exactly one of content or attachment_id")
	}
	file := File{Name: in.Name, MediaType: "text/plain; charset=utf-8", Data: []byte(in.Content)}
	if in.AttachmentID != "" {
		if attachment == nil {
			return File{}, errors.New("attachments are not available here")
		}
		found, err := attachment(ctx, in.AttachmentID)
		if err != nil {
			return File{}, fmt.Errorf("no attachment %q in this chat", in.AttachmentID)
		}
		file = found
		if in.Name != "" {
			file.Name = in.Name
		}
	}
	if len(in.Content) > maxSavedText {
		return File{}, fmt.Errorf("content is over %d KB", maxSavedText>>10)
	}
	return file, fileName(file.Name)
}

func fileName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed != name || strings.ContainsAny(name, "/\\\x00") || name == "." || name == ".." {
		return fmt.Errorf("name %q must be a plain file name such as notes.md", name)
	}
	return nil
}
