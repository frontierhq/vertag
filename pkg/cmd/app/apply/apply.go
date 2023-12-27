package apply

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gofrontier-com/go-utils/output"
)

// latestStableTag returns the most recent tag on the repository.
func latestStableTag(r *git.Repository) (string, error) {
	tagRefs, err := r.Tags()
	if err != nil {
		return "", err
	}

	var latestTagCommit *object.Commit
	var latestTagName string
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if strings.Contains(tagRef.Name().String(), "-unstable") {
			// output.PrintlnInfo("Skipping unstable tag: ", tagRef.Name().String())
			return nil
		}
		revision := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := r.ResolveRevision(revision)
		if err != nil {
			return err
		}

		commit, err := r.CommitObject(*tagCommitHash)
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
		return "", err
	}

	return latestTagName, nil
}

func latestTagContains(r *git.Repository, tagContains string) (string, error) {
	tagRefs, err := r.Tags()
	if err != nil {
		return "", err
	}

	var latestTagCommit *object.Commit
	var latestTagName string
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if strings.Contains(tagRef.Name().String(), tagContains) {
			revision := plumbing.Revision(tagRef.Name().String())
			tagCommitHash, err := r.ResolveRevision(revision)
			if err != nil {
				return err
			}

			commit, err := r.CommitObject(*tagCommitHash)
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
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return latestTagName, nil
}

func tagToSha(r *git.Repository, tag string) (string, error) {
	revision := plumbing.Revision(tag)
	tagCommitHash, err := r.ResolveRevision(revision)
	if err != nil {
		return "", err
	}

	return tagCommitHash.String(), nil
}

func branchName(r *git.Repository) (string, error) {

	head, err := r.Head()
	if err != nil {
		return "", err
	}

	return head.Name().String(), nil
}

func initalCommitHash(r *git.Repository) string {
	commits, _ := r.CommitObjects()
	var initialHash plumbing.Hash
	_ = commits.ForEach(func(c *object.Commit) error {
		if c.NumParents() == 0 {
			initialHash = c.Hash
		}
		return nil
	})
	return initialHash.String()
}

func getDiffRefs(r *git.Repository) (string, string) {
	cb, err := branchName(r)
	if err != nil {
		return "", ""
	}

	lt, err := latestStableTag(r)
	if err != nil {
		return "", ""
	}
	if lt == "" {
		lt = initalCommitHash(r)
	}

	return cb, lt
}

func diff(r *git.Repository, tag string) (*object.Patch, error) {
	revision := plumbing.Revision(tag)
	tagCommitHash, err := r.ResolveRevision(revision)
	if err != nil {
		return nil, err
	}

	tagCommit, err := r.CommitObject(*tagCommitHash)
	headRef, _ := r.Head()

	headHash := headRef.Hash()
	headCommit, _ := r.CommitObject(headHash)
	return tagCommit.Patch(headCommit)
}

func removeFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func changedFiles(r *git.Repository, latestTagOrHash string) []string {
	fileschanged := make([]string, 0)
	diff, _ := diff(r, latestTagOrHash)
	stats := diff.Stats()

	for _, stat := range stats {
		fileschanged = append(fileschanged, stat.Name)
	}

	fps := diff.FilePatches()
	for _, fp := range fps {
		f, t := fp.Files()
		if t == nil {
			fileschanged = removeFromSlice(fileschanged, f.Path())
		}
	}
	return fileschanged
}

func changedDirs(filesChanged []string, modulesDir string) []string {
	dirschanged := make([]string, 0)
	for _, fc := range filesChanged {
		if strings.HasPrefix(fc, modulesDir) {
			a := strings.Split(fc, "/")
			if len(a) > 2 { // make sure the changed file is of the form [azure resource-group main.tf]
				inDirschanged := false
				for _, dir := range dirschanged {
					if dir == a[1] {
						inDirschanged = true
					}
				}
				if inDirschanged == false {
					dirschanged = append(dirschanged, a[1])
				}
			}
		}
	}

	return dirschanged
}

func getVersion(dir string) (string, error) {
	file, err := os.Open(path.Join(dir, "VERSION"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	retval := strings.TrimSuffix(string(bytes), "\n")

	return retval, nil
}

func getTagSuffix(r *git.Repository) string {
	cb, err := branchName(r)
	if err != nil {
		fmt.Println(err)
	}
	if cb != "refs/heads/main" && cb != "refs/heads/master" {
		return "-unstable"
	}
	return ""
}

func createTag(r *git.Repository, tag string, authorName string, authorEmail string) error {
	headRef, _ := r.Head()
	headHash := headRef.Hash()
	tagger := &object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}
	_, err := r.CreateTag(tag, headHash, &git.CreateTagOptions{
		Tagger:  tagger,
		Message: tag,
	})
	if err != nil {
		return err
	}
	output.PrintlnInfo("Created tag: ", tag)
	return nil
}

func pushWithTags(r *git.Repository) error {
	rs := config.RefSpec("refs/tags/*:refs/tags/*")
	return r.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{rs},
	})
}

func findNextTags(repo *git.Repository, dirschanged []string, modulesFullPath string) ([]string, error) {
	tags := make([]string, 0)

	for _, d := range dirschanged {
		ltc, err := latestTagContains(repo, d)
		if err != nil {
			output.PrintlnError(err)
		}

		patchVersion := 0
		versionFromFile, _ := getVersion(path.Join(modulesFullPath, d))
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

		suffix := getTagSuffix(repo)
		tags = append(tags, fmt.Sprintf("%s/%s.%d%s", ns, versionFromFile, patchVersion, suffix))
	}

	return tags, nil
}

// AddRemove will add a named remote
func addRemote(r *git.Repository, name string, url string) {
	r.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
}

func createTags(r *git.Repository, nextTags []string, dryRun bool, authorName string, authorEmail string, remoteUrl string) error {
	for _, tag := range nextTags {
		if dryRun {
			output.Println("[Dry run] Would have created tag: ", tag)
		} else {
			addRemote(r, "ci", remoteUrl)
			err := createTag(r, tag, authorName, authorEmail)
			if err != nil {
				output.PrintlnError(err)
				return err
			}
			err = pushWithTags(r)
			if err != nil {
				output.PrintlnError(err)
				return err
			}
		}
	}

	return nil
}

func Apply(repoRoot string, modulesDir string, authorName string, authorEmail string, dryRun bool, remoteUrl string) error {

	myFigure := figure.NewFigure("VerTag", "", true)
	myFigure.Print()

	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		output.PrintlnError(err)
	}

	cb, lt := getDiffRefs(r)
	ltsha, _ := tagToSha(r, lt)
	output.PrintfInfo("Comparing\n\tCurrent Branch: %s\nto\n\tLatest Tagged SHA: %s\n\n", cb, ltsha)

	fileschanged := changedFiles(r, lt)
	dirschanged := changedDirs(fileschanged, modulesDir)
	output.PrintlnInfo("Modules changed")
	for _, d := range dirschanged {
		output.PrintfInfo("\t%s\n", d)
	}
	output.PrintlnInfo("")

	nextTags, err := findNextTags(r, dirschanged, path.Join(repoRoot, modulesDir))
	if err != nil {
		output.PrintlnError(err)
	}

	createTags(r, nextTags, dryRun, authorName, authorEmail, remoteUrl)

	return nil
}
