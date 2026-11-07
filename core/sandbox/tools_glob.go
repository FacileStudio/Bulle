package sandbox

import (
	"path/filepath"
	"strings"
)

func matchStar(pattern, name []string) bool {
	if len(pattern) == 0 {
		return true
	}
	for i := 0; i <= len(name); i++ {
		if matchSegments(pattern, name[i:], matchStar) {
			return true
		}
	}
	return false
}

func matchSegments(pattern, name []string, star func([]string, []string) bool) bool {
	p := pattern
	n := name
	for len(p) > 0 {
		if p[0] == "**" {
			return star(p[1:], n)
		}
		if len(n) == 0 {
			return false
		}
		if ok, err := filepath.Match(p[0], n[0]); err != nil || !ok {
			return false
		}
		p = p[1:]
		n = n[1:]
	}
	return len(n) == 0
}

func matchGlob(pattern, name string) bool {
	p := strings.Split(pattern, "/")
	n := strings.Split(name, "/")
	return matchSegments(p, n, matchStar)
}

const prunedDirs = "-name . -o -type d \\( -name '.*' -o -name node_modules -o -name vendor -o -name dist -o -name build -o -name target -o -name .svelte-kit -o -name .next -o -name __pycache__ -o -name .venv -o -name coverage \\) -prune -o -type f"

// findCmd is the remote find, run from dir, over every file under start
// with generated directories left out. An empty start is dir itself.
func findCmd(dir, start, action string) string {
	root := "."
	if start = strings.TrimSpace(start); start != "" {
		root = quoteArg(start)
	}
	return workdirPrefix(dir) + "find " + root + " " + prunedDirs + " " + action
}

// searchDir is the directory a glob runs from: the one the call named, or
// the working directory.
func searchDir(workDir, dir string) string {
	if strings.TrimSpace(dir) == "" {
		return workDir
	}
	return resolveRemotePath(workDir, dir)
}
