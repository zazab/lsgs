package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type checkFunc func(path, bool, bool) error

func stdCheck(
	path path,
	onlyDirty, quiet bool,
	commandLine ...string) error {
	cmd := exec.Command("git", commandLine...)
	cmd.Dir = path.linkTo
	if !quiet {
		cmd.Stderr = os.Stderr
	}

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(
			"error running 'git %s' in '%s': %s",
			strings.Join(commandLine, " "), path.linkTo, err,
		)
	}

	if len(out) > 0 {
		fmt.Printf("%s %s\n", red("✗"), path)
	} else {
		if !onlyDirty {
			fmt.Printf("%s %s\n", green("•"), path)
		}
	}
	return nil
}

func pushCheck(path path, onlyDirty, quiet bool) error {
	return stdCheck(path, onlyDirty, quiet, "log", "@{push}..")
}

func dirtyCheck(path path, onlyDirty, quiet bool) error {
	return stdCheck(path, onlyDirty, quiet, "status", "--porcelain")
}

func showBranch(path path, onlyDirty, quiet bool) error {
	cmd := exec.Command("git", "branch", "--points-at", "HEAD")
	cmd.Dir = path.linkTo
	if !quiet {
		cmd.Stderr = os.Stderr
	}

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(
			"error running 'git branch' in '%s': %s", path.linkTo, err,
		)
	}

	branches := strings.Split(string(out), "\n")

	var currentBranch string
	for _, branch := range branches {
		if branch == "" {
			continue
		}
		if branch[0] == '*' {
			currentBranch = strings.TrimLeft(branch, "* ")
			currentBranch = strings.TrimRight(currentBranch, "\n")
			break
		}
	}

	if currentBranch == "master" {
		currentBranch = green(currentBranch)
	} else {
		currentBranch = red(currentBranch)
	}

	fmt.Printf("%s %s\n", path, currentBranch)
	return nil
}
