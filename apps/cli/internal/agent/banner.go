package agent

import (
	"fmt"
	"path/filepath"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/anthropic"

	"github.com/FacileStudio/Bulle/core/settings"
)

// banner is the two lines the transcript opens with: the build and the
// signed-in assistant's name, then the backend and model that answer, and the
// root, skill and context-file counts that say what this session loaded.
// MCP earns words only when a server was configured.
func banner(backend nacelle.Backend, config settings.Config, found loaded, mcp connected, head string) string {
	line := fmt.Sprintf("%s · %s · %s\n%s · %s · %s", head, backend.Name(), resolvedModel(backend, config),
		absolute(config.Root), countedNoun(len(found.skills), "skill"), countedNoun(found.contextFiles, "context file"))
	if mcp.servers > 0 {
		line += " · " + countedNoun(mcp.servers, "MCP server")
	}
	return line
}

// title is the banner's opening: the build, then the assistant's name when
// the session is signed in.
func title(version, assistant string) string {
	if assistant == "" {
		return "bulle " + version
	}
	return "bulle " + version + " · " + assistant
}

// resolvedModel names the model this session runs on: the configured name,
// else the backend's own default, else anthropic's. An empty model in the
// config is the ordinary case — the provider fills its own default — and the
// banner and the status line both have to name the one that actually answers,
// so neither reads the raw setting.
func resolvedModel(backend nacelle.Backend, config settings.Config) string {
	if config.Model != "" {
		return config.Model
	}
	if withModel, ok := backend.(interface{ Model() string }); ok && withModel.Model() != "" {
		return withModel.Model()
	}
	return anthropic.DefaultModel
}

// countedNoun is "N noun" or "N nouns" — the one piece of English this
// client bothers to pluralize, because the banner reads it on every launch.
func countedNoun(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// absolute is the root as a path that means the same thing from anywhere.
// "-root ." reads identically from every directory bulle was launched in
// and so answers nothing on its own — neither for the person reading the
// banner nor for the model reading the system prompt, which are the two
// callers.
//
// A target-labeled root ("target:<name>", set by remote and sandbox
// sessions with no configured workdir) is echoed back as-is: it names the
// machine on the other end of the SSH connection, not a directory on this
// one, and resolving it would only rename the directory bulle was launched
// from.
func absolute(root string) string {
	if settings.IsTargetRoot(root) {
		return root
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}
