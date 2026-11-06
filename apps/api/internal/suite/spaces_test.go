package suite

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func spaceListRuntime(spaces []store.Space) Runtime {
	return Runtime{Spaces: func(context.Context) ([]store.Space, error) { return spaces, nil }, CurrentSpace: "sp2"}
}

func TestSpacesListIsRegisteredAndReadOnly(t *testing.T) {
	w := newWorld(t)
	tools, note, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, spaceListRuntime(nil))
	if err != nil {
		t.Fatal(err)
	}
	var registered nacelle.Tool
	for _, tool := range tools {
		if tool.Name() == "spaces_list" {
			registered = tool
		}
	}
	if registered == nil {
		t.Fatal("spaces_list is not registered")
	}
	readOnly, ok := registered.(nacelle.ReadOnlyTool)
	if !ok || !readOnly.IsReadOnly() {
		t.Error("spaces_list must be read-only")
	}
	if !strings.Contains(note, "spaces_list") {
		t.Errorf("prompt must list spaces_list: %q", note)
	}
}

func TestSpacesListReturnsExactlyTheClosuresSpaces(t *testing.T) {
	w := newWorld(t)
	spaces := []store.Space{
		{ID: "sp1", Name: "Work", Icon: "briefcase", MyceliumSpaceID: "my1", OpusWorkspace: "op1", AgendaCalendar: 7, CourrierAccount: 3},
		{ID: "sp2", Name: "Acme", Icon: "handshake", MyceliumSpaceID: "my2", OpusWorkspace: "op2", AgendaCalendar: 9, CourrierAccount: 5, RegistreSpaceID: "rg2"},
	}
	defaults := Defaults{MyceliumSpace: "my2", OpusWorkspace: "op2", AgendaCalendar: 9, CourrierAccount: 5, SharedSpace: "Acme"}
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, defaults, spaceListRuntime(spaces))
	if err != nil {
		t.Fatal(err)
	}
	out, err := runTool(t, tools, "spaces_list", "{}")
	if err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("spaces_list %q: %v", out, err)
	}
	want := []map[string]any{
		{"id": "sp1", "name": "Work", "icon": "briefcase", "shared": false, "mycelium_space_id": "my1",
			"opus_workspace": "op1", "agenda_calendar": float64(7), "courrier_account": float64(3), "current": false},
		{"id": "sp2", "name": "Acme", "icon": "handshake", "shared": true, "mycelium_space_id": "my2",
			"opus_workspace": "op2", "agenda_calendar": float64(9), "courrier_account": float64(5), "current": true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("spaces_list:\n got %v\nwant %v", got, want)
	}
	failing, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{
		Spaces: func(context.Context) ([]store.Space, error) { return nil, errors.New("list broke") },
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runTool(t, failing, "spaces_list", "{}"); err == nil || !strings.Contains(err.Error(), "list broke") {
		t.Errorf("want the closure error, got %v", err)
	}
}

func TestSpacesListMarksOnlyTheSessionSpaceWhenPointersMatch(t *testing.T) {
	set := &toolSet{}
	spacesTools(set, spaceListRuntime([]store.Space{{ID: "sp1", Name: "Home"}, {ID: "sp2", Name: "Side"}}))
	out, err := runTool(t, set.tools, "spaces_list", "{}")
	if err != nil {
		t.Fatal(err)
	}
	var got []spaceListEntry
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Current || !got[1].Current {
		t.Errorf("only sp2 is current: %+v", got)
	}
}

func TestSpacesListWithoutSpacesRuntimeYieldsClearError(t *testing.T) {
	set := &toolSet{}
	spacesTools(set, Runtime{})
	out, err := runTool(t, set.tools, "spaces_list", "{}")
	if err == nil {
		t.Fatalf("want an error, got %q", out)
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("want a clear unavailable error, got %v", err)
	}
}
