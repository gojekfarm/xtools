package main

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
