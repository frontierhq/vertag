package apply

import (
	"os"

	"github.com/frontierdigital/vertag/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

var (
	modulesRoot string
	authorName  string
	authorEmail string
)

// NewCmdApply creates a command to apply config
func NewCmdApply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply config",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := apply.Apply(modulesRoot, authorName, authorEmail); err != nil {
				return err
			}

			return nil
		},
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&modulesRoot, "modules-dir", "m", wd, "Root directory for modules")
	cmd.Flags().StringVarP(&authorName, "author-name", "n", wd, "Name of the commiter")
	cmd.Flags().StringVarP(&authorEmail, "author-email", "e", wd, "Email of the commiter")
	return cmd
}
