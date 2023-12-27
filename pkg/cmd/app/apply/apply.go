package apply

import (
	"github.com/common-nighthawk/go-figure"
	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/vertag/pkg/core"
)

func Apply(repoRoot string, modulesDir string, authorName string, authorEmail string, dryRun bool, remoteUrl string) error {
	myFigure := figure.NewFigure("VerTag", "", true)
	myFigure.Print()

	vt := core.NewVertag(repoRoot, modulesDir, authorName, authorEmail, dryRun, remoteUrl)
	err := vt.Init()
	if err != nil {
		output.PrintlnError(err)
	}

	err = vt.GetRefs()
	if err != nil {
		output.PrintlnError(err)
	}

	err = vt.GetChanges()
	if err != nil {
		output.PrintlnError(err)
	}

	err = vt.CalculateNextTags()
	if err != nil {
		output.PrintlnError(err)
	}

	err = vt.WriteTags()
	if err != nil {
		output.PrintlnError(err)
	}

	return nil
}
