// Package myceliumtest is an in-process stand-in for Mycelium's MCP server
// and the Registre token exchange in front of it, for Bulle's tests. It
// serves the tools Bulle calls, keeps each person's spaces, pages and
// artifacts in memory, and confirms deletes through MRTR as Mycelium does.
package myceliumtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Artifact is one stored artifact.
type Artifact struct {
	ID      string
	Title   string
	Content string
	Created time.Time
}

// Call is one tool call the server answered.
type Call struct {
	User string
	Tool string
	Args map[string]any
}

// Server is the fake. Users are named by the Bulle user id their token was
// exchanged for, such as "1".
type Server struct {
	srv *httptest.Server

	mu        sync.Mutex
	owners    map[string]string
	names     map[string]string
	admins    map[string]bool
	pages     map[string]map[string]string
	artifacts map[string][]Artifact
	calls     []Call
	refused   map[string]bool
	seq       int
}

// New starts the fake.
func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{
		owners: map[string]string{}, names: map[string]string{}, admins: map[string]bool{},
		pages: map[string]map[string]string{}, artifacts: map[string][]Artifact{}, refused: map[string]bool{},
	}
	server := sdk.NewServer(&sdk.Implementation{Name: "mycelium", Version: "test"}, &sdk.ServerOptions{
		SetCacheable: func(_ context.Context, _ sdk.Request, c *sdk.Cacheable) { c.CacheScope = "private"; c.TTLMs = 60000 },
	})
	s.register(server)
	mcp := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return server }, &sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(rw http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(rw).Encode(map[string]string{"token_endpoint": s.srv.URL + "/token"})
	})
	mux.HandleFunc("/token", s.exchange)
	mux.Handle("/mcp", s.bearer(mcp))
	s.srv = httptest.NewServer(mux)
	t.Cleanup(s.srv.Close)
	return s
}

// Issuer is the fake Registre's issuer URL.
func (s *Server) Issuer() string { return s.srv.URL }

// Resource is the MCP endpoint, which is also its resource URI.
func (s *Server) Resource() string { return s.srv.URL + "/mcp" }

// ListByUser answers one identity per Bulle user, whose access token names
// the user, so the exchanged token does too.
func (s *Server) ListByUser(_ context.Context, userID int64) ([]porte.StoredIdentity, error) {
	return []porte.StoredIdentity{{UserID: userID, Provider: s.srv.URL, Tokens: porte.TokenSet{
		AccessToken: fmt.Sprint("user-", userID), Expiry: time.Now().Add(time.Hour),
	}}}, nil
}

// Save keeps nothing: the fake's tokens never refresh.
func (s *Server) Save(context.Context, porte.StoredIdentity) error { return nil }

// Refuse makes Mycelium answer 401 to user, as for a person it cannot know.
func (s *Server) Refuse(user string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refused[user] = true
}

// AddSpace gives user a space and returns its id.
func (s *Server) AddSpace(user, name string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addSpace(user, name)
}

func (s *Server) addSpace(user, name string) string {
	s.seq++
	id := fmt.Sprintf("sp-%d", s.seq)
	s.owners[id], s.names[id] = user, name
	return id
}

// AddSpaceID gives user a space with a chosen id.
func (s *Server) AddSpaceID(user, id, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owners[id], s.names[id] = user, name
}

// MakeAdmin lets user reach the common tree, space id "".
func (s *Server) MakeAdmin(user string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.admins[user] = true
}

// SetPage stores a page as is.
func (s *Server) SetPage(space, path, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pages[space] == nil {
		s.pages[space] = map[string]string{}
	}
	s.pages[space][path] = content
}

// Page answers a stored page.
func (s *Server) Page(space, path string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	content, ok := s.pages[space][path]
	return content, ok
}

// AddArtifact stores an artifact in a tree, "" being the common tree, and
// returns its id. Each one is a day newer than the one before.
func (s *Server) AddArtifact(space, title, content string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	id := fmt.Sprintf("2026-11-06-art-%d", s.seq)
	created := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC).AddDate(0, 0, s.seq)
	s.artifacts[space] = append(s.artifacts[space], Artifact{ID: id, Title: title, Content: content, Created: created})
	return id
}

// Artifacts answers a tree's artifacts.
func (s *Server) Artifacts(space string) []Artifact {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Artifact(nil), s.artifacts[space]...)
}

// Spaces answers the spaces user owns, by id.
func (s *Server) Spaces(user string) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]string{}
	for id, owner := range s.owners {
		if owner == user {
			out[id] = s.names[id]
		}
	}
	return out
}

// Calls answers every tool call so far.
func (s *Server) Calls() []Call {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Call(nil), s.calls...)
}

func (s *Server) exchange(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:token-exchange" || r.Form.Get("audience") != s.Resource() {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(map[string]string{"error": "invalid_target"})
		return
	}
	_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "mcp-for-" + r.Form.Get("subject_token"), "expires_in": 600})
}

type userKey struct{}

func (s *Server) bearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer mcp-for-user-")
		s.mu.Lock()
		refused := s.refused[user]
		s.mu.Unlock()
		if !ok || refused {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

type args = map[string]any

type handler func(user string, in args, req *sdk.CallToolRequest) (*sdk.CallToolResult, args, error)

func (s *Server) register(server *sdk.Server) {
	tools := map[string]handler{
		"mycelium_read_memory_index":  s.index,
		"mycelium_search_memory":      s.search,
		"mycelium_read_memory_page":   s.readPage,
		"mycelium_list_memory_pages":  s.listPages,
		"mycelium_write_memory_page":  s.writePage,
		"mycelium_delete_memory_page": s.deletePage,
		"mycelium_create_space":       s.createSpace,
		"mycelium_list_artifacts":     s.listArtifacts,
		"mycelium_search_artifacts":   s.searchArtifacts,
		"mycelium_read_artifact":      s.readArtifact,
		"mycelium_publish_artifact":   s.publish,
		"mycelium_update_artifact":    s.update,
		"mycelium_delete_artifact":    s.deleteArtifact,
	}
	for name, run := range tools {
		readOnly := strings.Contains(name, "_read_") || strings.Contains(name, "_list_") || strings.Contains(name, "_search_")
		sdk.AddTool(server, &sdk.Tool{Name: name, Annotations: &sdk.ToolAnnotations{ReadOnlyHint: readOnly}},
			func(ctx context.Context, req *sdk.CallToolRequest, in args) (*sdk.CallToolResult, args, error) {
				user, _ := ctx.Value(userKey{}).(string)
				if user == "" && req.Extra != nil && req.Extra.Header != nil {
					user, _ = strings.CutPrefix(req.Extra.Header.Get("Authorization"), "Bearer mcp-for-user-")
				}
				s.mu.Lock()
				s.calls = append(s.calls, Call{User: user, Tool: name, Args: in})
				s.mu.Unlock()
				return run(user, in, req)
			})
	}
}

// tree checks user reaches space, "" being the common tree of admins.
func (s *Server) tree(user string, in args) (string, error) {
	space, _ := in["space_id"].(string)
	if space == "" {
		if !s.admins[user] {
			return "", errors.New("you have no personal tree on this server, only spaces. Pass a space_id")
		}
		return "", nil
	}
	if s.owners[space] != user {
		return "", errors.New("no space with this space_id is yours")
	}
	return space, nil
}

func text(in args, key string) string {
	value, _ := in[key].(string)
	return value
}

func (s *Server) index(user string, _ args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spaces := []args{}
	if s.admins[user] {
		spaces = append(spaces, args{"space_id": "", "name": "Personal", "role": "owner"})
	}
	ids := []string{}
	for id, owner := range s.owners {
		if owner == user {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		spaces = append(spaces, args{"space_id": id, "name": s.names[id], "role": "owner"})
	}
	return nil, args{"spaces": spaces, "space_id": "", "index": ""}, nil
}

func (s *Server) search(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	results := []args{}
	for path, content := range s.pages[space] {
		if strings.Contains(strings.ToLower(content), strings.ToLower(text(in, "query"))) {
			results = append(results, args{"space_id": space, "path": path, "excerpt": content, "line": 1, "score": 1.0})
		}
	}
	return nil, args{"results": results, "degraded": true}, nil
}

func (s *Server) readPage(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	content, ok := s.pages[space][text(in, "path")]
	if !ok {
		return nil, nil, fmt.Errorf("there is no page %s", text(in, "path"))
	}
	return nil, args{"space_id": space, "path": text(in, "path"), "content": content}, nil
}

func (s *Server) listPages(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	pages := []args{}
	for path, content := range s.pages[space] {
		pages = append(pages, args{"path": path, "size": len(content), "updated": "2026-11-06T09:00:00Z"})
	}
	return nil, args{"space_id": space, "pages": pages, "total": len(pages)}, nil
}

func (s *Server) writePage(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	content, path := text(in, "content"), text(in, "path")
	if !strings.HasPrefix(content, "---\n") || text(in, "log") == "" {
		return nil, nil, errors.New("the page needs YAML frontmatter and a log line")
	}
	if s.pages[space] == nil {
		s.pages[space] = map[string]string{}
	}
	_, existed := s.pages[space][path]
	s.pages[space][path] = content
	return nil, args{"space_id": space, "path": path, "created": !existed}, nil
}

// confirmed asks first like Mycelium: the first call returns input_required,
// and only an accept that keeps the box ticked goes on.
func confirmed(req *sdk.CallToolRequest, question string) (*sdk.CallToolResult, bool) {
	if req.Params.RequestState == "" {
		form := map[string]any{"type": "object", "properties": map[string]any{"confirm": map[string]any{"type": "boolean", "default": true}}, "required": []string{"confirm"}}
		ask := &sdk.ElicitParams{Mode: "form", Message: question, RequestedSchema: form}
		return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": ask}, RequestState: "state"}, false
	}
	answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
	return nil, answer != nil && answer.Action == "accept" && answer.Content["confirm"] != false
}

func (s *Server) deletePage(user string, in args, req *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	space, err := s.tree(user, in)
	_, found := s.pages[space][text(in, "path")]
	s.mu.Unlock()
	if err != nil {
		return nil, nil, err
	}
	if !found {
		return nil, nil, fmt.Errorf("there is no page %s", text(in, "path"))
	}
	ask, yes := confirmed(req, "Delete the memory page "+text(in, "path")+" for good?")
	if ask != nil {
		return ask, nil, nil
	}
	if !yes {
		return nil, args{"status": "declined", "message": "The user declined."}, nil
	}
	s.mu.Lock()
	delete(s.pages[space], text(in, "path"))
	s.mu.Unlock()
	return nil, args{"status": "deleted", "space_id": space, "path": text(in, "path"), "message": "deleted"}, nil
}

func (s *Server) createSpace(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, owner := range s.owners {
		if owner == user && strings.EqualFold(s.names[id], text(in, "name")) {
			return nil, args{"space_id": id, "name": s.names[id], "role": "owner", "created": false}, nil
		}
	}
	id := s.addSpace(user, text(in, "name"))
	return nil, args{"space_id": id, "name": text(in, "name"), "role": "owner", "created": true}, nil
}

func (s *Server) summary(space string, a Artifact) args {
	name := "Personal"
	if space != "" {
		name = s.names[space]
	}
	stamp := a.Created.UTC().Format(time.RFC3339)
	return args{"id": a.ID, "title": a.Title, "space_id": space, "space_name": name, "created": stamp, "updated": stamp}
}

func (s *Server) listArtifacts(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	trees := []string{}
	if _, given := in["space_id"]; given {
		space, err := s.tree(user, in)
		if err != nil {
			return nil, nil, err
		}
		trees = append(trees, space)
	} else {
		trees = s.readable(user)
	}
	out := []args{}
	for _, space := range trees {
		for _, a := range s.artifacts[space] {
			out = append(out, s.summary(space, a))
		}
	}
	return nil, args{"artifacts": out, "total": len(out)}, nil
}

// readable answers the trees user can read: the common tree for an admin,
// then each owned space.
func (s *Server) readable(user string) []string {
	trees := []string{}
	if s.admins[user] {
		trees = append(trees, "")
	}
	for id, owner := range s.owners {
		if owner == user {
			trees = append(trees, id)
		}
	}
	sort.Strings(trees)
	return trees
}

// searchArtifacts matches like Mycelium: every word in the title or the
// body, title matches first, pages of SearchPage with the offset as cursor.
func (s *Server) searchArtifacts(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	words := strings.Fields(strings.ToLower(text(in, "query")))
	if len(words) == 0 {
		return nil, nil, errors.New("query must hold 1 to 200 characters")
	}
	offset := 0
	if cursor := text(in, "cursor"); cursor != "" {
		if _, err := fmt.Sscanf(cursor, "at-%d", &offset); err != nil || offset < 0 {
			return nil, nil, errors.New("invalid cursor")
		}
	}
	hits := s.matches(user, words)
	out := args{"results": hits[min(offset, len(hits)):], "total": len(hits)}
	if len(hits)-offset > SearchPage {
		out["results"] = hits[offset : offset+SearchPage]
		out["next_cursor"] = fmt.Sprintf("at-%d", offset+SearchPage)
	}
	return nil, out, nil
}

// matches answers the user's artifacts holding every word, title matches
// first, each with its body as the snippet.
func (s *Server) matches(user string, words []string) []args {
	titles, bodies := []args{}, []args{}
	for _, space := range s.readable(user) {
		for _, a := range s.artifacts[space] {
			hit := s.summary(space, a)
			hit["snippet"] = a.Content
			switch {
			case holdsAll(a.Title, words):
				hit["match"] = "title"
				titles = append(titles, hit)
			case holdsAll(a.Title+" "+a.Content, words):
				hit["match"] = "content"
				bodies = append(bodies, hit)
			}
		}
	}
	return append(titles, bodies...)
}

// SearchPage is how many hits one search answer holds, as in Mycelium.
const SearchPage = 20

func holdsAll(text string, words []string) bool {
	lower := strings.ToLower(text)
	for _, word := range words {
		if !strings.Contains(lower, word) {
			return false
		}
	}
	return true
}

func (s *Server) find(space, id string) int {
	for i, a := range s.artifacts[space] {
		if a.ID == id {
			return i
		}
	}
	return -1
}

func (s *Server) readArtifact(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	i := s.find(space, text(in, "id"))
	if i < 0 {
		return nil, nil, fmt.Errorf("no artifact %s", text(in, "id"))
	}
	a := s.artifacts[space][i]
	out := s.summary(space, a)
	out["content"] = "---\ntitle: \"" + a.Title + "\"\ntype: artifact\n---\n\n" + a.Content
	return nil, out, nil
}

func (s *Server) publish(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	for i, a := range s.artifacts[space] {
		if strings.EqualFold(a.Title, text(in, "title")) {
			s.artifacts[space][i].Content = text(in, "content")
			out := s.summary(space, s.artifacts[space][i])
			out["replaced"] = true
			return nil, out, nil
		}
	}
	s.seq++
	a := Artifact{ID: fmt.Sprintf("2026-11-06-art-%d", s.seq), Title: text(in, "title"), Content: text(in, "content"),
		Created: time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)}
	s.artifacts[space] = append(s.artifacts[space], a)
	return nil, s.summary(space, a), nil
}

func (s *Server) update(user string, in args, _ *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	space, err := s.tree(user, in)
	if err != nil {
		return nil, nil, err
	}
	i := s.find(space, text(in, "id"))
	if i < 0 {
		return nil, nil, fmt.Errorf("no artifact %s", text(in, "id"))
	}
	if title := text(in, "title"); title != "" {
		s.artifacts[space][i].Title = title
	}
	s.artifacts[space][i].Content = text(in, "content")
	return nil, s.summary(space, s.artifacts[space][i]), nil
}

func (s *Server) deleteArtifact(user string, in args, req *sdk.CallToolRequest) (*sdk.CallToolResult, args, error) {
	s.mu.Lock()
	space, err := s.tree(user, in)
	i := s.find(space, text(in, "id"))
	s.mu.Unlock()
	if err != nil {
		return nil, nil, err
	}
	if i < 0 {
		return nil, nil, fmt.Errorf("no artifact %s", text(in, "id"))
	}
	ask, yes := confirmed(req, "Delete the artifact "+text(in, "id")+" for good?")
	if ask != nil {
		return ask, nil, nil
	}
	if !yes {
		return nil, args{"status": "declined", "message": "The user declined."}, nil
	}
	s.mu.Lock()
	s.artifacts[space] = append(s.artifacts[space][:i], s.artifacts[space][i+1:]...)
	s.mu.Unlock()
	return nil, args{"status": "deleted", "space_id": space, "id": text(in, "id"), "message": "deleted"}, nil
}
