package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/fatih/color"
)

var (
	version       string = "0.1"
	versionString string = "lsgs " + version
	usage         string = versionString + `

Usage:
	lsgs [<path>] [options]
	lsgs -R [<path>] [options]

Options:
	<path>               Path to working tree, which you want to list status
	--max-depth <level>  Maximum recursion depth [default: 1]
	-d --dirty           Show only dirty repos
	-R --remote          Checks if repo is pushed to origin
	-q --quiet           Be quiet
	-b --branch          Show repo branch
`

	red   = color.New(color.FgHiRed).SprintFunc()
	green = color.New(color.FgHiBlue).SprintFunc()
)

type (
	checkFunc func(Path, bool, bool) error
)

func stdCheck(
	path Path,
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

func pushCheck(path Path, onlyDirty, quiet bool) error {
	return stdCheck(path, onlyDirty, quiet, "log", "@{push}..")
}

func dirtyCheck(path Path, onlyDirty, quiet bool) error {
	return stdCheck(path, onlyDirty, quiet, "status", "--porcelain")
}

func showBranch(path Path, onlyDirty, quiet bool) error {
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
			currentBranch = strings.TrimLeft(string(branch), "* ")
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

type Path struct {
	path   string
	linkTo string
}

func (path Path) String() string {
	if path.linkTo != path.path {
		return fmt.Sprintf("%s -> %s", path.path, path.linkTo)
	}
	return path.path
}

func newPath(path string) (Path, error) {
	linkTo, err := filepath.EvalSymlinks(path)
	if err != nil {
		return Path{}, err
	}

	return Path{path, linkTo}, nil
}

func main() {
	args, err := docopt.Parse(usage, nil, true, versionString, false, true)
	if err != nil {
		panic(err)
	}

	var (
		workingDir string = "."
		maxDepth   int

		remote = args["--remote"].(bool)
		branch = args["--branch"].(bool)
		quiet  = args["--quiet"].(bool)
	)

	if args["<path>"] != nil {
		workingDir = args["<path>"].(string)
	}

	if args["--max-depth"] != nil {
		maxDepth, err = strconv.Atoi(args["--max-depth"].(string))
		if err != nil {
			log.Fatal("can't convert max-depth to int:", err)
		}
	}

	onlyDirty := args["--dirty"].(bool)

	path, err := newPath(workingDir)
	if err != nil {
		log.Fatal(err)
	}

	var checker checkFunc
	switch {
	case remote:
		checker = pushCheck
	case branch:
		checker = showBranch
	default:
		checker = dirtyCheck
	}

	err = walkDirs(path, 1, maxDepth, onlyDirty, quiet, checker)
	if err != nil {
		log.Fatal(err)
	}
}

func walkDirs(
	path Path, depth, maxDepth int, onlyDirty, quiet bool, checker checkFunc,
) error {
	info, err := os.Stat(path.linkTo)
	if err != nil {
		return fmt.Errorf("can't stat '%s': %s", path, err)
	}

	if !info.IsDir() {
		return nil
	}

	info, err = os.Stat(filepath.Join(path.linkTo, ".git"))
	switch {
	case os.IsNotExist(err): // not a git repo
		if depth > maxDepth {
			if !onlyDirty {
				fmt.Printf("  %s\n", path)
			}
			return nil
		}

		files, err := ioutil.ReadDir(path.path)
		if err != nil {
			return fmt.Errorf("can't read dir '%s': %s", path, err)
		}

		failed := false
		goneDeeper := false
		for _, file := range files {
			filePath, err := newPath(filepath.Join(path.path, file.Name()))
			if err != nil {
				failed = true
				log.Println(err)
				continue
			}
			info, err := os.Stat(filePath.linkTo)
			if err != nil {
				failed = true
				log.Println(err)
				continue
			}
			if info.IsDir() {
				err := walkDirs(
					filePath, depth+1, maxDepth, onlyDirty, quiet, checker,
				)
				if err != nil {
					failed = true
					log.Println(err)
				}
			}
		}
		if failed {
			return errors.New("errors occured")
		}
		if !goneDeeper {
			if !onlyDirty {
				fmt.Printf("  %s\n", path)
			}
		}
		return nil
	case err == nil:
		return checker(path, onlyDirty, quiet)
	default:
		return fmt.Errorf(
			"can't stat '%s': %s", filepath.Join(path.path, ".git"),
		)
	}

}
