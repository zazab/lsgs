package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/seletskiy/hierr"
)

var (
	red   = color.New(color.FgHiRed).SprintFunc()
	green = color.New(color.FgHiBlue).SprintFunc()

	dirtyMarker = red("✗")
	cleanMarker = green("•")
)

type checkFunc func(path, bool, bool) error

func stdCheck(path path, onlyDirty, quiet bool, commandLine ...string) error {
	out, err := execute(path.linkTo, quiet, commandLine...)
	if err != nil {
		return err
	}

	if out != "" {
		fmt.Printf("%s %s\n", dirtyMarker, path)
		return nil
	}

	if !onlyDirty {
		fmt.Printf("%s %s\n", cleanMarker, path)
	}
	return nil
}

func pushCheck(path path, onlyDirty, quiet bool) error {
	branch, err := getCurrentBranch(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if strings.Contains(branch, "detached") {
		if !onlyDirty {
			fmt.Printf("%s %s\n", cleanMarker, path)
		}

		return nil
	}

	return stdCheck(path, onlyDirty, quiet, "log", "@{push}..")
}

func dirtyCheck(path path, onlyDirty, quiet bool) error {
	return stdCheck(path, onlyDirty, quiet, "status", "--porcelain")
}

func getCurrentBranch(dir string, quiet bool) (string, error) {
	out, err := execute(dir, quiet, "branch", "--points-at", "HEAD")
	if err != nil {
		return "", err
	}

	branches := strings.Split(out, "\n")

	for _, branch := range branches {
		if branch[0] == '*' {
			branch = strings.TrimLeft(branch, "* ")
			branch = strings.TrimRight(branch, "\n")
			return strings.TrimSpace(branch), nil
		}
	}

	return "", errors.New("can't find current branch")
}

func showBranch(path path, onlyDirty, quiet bool) error {
	branch, err := getCurrentBranch(path.linkTo, quiet)
	if err != nil {
		return hierr.Errorf(err, "error processing %s", path.path)
	}

	if branch == "master" {
		branch = green(branch)
	} else {
		branch = red(branch)
	}

	fmt.Printf("%s %s\n", path, branch)
	return nil
}

func execute(dir string, quiet bool, commandLine ...string) (string, error) {
	cmd := exec.Command("git", commandLine...)
	cmd.Dir = dir
	if !quiet {
		cmd.Stderr = os.Stderr
	}

	out, err := cmd.Output()
	if err != nil {
		return "", hierr.Errorf(
			err, "can't run 'git %s' in '%s'",
			strings.Join(commandLine, " "), dir,
		)
	}

	return string(out), nil
}
