package main

import (
	"time"

	"github.com/fatih/color"
	"github.com/reconquest/hierr-go"
)

var (
	red    = color.New(color.FgHiRed).SprintFunc()
	green  = color.New(color.FgHiGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	yellow = color.New(color.FgHiYellow).SprintFunc()

	dirtyMarker     = red("✗")
	unpushedMarker  = yellow("↑")
	cleanMarker     = green("✓")
	emptyMarker     = blue("∅")
	untrackedMarker = yellow("?")
)

type checkFunc func(path, bool, bool) error

func newDateChecker(offset time.Duration) func(path, bool, bool) error {
	threshold := time.Now().Add(-offset)

	return func(path path, onlyDirty, quiet bool) error {
		empty, err := isEmpty(path.linkTo, quiet)
		if err != nil {
			return err
		}

		if empty {
			printEmpty(path, "(empty)")
			return nil
		}

		commitDate, err := lastCommitDate(path.linkTo, quiet)
		if err != nil {
			return err
		}

		if commitDate.After(threshold) {
			printDirty(path, commitDate.Format(time.RFC3339))
		} else {
			printClean(path, onlyDirty, commitDate.Format(time.RFC3339))
		}

		return nil
	}
}

func pushCheck(path path, onlyDirty, quiet bool) error {
	empty, err := isEmpty(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if empty {
		printEmpty(path, "(empty)")
		return nil
	}

	detached, err := isDetachedHead(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if detached {
		printClean(path, onlyDirty)
		return nil
	}

	tracked, err := isTracked(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if !tracked {
		printDirty(path, "(not tracked)")
		return nil
	}

	untracked, err := isUntracked(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if untracked {
		printUntracked(path, "(untracked)")
		return nil
	}

	dirty, err := isDirty(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if dirty {
		printDirty(path, "(dirty)")
		return nil
	}

	pushed, err := isPushed(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if pushed {
		printClean(path, onlyDirty)
	} else {
		printUnpushed(path)
	}

	return nil
}

func dirtyCheck(path path, onlyDirty, quiet bool) error {
	dirty, err := isDirty(path.linkTo, quiet)
	if err != nil {
		return err
	}

	if dirty {
		printDirty(path)
	} else {
		printClean(path, onlyDirty)
	}

	return nil
}

func showBranch(path path, onlyDirty, quiet bool) error {
	branch, err := getCurrentBranch(path.linkTo, quiet)
	if err != nil {
		return hierr.Errorf(err, "error processing %s", path.path)
	}

	if branch == "master" {
		printClean(path, onlyDirty, branch)
	} else {
		printDirty(path, branch)
	}

	return nil
}
