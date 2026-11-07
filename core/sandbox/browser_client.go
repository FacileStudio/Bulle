package sandbox

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FacileStudio/nacelle"
)

// BrowserBundle is where the bulle-browser daemon bundle lives in the VM,
// relative to the remote user's home.
const BrowserBundle = ".bulle/bin/bulle-browser.mjs"

// HeldRequest is the daemon's own record of a request it is holding for the
// owner. It is what an approval card shows, never what the model wrote.
type HeldRequest struct {
	ID           string `json:"id"`
	Method       string `json:"method"`
	URL          string `json:"url"`
	Origin       string `json:"origin"`
	BodyPreview  string `json:"body_preview"`
	BodyBytes    int    `json:"body_bytes"`
	ResourceType string `json:"resource_type"`
}

// BrowserReply is one answer from the bulle-browser CLI.
type BrowserReply struct {
	OK    bool          `json:"ok"`
	Error string        `json:"error"`
	Text  string        `json:"text"`
	Image *browserImage `json:"image"`
	Held  []HeldRequest `json:"held"`
	URL   string        `json:"url"`
}

type browserImage struct {
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// Browser runs bulle-browser commands in a VM over an SSH session.
type Browser interface {
	Browser(ctx context.Context, cmd string, args any) (BrowserReply, error)
}

// Browser runs one bulle-browser command in the VM and decodes its answer.
// A reply the daemon marked as failed comes back as an error.
func (s *remoteSession) Browser(ctx context.Context, cmd string, args any) (BrowserReply, error) {
	encoded, err := json.Marshal(args)
	if err != nil {
		return BrowserReply{}, err
	}
	remote := fmt.Sprintf(`node "$HOME"/%s %s %s`, BrowserBundle, shellQuote(cmd), shellQuote(string(encoded)))
	out, runErr := s.runSSH(ctx, remote)
	reply, err := decodeBrowserReply(out)
	if err != nil && runErr != nil {
		return BrowserReply{}, fmt.Errorf("bulle-browser %s: %w (%w)", cmd, err, runErr)
	}
	if err != nil {
		return BrowserReply{}, fmt.Errorf("bulle-browser %s: %w", cmd, err)
	}
	if !reply.OK {
		return reply, fmt.Errorf("%s", reply.Error)
	}
	return reply, nil
}

// decodeBrowserReply reads the last JSON line the CLI printed, so a stray
// warning above it does not hide the answer.
func decodeBrowserReply(out []byte) (BrowserReply, error) {
	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	last := lines[len(lines)-1]
	var reply BrowserReply
	if err := json.Unmarshal(last, &reply); err != nil {
		return BrowserReply{}, fmt.Errorf("unreadable answer %q", strings.TrimSpace(string(out)))
	}
	return reply, nil
}

// images decodes the reply's screenshot, if it carries one.
func (r BrowserReply) images() ([]nacelle.Image, error) {
	if r.Image == nil {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(r.Image.Data)
	if err != nil {
		return nil, fmt.Errorf("decoding the screenshot: %w", err)
	}
	return []nacelle.Image{{MediaType: r.Image.MediaType, Data: data}}, nil
}

// Find returns the request the daemon is holding under id.
func (r BrowserReply) Find(id string) (HeldRequest, bool) {
	for _, held := range r.Held {
		if held.ID == id {
			return held, true
		}
	}
	return HeldRequest{}, false
}
