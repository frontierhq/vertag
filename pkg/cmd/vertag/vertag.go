package vertag

import (
	"github.com/gofrontier-com/vertag/pkg/cmd/cli/apply"
	vers "github.com/gofrontier-com/vertag/pkg/cmd/cli/version"
	"github.com/spf13/cobra"
)

func NewRootCmd(version string, commit string, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                   "vertag",
		DisableFlagsInUseLine: true,
		Short:                 "vertag is the command line tool for vertag",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.AddCommand(apply.NewCmdApply())
	rootCmd.AddCommand(vers.NewCmdVersion(version, commit, date))

	return rootCmd
}
