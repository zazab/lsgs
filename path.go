package main

import (
	"fmt"
	"path/filepath"
)

type path struct {
	path   string
	linkTo string
}

func (path path) String() string {
	if path.linkTo != path.path {
		return fmt.Sprintf("%s -> %s", path.path, path.linkTo)
	}

	return path.path
}

func newPath(pathString string) (path, error) {
	linkTo, err := filepath.EvalSymlinks(pathString)
	if err != nil {
		return path{}, err
	}

	return path{pathString, linkTo}, nil
}
