package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestInstallBrowserCopiesOnlyWhenTheBundleDiffers(t *testing.T) {
	bundle := []byte("console.log(1)")
	digest := sha256.Sum256(bundle)
	var remote []string
	answer := "stale  /home/boite/.bulle/bin/bulle-browser.mjs"
	runner := &mockRunner{runFunc: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		remote = append(remote, args[len(args)-1])
		return []byte(answer), nil
	}}
	target := &Target{Host: "127.0.0.1", Port: 2226, User: "boite"}
	copied, err := InstallBrowser(context.Background(), target, runner, bundle)
	if err != nil || !copied || len(remote) != 2 || !strings.Contains(remote[1], "base64 -d") {
		t.Fatalf("copied %v err %v remote %q", copied, err, remote)
	}
	answer = hex.EncodeToString(digest[:]) + "  /home/boite/.bulle/bin/bulle-browser.mjs"
	remote = nil
	copied, err = InstallBrowser(context.Background(), target, runner, bundle)
	if err != nil || copied || len(remote) != 1 {
		t.Fatalf("copied %v err %v remote %q", copied, err, remote)
	}
}
