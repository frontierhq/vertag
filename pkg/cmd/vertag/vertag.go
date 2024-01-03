package vertag

import (
	"os"
	"path"

	"github.com/common-nighthawk/go-figure"
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/vertag/pkg/core"
	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
)

var (
	outputFmt   string
	shortened   bool
	modulesDir  string
	repoRoot    string
	authorName  string
	authorEmail string
	remoteUrl   string
	dryRun      bool
	vers        bool
	help        bool
)

func Apply(repoRoot string, modulesDir string, authorName string, authorEmail string, dryRun bool, remoteUrl string) error {
	myFigure := figure.NewFigure("VerTag", "", true)
	myFigure.Print()

	vt := core.NewVertag(repoRoot, modulesDir, authorName, authorEmail, dryRun, remoteUrl)
	err := vt.Init()
	if err != nil {
		output.PrintlnError(err)
		return err
	}

	err = vt.GetRefs()
	if err != nil {
		output.PrintlnError(err)
		return err
	}

	err = vt.GetChanges()
	if err != nil {
		output.PrintlnError(err)
		return err
	}

	err = vt.CalculateNextTags()
	if err != nil {
		output.PrintlnError(err)
		return err
	}

	err = vt.WriteTags()
	if err != nil {
		output.PrintlnError(err)
		return err
	}

	return nil
}

func NewRootCmd(version string, commit string, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vertag",
		Short: "vertag is the command line tool for vertag",
		RunE: func(cmd *cobra.Command, args []string) error {
			if help {
				if err := cmd.Help(); err != nil {
					return err
				}
			}

			if vers {
				resp := goVersion.FuncWithOutput(shortened, version, commit, date, outputFmt)
				output.PrintfInfo(resp)

				return nil
			}

			if err := Apply(repoRoot, modulesDir, authorName, authorEmail, dryRun, remoteUrl); err != nil {
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
	cmd.Flags().BoolVarP(&vers, "version", "v", false, "Version")
	cmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number")
	cmd.Flags().StringVarP(&outputFmt, "output", "o", "json", "Output format")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Version")

	return cmd
}
