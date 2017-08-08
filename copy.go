package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

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
