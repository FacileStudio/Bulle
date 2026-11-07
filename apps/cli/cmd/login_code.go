package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/FacileStudio/porte/loopback"
)

// loginCode gets a one-time login code. A local browser hands it back through
// the loopback listener. Any other browser, such as the laptop in front of an
// SSH session, signs in through an address without a port, where porte shows
// the code to paste here instead of redirecting to 127.0.0.1.
func loginCode(ctx context.Context, out io.Writer, in io.Reader, server string, pasted bool) (string, error) {
	if !pasted {
		code, opened, err := browserCode(ctx, out, in, server)
		if err != nil || opened {
			return code, err
		}
	}
	return pastedCode(out, in, server)
}

// browserCode opens a local browser on the login page and waits for the code
// on the loopback listener. It reports false when no browser could be opened.
func browserCode(ctx context.Context, out io.Writer, in io.Reader, server string) (string, bool, error) {
	state, err := loopback.RandomState()
	if err != nil {
		return "", false, err
	}
	listener, err := loopback.Listen()
	if err != nil {
		return "", false, fmt.Errorf("opening the loopback listener: %w", err)
	}
	defer func() { _ = listener.Close() }()
	listener.AppName = "Bulle"
	if !loopback.OpenBrowser(listener.LoginURL(server, apiMount, state)) {
		return "", false, nil
	}
	if _, err := fmt.Fprintln(out, "Opened the browser. Waiting for you to sign in."); err != nil {
		return "", true, err
	}
	code, err := awaitCode(ctx, listener, state, in)
	return code, true, err
}

// pastedCode shows the address that ends on a page with the code, then reads
// the code the user pastes.
func pastedCode(out io.Writer, in io.Reader, server string) (string, error) {
	if _, err := fmt.Fprintf(out, "Open this address in any browser to sign in:\n\n  %s\n\nThen paste the code it shows: ", pasteLoginURL(server)); err != nil {
		return "", err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if code := strings.TrimSpace(line); code != "" {
		return code, nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return "", errors.New("no login code pasted")
}

// pasteLoginURL starts the CLI flow with no loopback port, so porte ends it on
// a page that shows the code.
func pasteLoginURL(server string) string {
	return server + apiMount + "/auth/oidc?" + url.Values{"flow": {"cli"}}.Encode()
}

func awaitCode(ctx context.Context, listener *loopback.Listener, state string, in io.Reader) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		code string
		err  error
	}
	results := make(chan result, 2)
	go func() {
		code, err := listener.WaitForCode(ctx, state)
		results <- result{code, err}
	}()
	go func() {
		line, err := bufio.NewReader(in).ReadString('\n')
		if code := strings.TrimSpace(line); code != "" {
			results <- result{code: code}
			return
		}
		if err != nil && !errors.Is(err, io.EOF) {
			results <- result{err: err}
		}
	}()
	first := <-results
	return first.code, first.err
}
