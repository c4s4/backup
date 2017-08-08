package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
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

// Tells if source file should be copied to destination
func ShouldCopy(source, dest string) (bool, error) {
	// if destination doesn't exit, we copy file
	statDest, err := os.Stat(dest)
	if err != nil {
		return true, nil
	}
	// if source doesn't exist, error
	statSource, err := os.Stat(source)
	if err != nil {
		return false, err
	}
	// if files are not the same size, we copy
	if statSource.Size() != statDest.Size() {
		return true, nil
	}
	// compare files
	equal, err := CompareFiles(source, dest)
	if err != nil {
		return false, err
	}
	return !equal, nil
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
