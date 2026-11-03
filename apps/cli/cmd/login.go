package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/FacileStudio/porte/loopback"
	"github.com/spf13/cobra"

	"github.com/FacileStudio/Bulle/core/settings"
)

func newLoginCmd() *cobra.Command {
	var device bool
	cmd := &cobra.Command{
		Use:   "login [url]",
		Short: "Sign in to a Bulle server and store the token",
		Long: `Sign in to a Bulle server through the browser. The server keeps your provider
keys and the CLI sends model calls through it with the token this command stores.

Without a url, login uses BULLE_SERVER_URL, server.url from the config file, the
server of the last login, or https://bulle.facile.studio. Over SSH, with --device, or when no browser can
be opened, the command prints an address to open in any browser and reads the
one-time code that page shows.`,
		Example: `  bulle login
  bulle login https://bulle.example.com
  bulle login --device`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			server, err := loginServer(args)
			if err != nil {
				return err
			}
			return runLogin(c.Context(), c.OutOrStdout(), c.InOrStdin(), server, device)
		},
	}
	cmd.Flags().BoolVar(&device, "device", false, "sign in from a browser on another device and paste the code")
	return cmd
}

func loginServer(args []string) (string, error) {
	raw := ""
	if len(args) > 0 {
		raw = args[0]
	} else {
		cfg, err := settings.Settings("", settings.Config{})
		if err != nil {
			return "", err
		}
		raw = cfg.Server.URL
	}
	if raw == "" {
		raw = settings.DefaultServerURL
	}
	return settings.NormalizeServerURL(raw)
}

func runLogin(ctx context.Context, out io.Writer, in io.Reader, server string, device bool) error {
	code, err := loginCode(ctx, out, in, server, device || remoteSession())
	if err != nil {
		return err
	}
	token, err := exchangeCode(ctx, server, code)
	if err != nil {
		return err
	}
	if err := settings.SaveLogin(server, token); err != nil {
		return fmt.Errorf("storing the token: %w", err)
	}
	fmt.Fprintf(out, "Signed in to %s\n", server)
	return nil
}

// loginCode gets a one-time login code. A local browser hands it back through
// the loopback listener. Any other browser, such as the laptop in front of an
// SSH session, signs in through an address without a port, where porte shows
// the code to paste here instead of redirecting to 127.0.0.1.
func loginCode(ctx context.Context, out io.Writer, in io.Reader, server string, pasted bool) (string, error) {
	if !pasted {
		state, err := loopback.RandomState()
		if err != nil {
			return "", err
		}
		listener, err := loopback.Listen()
		if err != nil {
			return "", fmt.Errorf("opening the loopback listener: %w", err)
		}
		defer func() { _ = listener.Close() }()
		listener.AppName = "Bulle"
		if loopback.OpenBrowser(listener.LoginURL(server, apiMount, state)) {
			fmt.Fprintln(out, "Opened the browser. Waiting for you to sign in.")
			return awaitCode(ctx, listener, state, in)
		}
	}
	fmt.Fprintf(out, "Open this address in any browser to sign in:\n\n  %s\n\nThen paste the code it shows: ", pasteLoginURL(server))
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

// remoteSession reports an SSH login, where a browser this process opens is
// not the one in front of the user.
func remoteSession() bool {
	return os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != ""
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

func tokenOrigin() string {
	if settings.TokenFromEnv() {
		return "BULLE_TOKEN"
	}
	return "stored login"
}
