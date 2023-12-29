package core

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gofrontier-com/go-utils/output"
)

func NewVertag(repoRoot string, modulesDir string, authorName string, authorEmail string, dryRun bool, remoteUrl string) *Vertag {
	r := &GitRepo{
		Author: &GitAuthor{
			Name:  authorName,
			Email: authorEmail,
		},
		RemoteUrl: remoteUrl,
	}

	return &Vertag{
		Repo:            r,
		RepoRoot:        repoRoot,
		ModulesDir:      modulesDir,
		ModulesFullPath: path.Join(repoRoot, modulesDir),
		DryRun:          dryRun,
	}
}

func (v *Vertag) Init() error {
	r, err := git.PlainOpen(v.RepoRoot)
	if err != nil {
		return err
	}
	v.Repo.Repo = r

	return nil
}

// GetLatestStableTag returns the most recent tag on the repository.
func (v *Vertag) GetLatestStableTag() error {
	tagRefs, err := v.Repo.Repo.Tags()
	if err != nil {
		return err
	}

	var latestTagCommit *object.Commit
	var latestTagName string
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if strings.Contains(tagRef.Name().String(), "-unstable") {
			// output.PrintlnInfo("Skipping unstable tag: ", tagRef.Name().String())
			return nil
		}
		revision := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := v.Repo.Repo.ResolveRevision(revision)
		if err != nil {
			return err
		}

		commit, err := v.Repo.Repo.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}

		if latestTagCommit == nil {
			latestTagCommit = commit
			latestTagName = tagRef.Name().String()
		}

		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = tagRef.Name().String()
		}

		return nil
	})
	if err != nil {
		return err
	}

	v.LatestStableTag = latestTagName
	if latestTagCommit == nil {
		v.LatestStableSHA = v.Repo.initialCommitHash()
	} else {
		v.LatestStableSHA = latestTagCommit.Hash.String()
	}
	return nil
}

func (v *Vertag) GetRefs() error {
	err := v.getDiffRefs()
	if err != nil {
		return err
	}

	output.PrintfInfo("Comparing\n\tCurrent Branch: %s\nto\n\tLatest Tagged SHA: %s\n\n", v.CurrentBranch, v.LatestStableSHA)

	return nil
}

func (v *Vertag) getDiffRefs() error {
	cb, err := v.Repo.branchName()
	if err != nil {
		return err
	}
	v.CurrentBranch = cb

	err = v.GetLatestStableTag()
	if err != nil {
		return err
	}

	return nil
}

func (v *Vertag) GetChanges() error {
	fileschanged := v.Repo.changedFiles(v.LatestStableSHA)
	dirschanged := changedDirs(fileschanged, v.ModulesDir)
	output.PrintlnInfo("Modules changed")
	for _, d := range dirschanged {
		output.PrintfInfo("\t%s\n", d)
	}
	output.PrintlnInfo("")
	v.ModulesChanged = dirschanged
	return nil
}

func (v *Vertag) CalculateNextTags() error {
	tags := make([]string, 0)

	for _, d := range v.ModulesChanged {
		ltc, err := v.Repo.latestTagContains(d)
		if err != nil {
			output.PrintlnError(err)
		}

		patchVersion := 0
		versionFromFile, _ := getVersion(path.Join(v.ModulesFullPath, d))
		ns := d
		if ltc != "" {
			ltcSplit := strings.Split(ltc, "/") // gives /refs/tags/<namespace>/<version>
			ns = ltcSplit[2]
			versionFromTagSplit := strings.Split(ltcSplit[3], ".")
			versionFromTag := versionFromTagSplit[0] + "." + versionFromTagSplit[1]
			patchFromTagIncSuffix := versionFromTagSplit[2]
			patchFromTag := strings.TrimSuffix(patchFromTagIncSuffix, "-unstable")
			latestPatch, _ := strconv.Atoi(patchFromTag)
			if versionFromFile == versionFromTag {
				patchVersion = latestPatch + 1
			} else {
				patchVersion = 0
			}
		}

		suffix := v.Repo.getTagSuffix()
		tags = append(tags, fmt.Sprintf("%s/%s.%d%s", ns, versionFromFile, patchVersion, suffix))
	}

	v.NextTags = tags

	return nil
}

func (v *Vertag) WriteTags() error {
	for _, tag := range v.NextTags {
		if v.DryRun {
			output.Println("[Dry run] Would have created tag: ", tag)
		} else {
			v.Repo.AddRemote("ci", v.Repo.RemoteUrl)
			err := v.Repo.CreateTag(tag)
			if err != nil {
				output.PrintlnError(err)
				return err
			}
			err = v.Repo.PushWithTagsTo("ci")
			if err != nil {
				output.PrintlnError(err)
				return err
			}
		}
	}

	return nil
}
