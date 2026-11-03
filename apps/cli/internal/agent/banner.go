package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/anthropic"

	"github.com/FacileStudio/Bulle/core/settings"
)

// banner is the two lines the transcript opens with: the build and the
// backend that answers, then the root and the skill count. The model and the
// assistant's name live in the prompt bar, so the banner does not repeat them.
// MCP earns words only when a server was configured.
func banner(backend nacelle.Backend, config settings.Config, found loaded, mcp connected, version string) string {
	line := fmt.Sprintf("bulle %s · %s\n%s · %s", version, backend.Name(),
		homeTilde(absolute(config.Root)), countedNoun(len(found.skills), "skill"))
	if mcp.servers > 0 {
		line += " · " + countedNoun(mcp.servers, "MCP server")
	}
	return line
}

// homeTilde shortens a path under the home directory to ~/..., the way a
// Linux or macOS shell prints it.
func homeTilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || runtime.GOOS == "windows" {
		return path
	}
	if path == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return "~/" + rest
	}
	return path
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
