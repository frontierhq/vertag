package core

import (
	"github.com/go-git/go-git/v5"
)

type Vertag struct {
	Repo            *GitRepo
	RepoRoot        string
	ModulesDir      string
	ModulesFullPath string
	DryRun          bool
	LatestStableTag string
	LatestStableSHA string
	CurrentBranch   string
	ModulesChanged  []string
	NextTags        []string
}

type GitRepo struct {
	Repo      *git.Repository
	Author    *GitAuthor
	RemoteUrl string
}

type GitAuthor struct {
	Name  string
	Email string
}
