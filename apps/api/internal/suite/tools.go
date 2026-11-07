package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const maxToolOutput = 24 << 10

type toolSet struct {
	userID   int64
	defaults Defaults
	seeMail  func(ctx context.Context, emailIDs []int64) (map[int64]time.Time, error)
	tools    []nacelle.Tool
	names    []string
	writes   []string
	err      error
}

func (s *toolSet) add(tool nacelle.Tool, err error) {
	if err != nil && s.err == nil {
		s.err = err
	}
	s.tools = append(s.tools, Aliased(tool))
	s.names = append(s.names, tool.Name())
}

func readTool[In any](set *toolSet, name, description string, run func(ctx context.Context, in In) (string, error)) {
	tool, err := nacelle.NewToolWithOptions(name, description, run, nacelle.ToolOptions{ReadOnly: true})
	set.add(tool, err)
}

func WriteTools() []string {
	return []string{
		"courrier_send", "courrier_save_draft", "courrier_delete_draft", "courrier_update", "courrier_move",
		"agenda_create_event", "agenda_update_event", "agenda_delete_event",
		"opus_create_task", "opus_update_task", "opus_delete_task",
		"opus_create_project", "opus_update_project", "opus_delete_project",
		"sablier_start", "sablier_stop",
		"nuage_save",
		"glouton_create_company", "glouton_update_company", "glouton_create_person", "glouton_update_person",
	}
}

// MemorySpace answers the Mycelium space this conversation's memory lives in.
type MemorySpace func(ctx context.Context) (string, error)

// File is an attachment the user sent, as nuage_save reads it.
type File struct {
	Name      string
	MediaType string
	Data      []byte
}

// Runtime is what the runner lends the tools for one conversation: the
// user's personal memory space, used when the Bulle space names no Mycelium
// space of its own, a reader for the user's attachments, the user's Bulle
// spaces, and the id of the space this conversation is in. Any may be empty.
type Runtime struct {
	PersonalSpace MemorySpace
	Attachment    func(ctx context.Context, id string) (File, error)
	Spaces        func(ctx context.Context) ([]store.Space, error)
	CurrentSpace  string
	SeeMail       func(ctx context.Context, emailIDs []int64) (map[int64]time.Time, error)
}

func Tools(gw *Gateway, userID int64, defaults Defaults, rt Runtime) ([]nacelle.Tool, string, error) {
	set := &toolSet{userID: userID, defaults: defaults, seeMail: rt.SeeMail}
	cfg := gw.Config()
	for _, app := range cfg.Enabled() {
		switch app {
		case AppCourrier:
			courrierTools(set, NewCourrier(gw))
		case AppAgenda:
			agendaTools(set, NewAgenda(gw))
		case AppOpus:
			opusTools(set, NewOpus(gw))
		case AppSablier:
			sablierTools(set, NewSablier(gw))
		case AppMycelium:
			memoryTools(set, NewMycelium(gw), defaults.memorySpace(NewMycelium(gw), userID, rt.PersonalSpace))
			artifactTools(set, NewMycelium(gw), defaults.memorySpace(NewMycelium(gw), userID, rt.PersonalSpace), rt.CurrentSpace)
		case AppNuage:
			nuageTools(set, NewNuage(gw), rt.Attachment)
		case AppGlouton:
			gloutonTools(set, NewGlouton(gw))
		}
	}
	if rt.Spaces != nil {
		spacesTools(set, rt)
	}
	if set.err != nil {
		return nil, "", set.err
	}
	note := Prompt(set.names, set.writes)
	if cfg.Apps[AppMycelium].Enabled() {
		note += "\n\n" + memoryNote + "\n\n" + artifactNote
	} else {
		note = strings.TrimSpace(note + "\n\n" + NoMemoryNote)
	}
	return set.tools, note, nil
}

func Prompt(names, writes []string) string {
	if len(names) == 0 {
		return ""
	}
	note := "You can work with the signed-in user's own Facile data (mail, calendar, tasks, timers, files, memory, prospects) through " +
		strings.Join(names, ", ") + ". These tools only ever see this user's data. " +
		"When one says the account needs reconnecting, tell the user to sign out of Bulle and sign in again."
	if len(writes) == 0 {
		return note
	}
	return note + " The tools " + strings.Join(writes, ", ") + " change the user's data. " +
		"Bulle shows the user an approval card before each one runs, so call them directly instead of asking for confirmation in chat; " +
		"if the user denies one, do not retry it. Read the item first when you need its id or current state, and pass only the fields that change." +
		mailRules(names)
}

func mailRules(names []string) string {
	if !slices.Contains(names, "courrier_send") {
		return ""
	}
	return mailNote
}

const mailNote = " For mail: search every mailbox before asking the user anything; if a full name finds nothing, try the surname, the first name and the company. " +
	"An address the user calls theirs is one of their mailboxes, listed by courrier_list_accounts. " +
	"To send, reply or follow up, call courrier_send with reply_to and account taken from the email you answer; the approval card lets the user edit it first, so never stop at a draft or ask permission in chat. " +
	"Write the email in the language of the thread. " +
	"Once the user has dealt with an email, or asks you to, mark it read with courrier_mark; it needs no approval."

func jsonResult(value any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return clip(string(data)), nil
}

func rawResult(raw []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return clip(string(raw)), nil
}

func clip(text string) string {
	if len(text) <= maxToolOutput {
		return text
	}
	return fmt.Sprintf("%s\n[output truncated at %d bytes, narrow the query]", text[:maxToolOutput], maxToolOutput)
}
