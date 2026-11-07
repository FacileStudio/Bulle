package agent

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/core/settings"
)

// runStats is what one headless run measured: the run's total cost and token
// counts, the tool calls made along the way, and the final conversation size.
type runStats struct {
	nacelle.Usage
	ToolCalls          int
	FinalContextTokens int64
}

// UsageError marks a malformed invocation or config the user has to fix.
type UsageError struct {
	err error
}

func (e *UsageError) Error() string { return e.err.Error() }

func usagef(format string, args ...any) error {
	return &UsageError{err: fmt.Errorf(format, args...)}
}

// findCommand reports whether the first non-flag argument is name, and splits
// the arguments around it.
func findCommand(name string) (prefix, sub []string, ok bool) {
	args := os.Args[1:]
	first := 0
	for first < len(args) && strings.HasPrefix(args[first], "-") {
		first++
	}
	if first >= len(args) || args[first] != name {
		return nil, nil, false
	}
	return args[:first], args[first+1:], true
}

var configOnce = sync.OnceValues(func() (settings.Config, error) {
	ensureUserPath()
	flags := settings.FromFlags(settings.Defaults(""))
	return settings.Settings(DefaultSystemPrompt(), flags)
})

func loadConfig() (settings.Config, error) {
	return configOnce()
}
