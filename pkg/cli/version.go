package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time via
// -ldflags="-X github.com/nicklasfrahm/kontinuum/pkg/cli.version=...".
var version = "dev"

// NewVersionCmd builds the version command, which prints the kontinuum version.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the kontinuum version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}
