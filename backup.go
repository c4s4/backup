package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"flag"
)

var quiet bool

// Configuration is a map that gives configuration for given host
type Configuration map[string]HostConfiguration

// Configuration for given host includes and excludes files
type HostConfiguration struct {
	Includes []string
	Excludes []string
}

// Return found configuration along with an error if any. Configuration file is
// searched as /media/<user>/<unique-id>/.backup
func FindConfigurationFile() (string, error) {
	user, _ := user.Current()
	name := user.Username
	dir := path.Join("/media", name)
	filesInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("reading directory '%s'", dir)
	}
	for _, fileInfo := range filesInfo {
		file := filepath.Join(dir, fileInfo.Name(), ".backup")
		if stat, err := os.Stat(file); err == nil && !stat.IsDir() {
			return file, nil
		}
	}
	return "", fmt.Errorf("no .backup file found in %s subdirectory", dir)
}

// Parse configuration for given file
func ParseConfiguration(file string) (Configuration, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading configuration file '%s': %v", file, err)
	}
	var configuration Configuration
	err = yaml.Unmarshal(bytes, &configuration)
	if err != nil {
		return nil, fmt.Errorf("parsing configuration file '%s': %v", file, err)
	}
	return configuration, nil
}

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

// Return sha1 hash for given file and error if any
func Sha1File(file string) ([]byte, error) {
	f, err := os.Open(file)
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

// Tells if a file should be copied to destination
func ShouldCopy(source, dest string) (bool, error) {
	statDest, err := os.Stat(dest)
	if err != nil {
		return true, nil
	}
	statSource, err := os.Stat(source)
	if err != nil {
		return false, err
	}
	if statSource.Size() != statDest.Size() {
		return true, nil
	}
	sha1Source, err := Sha1File(source)
	if err != nil {
		return false, fmt.Errorf("getting sha1 for file %s: %v", source, err)
	}
	sha1Dest, err := Sha1File(dest)
	if err != nil {
		return false, fmt.Errorf("getting sha1 for file %s: %v", dest, err)
	}
	if len(sha1Source) != len(sha1Dest) {
		return false, nil
	}
	for index, byte := range sha1Source {
		if sha1Dest[index] != byte {
			return true, nil
		}
	}
	return false, nil
}

// Copy source file to destination, preserving mode and time
func CopyFile(source, dest string) error {
	copy, err := ShouldCopy(source, dest)
	if err != nil {
		return err
	}
	if !copy {
		return nil
	}
	if !quiet {
		fmt.Println("-", source)
	}
	dir := filepath.Dir(dest)
	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("making destination directory %s: %v", dir, err)
		}
	}
	from, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("opening source file '%s': %v", source, err)
	}
	info, err := from.Stat()
	if err != nil {
		return fmt.Errorf("getting mode of source file '%s': %v", source, err)
	}
	defer from.Close()
	to, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating desctination file '%s': %v", dest, err)
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	if err != nil {
		return fmt.Errorf("copying file: %v", err)
	}
	err = to.Sync()
	if err != nil {
		return fmt.Errorf("syncing destination file: %v", err)
	}
	if runtime.GOOS != "windows" {
		err = to.Chmod(info.Mode())
		if err != nil {
			return fmt.Errorf("changing mode of destination file '%s': %v", dest, err)
		}
	}
	return nil
}

// Copy a list of files to destination directory
func CopyFiles(files []string, dir string) error {
	for _, file := range files {
		destination := filepath.Join(dir, file)
		err := CopyFile(file, destination)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run backup
func run() error {
	file, err := FindConfigurationFile()
	if err != nil {
		return err
	}
	destination := filepath.Dir(file)
	configuration, err := ParseConfiguration(file)
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("getting hostname: %v", err)
	}
	config, ok := configuration[hostname]
	if !ok {
		fmt.Errorf("hostname '%s' not found in configuration", hostname)
	}
	files, err := FindFiles(config.Includes, config.Excludes)
	if err != nil {
		return fmt.Errorf("getting files to backup: %v", err)
	}
	err = CopyFiles(files, destination)
	if err != nil {
		return fmt.Errorf("copying file: %v", err)
	}
	return nil
}

// Main function
func main() {
	q := flag.Bool("quiet", false, "Don't print files to copy")
	flag.Parse()
	quiet = *q
	err := run()
	if err != nil {
		println("ERROR", err.Error())
		os.Exit(1)
	}
}
