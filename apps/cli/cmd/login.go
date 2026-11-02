package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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

Without a url, login uses BULLE_SERVER_URL, server.url from the config file, or
the server of the last login. With --device the command prints the sign-in
address and also accepts the one-time code pasted on stdin, for a machine whose
browser cannot reach the loopback listener.`,
		Example: `  bulle login https://bulle.example.com
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
	cmd.Flags().BoolVar(&device, "device", false, "print the sign-in address and accept a pasted code")
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
		return "", errors.New("no server: run bulle login <url> or set BULLE_SERVER_URL")
	}
	return settings.NormalizeServerURL(raw)
}

func runLogin(ctx context.Context, out io.Writer, in io.Reader, server string, device bool) error {
	state, err := loopback.RandomState()
	if err != nil {
		return err
	}
	listener, err := loopback.Listen()
	if err != nil {
		return fmt.Errorf("opening the loopback listener: %w", err)
	}
	defer func() { _ = listener.Close() }()
	listener.AppName = "Bulle"

	target := listener.LoginURL(server, apiMount, state)
	if device || !loopback.OpenBrowser(target) {
		fmt.Fprintf(out, "Open this address to sign in:\n\n  %s\n\n", target)
		fmt.Fprintln(out, "If the browser cannot reach this machine, paste the code it shows here.")
	} else {
		fmt.Fprintln(out, "Opened the browser. Waiting for you to sign in.")
	}

	code, err := awaitCode(ctx, listener, state, in)
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
