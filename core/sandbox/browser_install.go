package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// InstallBrowser copies the bulle-browser bundle into the VM when the copy
// there differs from bundle, and stops a daemon still running the old one so
// the next command starts the new code. A matching copy costs one SSH call.
func InstallBrowser(ctx context.Context, target *Target, runner Runner, bundle []byte) (bool, error) {
	digest := sha256.Sum256(bundle)
	sum := hex.EncodeToString(digest[:])
	check := fmt.Sprintf(`sha256sum "$HOME"/%s 2>/dev/null`, BrowserBundle)
	out, _ := runner.Run(ctx, "ssh", buildRemoteSSHArgs(target, target.User, check, "")...)
	if strings.HasPrefix(strings.TrimSpace(string(out)), sum) {
		return false, nil
	}
	path := `"$HOME"/` + BrowserBundle
	install := fmt.Sprintf(
		`mkdir -p "$HOME"/.bulle/bin && printf '%%s' %s | base64 -d > %s.tmp && mv %s.tmp %s && (node %s shutdown >/dev/null 2>&1; true)`,
		shellQuote(base64.StdEncoding.EncodeToString(bundle)), path, path, path, path,
	)
	out, err := runner.Run(ctx, "ssh", buildRemoteSSHArgs(target, target.User, install, "")...)
	if err != nil {
		return false, fmt.Errorf("installing bulle-browser: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return true, nil
}
