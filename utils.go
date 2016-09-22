package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/seletskiy/hierr"
)

func printUntracked(path path, postfix ...string) {
	if len(postfix) > 0 {
		fmt.Printf("%s %s %s\n", untrackedMarker, path, yellow(postfix[0]))
	} else {
		fmt.Printf("%s %s\n", untrackedMarker, path)
	}
}

func printEmpty(path path, postfix ...string) {
	if len(postfix) > 0 {
		fmt.Printf("%s %s %s\n", emptyMarker, path, blue(postfix[0]))
	} else {
		fmt.Printf("%s %s\n", emptyMarker, path)
	}
}

func printDirty(path path, postfix ...string) {
	if len(postfix) > 0 {
		fmt.Printf("%s %s %s\n", dirtyMarker, path, red(postfix[0]))
	} else {
		fmt.Printf("%s %s\n", dirtyMarker, path)
	}
}

func printClean(path path, onlyDirty bool, postfix ...string) {
	if onlyDirty {
		return
	}

	if len(postfix) > 0 {
		fmt.Printf("%s %s %s\n", cleanMarker, path, green(postfix[0]))
	} else {
		fmt.Printf("%s %s\n", cleanMarker, path)
	}
}

func printUnpushed(path path, postfix ...string) {
	if len(postfix) > 0 {
		fmt.Printf("%s %s %s\n", unpushedMarker, path, yellow(postfix[0]))
	} else {
		fmt.Printf("%s %s\n", unpushedMarker, path)
	}
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

func isEmpty(dir string, quiet bool) (bool, error) {
	out, err := execute(dir, quiet, "branch", "-lr")
	if err != nil {
		return false, err
	}

	if len(out) == 0 {
		return true, nil
	}

	return false, nil
}

func isUntracked(dir string, quiet bool) (bool, error) {
	out, err := execute(dir, quiet, "status", "--porcelain", "-b")
	if err != nil {
		return false, err
	}

	untrackedRegexp := regexp.MustCompile(`^\?\?\s`)
	for _, row := range strings.Split(out, "\n") {
		if untrackedRegexp.MatchString(row) {
			return true, nil
		}
	}

	return false, nil
}

func isTracked(dir string, quiet bool) (bool, error) {
	out, err := execute(dir, quiet, "status", "--porcelain", "-b")
	if err != nil {
		return false, err
	}

	branchInfo := strings.Split(out, "\n")[0]
	trackedReg := regexp.MustCompile(`## [^.]+\.\.\.[^/]+/.+`)

	return trackedReg.MatchString(branchInfo), nil
}

func isDetachedHead(dir string, quiet bool) (bool, error) {
	branch, err := getCurrentBranch(dir, quiet)
	if err != nil {
		return false, err
	}

	return strings.Contains(branch, "detached"), nil
}

func isDirty(dir string, quiet bool) (bool, error) {
	out, err := execute(dir, quiet, "status", "--porcelain")
	if err != nil {
		return false, err
	}

	return out != "", nil
}

func isPushed(dir string, quiet bool) (bool, error) {
	out, err := execute(dir, quiet, "log", "@{push}..")
	if err != nil {
		return false, err
	}

	return out == "", nil
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
