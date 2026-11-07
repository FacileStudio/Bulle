package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/Bulle/core/settings"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Forget the stored Bulle server login",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if err := settings.ClearLogin(); err != nil {
				return err
			}
			_, err := fmt.Fprintln(c.OutOrStdout(), "Signed out. BULLE_TOKEN, if set, is untouched.")
			return err
		},
	}
}
