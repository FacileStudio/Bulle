package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

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
	_, err = fmt.Fprintf(out, "Signed in to %s\n", server)
	return err
}

// remoteSession reports an SSH login, where a browser this process opens is
// not the one in front of the user.
func remoteSession() bool {
	return os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != ""
}

func tokenOrigin() string {
	if settings.TokenFromEnv() {
		return "BULLE_TOKEN"
	}
	return "stored login"
}
