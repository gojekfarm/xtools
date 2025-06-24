package main

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/semver"
)

func GetAllTags(dir string) (map[string][]string, error) {
	git, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	refs, err := git.Tags()
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string][]string)

	refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		parts := strings.Split(name, "/")

		semver := parts[len(parts)-1]
		module := strings.Join(parts[:len(parts)-1], "/")

		if _, ok := tagMap[module]; !ok {
			tagMap[module] = []string{}
		}

		tagMap[module] = append(tagMap[module], semver)

		return nil
	})

	return tagMap, nil
}

func GetLatestTag(dir string) (string, error) {
	allTags, err := GetAllTags(dir)
	if err != nil {
		return "", err
	}

	versions := allTags[""]
	semver.Sort(versions)

	return versions[len(versions)-1], nil
}

func HasUncommittedChanges(dir string) (bool, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return false, err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return false, err
	}
	status, err := wt.Status()
	if err != nil {
		return false, err
	}
	return !status.IsClean(), nil
}
