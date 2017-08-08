package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

// Compute hash for file and put it in the channel
func HashFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	hash := sha1.New()
	if _, err := io.Copy(hash, f); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// Compare two files and tell if they are equal, return an error if a file
// doesn't exist of is not readable
func CompareFiles(file1, file2 string) (bool, error) {
	sha1File1, err := HashFile(file1)
	if err != nil {
		return false, fmt.Errorf("getting sha1 for file %s: %v", file1, err)
	}
	sha1File2, err := HashFile(file2)
	if err != nil {
		return false, fmt.Errorf("getting sha1 for file %s: %v", file2, err)
	}
	if len(sha1File1) != len(sha1File2) {
		return false, nil
	}
	for i := 0; i < len(sha1File1); i++ {
		if sha1File1[i] != sha1File2[i] {
			return false, nil
		}
	}
	return true, nil
}
