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

func NewVertag(repoRoot string, modulesDir string, authorName string, authorEmail string, dryRun bool, remoteUrl string, branchDiff bool) *Vertag {
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
		BranchDiff:      branchDiff,
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

func compareVersions(tag1, tag2 string) bool {
	// Extract version numbers and stability flag
	parts1 := strings.Split(tag1, "/")
	parts2 := strings.Split(tag2, "/")

	version1 := strings.Split(parts1[len(parts1)-1], "-")[0] // "1.1.9"
	version2 := strings.Split(parts2[len(parts2)-1], "-")[0] // "1.1.10"

	isUnstable1 := strings.HasSuffix(parts1[len(parts1)-1], "-unstable")
	isUnstable2 := strings.HasSuffix(parts2[len(parts2)-1], "-unstable")

	// If stability differs, stable version is greater
	if isUnstable1 != isUnstable2 {
		return isUnstable1 && !isUnstable2
	}

	// Compare version numbers
	v1 := strings.Split(version1, ".")
	v2 := strings.Split(version2, ".")

	for i := 0; i < 3; i++ {
		num1, _ := strconv.Atoi(v1[i])
		num2, _ := strconv.Atoi(v2[i])
		if num1 != num2 {
			return num1 < num2
		}
	}
	return false
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

		if commit.Committer.When.Equal(latestTagCommit.Committer.When) {
			if compareVersions(latestTagName, tagRef.Name().String()) {
				latestTagCommit = commit
				latestTagName = tagRef.Name().String()
			}
		} else if commit.Committer.When.After(latestTagCommit.Committer.When) {
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
	v.ComparisonSHA = v.LatestStableSHA
	return nil
}

func (v *Vertag) latestTagContains(tagContains string) error {
	tagRefs, err := v.Repo.Repo.Tags()
	if err != nil {
		return err
	}

	var latestTagCommit *object.Commit
	var latestTagName string
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		moduleName := strings.Split(tagRef.Name().String(), "/")[2]
		if moduleName == tagContains {
			revision := plumbing.Revision(tagRef.Name().String())
			tagCommitHash, err := v.Repo.Repo.ResolveRevision(revision)
			if err != nil {
				return err
			}

			commit, err := v.Repo.Repo.CommitObject(*tagCommitHash)
			if err != nil {
				return err
			}
			output.PrintlnInfo("Checking tag", tagRef.Name().String(), "for module", tagContains, "date:", commit.Committer.When)

			if latestTagCommit == nil {
				latestTagCommit = commit
				latestTagName = tagRef.Name().String()
			}

			if commit.Committer.When.Equal(latestTagCommit.Committer.When) {
				if compareVersions(latestTagName, tagRef.Name().String()) {
					latestTagCommit = commit
					latestTagName = tagRef.Name().String()
				}
			} else if commit.Committer.When.After(latestTagCommit.Committer.When) {
				latestTagCommit = commit
				latestTagName = tagRef.Name().String()
			}

			if commit.Committer.When.Equal(latestTagCommit.Committer.When) {
				if !strings.Contains(tagRef.Name().String(), "-unstable") {
					latestTagCommit = commit
					latestTagName = tagRef.Name().String()
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	output.PrintlnInfo("Latest tag for module", tagContains, "is", latestTagName)

	v.LatestTag = latestTagName

	return nil
}

func (v *Vertag) GetLatestBranchUnstableTag() error {
	tagRefs, err := v.Repo.Repo.Tags()
	if err != nil {
		return err
	}

	head, err := v.Repo.Repo.Head()
	if err != nil {
		return err
	}
	currentCommit, err := v.Repo.Repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	tagRefMap := make(map[string]string)
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagRefMap[tagRef.Name().String()] = tagRef.Hash().String()
		return nil
	})
	if err != nil {
		return err
	}

	for {
		found := false
		for tagRefName, tagRefHash := range tagRefMap {
			fmt.Println(tagRefName, tagRefHash, currentCommit.Hash.String())
			if strings.Contains(tagRefName, "-unstable") {
				if tagRefHash == currentCommit.Hash.String() {
					found = true
				}
			}
		}
		if err != nil {
			return err
		}

		if found {
			v.LatestBranchUnstableSHA = currentCommit.Hash.String()
			v.ComparisonSHA = v.LatestBranchUnstableSHA
			break
		}

		currentCommit, err = currentCommit.Parents().Next()
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *Vertag) GetRefs() error {
	err := v.getDiffRefs()
	if err != nil {
		return err
	}

	output.PrintfInfo("Comparing\n\tCurrent Branch: %s\nto\n\tLatest Tagged SHA: %s\n\n", v.CurrentBranch, v.ComparisonSHA)

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

	if v.BranchDiff {
		err = v.GetLatestBranchUnstableTag()
	}

	return nil
}

func (v *Vertag) GetChanges() error {
	fileschanged := v.Repo.changedFiles(v.ComparisonSHA)
	output.PrintlnInfo("Files changed", fileschanged)
	dirschanged := changedDirs(fileschanged, v.ModulesDir, v.ModulesFullPath)
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
		err := v.latestTagContains(d)
		if err != nil {
			output.PrintlnError(err)
		}

		patchVersion := 0
		versionFromFile, _ := getVersion(path.Join(v.ModulesFullPath, d))
		ns := d

		if v.LatestTag != "" {
			ltcSplit := strings.Split(v.LatestTag, "/") // gives /refs/tags/<namespace>/<version>
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
	if len(v.NextTags) == 0 {
		output.PrintlnInfo("No tags to write and push")
		return nil
	}

	if v.Repo.RemoteUrl != "" {
		v.Repo.AddRemote("ci", v.Repo.RemoteUrl)
	}

	for _, tag := range v.NextTags {
		if v.DryRun {
			output.Println("[Dry run] Would have created tag: ", tag)
		} else {
			err := v.Repo.CreateTag(tag)
			if err != nil {
				return fmt.Errorf("failed to create tag %s: %w", tag, err)
			}

			if v.Repo.RemoteUrl != "" {
				if err := v.Repo.PushWithTagsTo("ci"); err != nil {
					return fmt.Errorf("failed to push tags to remote 'ci': %w", err)
				}
			} else {
				if err := v.Repo.PushWithTags(); err != nil {
					return fmt.Errorf("failed to push tags: %w", err)
				}
			}
		}
	}

	return nil
}
