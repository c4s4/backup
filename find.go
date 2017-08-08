package main

import (
	"fmt"
	"github.com/mattn/go-zglob"
	"os"
	"os/user"
	"sort"
)

// Find files with:
// - includes: the list of globs to include
// - excludes: the list of globs to exclude
// Return the list of files as a slice of strings and error if any
// Relative paths are relative to user's home directory
func FindFiles(includes, excludes []string) ([]string, error) {
	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("getting user home: %v", err)
	}
	err = os.Chdir(user.HomeDir)
	if err != nil {
		return nil, fmt.Errorf("changing to home directory: %v", err)
	}
	var candidates []string
	for _, include := range includes {
		list, _ := zglob.Glob(include)
		for _, file := range list {
			stat, err := os.Stat(file)
			if err == nil && stat.Mode().IsRegular() {
				candidates = append(candidates, file)
			}
		}
	}
	var files []string
	if excludes != nil {
		for index, file := range candidates {
			for _, exclude := range excludes {
				match, err := zglob.Match(exclude, file)
				if match || err != nil {
					candidates[index] = ""
				}
			}
		}
		for _, file := range candidates {
			if file != "" {
				files = append(files, file)
			}
		}
	} else {
		files = candidates
	}
	sort.Strings(files)
	return files, nil
}
