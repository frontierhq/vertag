package apply

import (
	"os"
	"path"

	"github.com/gofrontier-com/vertag/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

var (
	modulesDir  string
	repoRoot    string
	authorName  string
	authorEmail string
	remoteUrl   string
	dryRun      bool
)

// NewCmdApply creates a command to apply config
func NewCmdApply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply config",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := apply.Apply(repoRoot, modulesDir, authorName, authorEmail, dryRun, remoteUrl); err != nil {
				return err
			}

			return nil
		},
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&modulesDir, "modules-dir", "m", path.Join(wd, "modules"), "Directory of the modules")
	cmd.Flags().StringVarP(&repoRoot, "repo", "r", wd, "Root directory of the repo")
	cmd.Flags().StringVarP(&authorName, "author-name", "n", wd, "Name of the commiter")
	cmd.Flags().StringVarP(&authorEmail, "author-email", "e", wd, "Email of the commiter")
	cmd.Flags().StringVarP(&remoteUrl, "remote-url", "u", "", "CI Remote URL")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Email of the commiter")
	return cmd
}
