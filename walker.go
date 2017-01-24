package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func walkDirs(
	path path, depth, maxDepth int, onlyDirty, quiet bool, checker checkFunc,
	exclude *regexp.Regexp,
) error {
	_, last := filepath.Split(path.linkTo)
	if last == ".git" {
		// skipping .git dir
		return nil
	}

	info, err := os.Stat(path.linkTo)
	if err != nil {
		return fmt.Errorf("can't stat '%s': %s", path, err)
	}

	if !info.IsDir() {
		return nil
	}

	_, err = os.Stat(filepath.Join(path.linkTo, ".git"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(
			"can't stat '%s': %s", filepath.Join(path.path, ".git"), err,
		)
	}

	if err == nil { // this is git repo
		err = checker(path, onlyDirty, quiet)
		if err != nil {
			logger.Error(err)
		}
	}

	if depth > maxDepth {
		return nil
	}

	return goDeeper(path, depth, maxDepth, onlyDirty, quiet, checker, exclude)
}

func goDeeper(
	path path, depth, maxDepth int, onlyDirty, quiet bool, checker checkFunc,
	exclude *regexp.Regexp,
) error {
	files, err := ioutil.ReadDir(path.path)
	if err != nil {
		return fmt.Errorf("can't read dir '%s': %s", path, err)
	}

	for _, file := range files {
		filePath, err := newPath(filepath.Join(path.path, file.Name()))
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error(err)
			}

			continue
		}

		info, err := os.Stat(filePath.linkTo)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error(err)
			}

			continue
		}

		if info.IsDir() {
			if exclude != nil && exclude.MatchString(info.Name()) {
				continue
			}

			err := walkDirs(
				filePath, depth+1, maxDepth, onlyDirty, quiet, checker, exclude,
			)
			if err != nil {
				logger.Error(err)
			}
		}
	}

	return nil

}
