package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/Bulle/core/settings"
)

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the account the stored token signs in as",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := settings.Settings("", settings.Config{})
			if err != nil {
				return err
			}
			if cfg.Server.URL == "" {
				return errors.New("no server configured: run bulle login <url>")
			}
			token := settings.Token()
			if token == "" {
				return errors.New("not signed in: run bulle login")
			}
			server, err := settings.NormalizeServerURL(cfg.Server.URL)
			if err != nil {
				return err
			}
			user, code, err := fetchUser(c.Context(), server, token)
			if err != nil {
				return fmt.Errorf("asking %s: %w", server, err)
			}
			if code == http.StatusUnauthorized {
				return fmt.Errorf("the token from the %s is not valid on %s: run bulle login", tokenOrigin(), server)
			}
			if code != http.StatusOK {
				return fmt.Errorf("%s answered status %d", server, code)
			}
			_, err = fmt.Fprintf(c.OutOrStdout(), "%s <%s>\n%s\n", user.Name, user.Email, server)
			return err
		},
	}
}
