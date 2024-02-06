package core

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gofrontier-com/go-utils/output"
)

// AddRemote will add a named remote
func (r *GitRepo) AddRemote(name string, url string) {
	r.Repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
}

func (r *GitRepo) PushWithTags() error {
	rs := config.RefSpec("refs/tags/*:refs/tags/*")
	return r.Repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{rs},
	})
}

func (r *GitRepo) PushWithTagsTo(remoteName string) error {
	rs := config.RefSpec("refs/tags/*:refs/tags/*")
	return r.Repo.Push(&git.PushOptions{
		RefSpecs:   []config.RefSpec{rs},
		RemoteName: remoteName,
	})
}

func (r *GitRepo) CreateTag(tag string) error {
	headRef, _ := r.Repo.Head()
	headHash := headRef.Hash()
	tagger := &object.Signature{
		Name:  r.Author.Name,
		Email: r.Author.Email,
		When:  time.Now(),
	}
	_, err := r.Repo.CreateTag(tag, headHash, &git.CreateTagOptions{
		Tagger:  tagger,
		Message: tag,
	})
	if err != nil {
		return err
	}

	output.PrintlnInfo("Created tag: ", tag)
	return nil
}

func (r *GitRepo) diff(tag string) (*object.Patch, error) {
	revision := plumbing.Revision(tag)
	tagCommitHash, err := r.Repo.ResolveRevision(revision)
	if err != nil {
		return nil, err
	}

	tagCommit, err := r.Repo.CommitObject(*tagCommitHash)
	headRef, _ := r.Repo.Head()

	headHash := headRef.Hash()
	headCommit, _ := r.Repo.CommitObject(headHash)
	return tagCommit.Patch(headCommit)
}

func (r *GitRepo) branchName() (string, error) {
	head, err := r.Repo.Head()
	if err != nil {
		return "", err
	}

	return head.Name().String(), nil
}

func (r *GitRepo) initialCommitHash() string {
	commits, _ := r.Repo.CommitObjects()
	var initialHash plumbing.Hash
	_ = commits.ForEach(func(c *object.Commit) error {
		if c.NumParents() == 0 {
			initialHash = c.Hash
		}
		return nil
	})
	return initialHash.String()
}

func (r *GitRepo) changedFiles(latestTagOrHash string) []string {
	fileschanged := make([]string, 0)
	diff, _ := r.diff(latestTagOrHash)
	stats := diff.Stats()

	for _, stat := range stats {
		fileschanged = append(fileschanged, stat.Name)
	}

	return fileschanged
}

func (r *GitRepo) getTagSuffix() string {
	cb, err := r.branchName()
	if err != nil {
		fmt.Println(err)
	}
	if cb != "refs/heads/main" && cb != "refs/heads/master" {
		return "-unstable"
	}
	return ""
}
