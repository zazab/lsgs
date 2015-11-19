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

	"github.com/docopt/docopt-go"
	"github.com/fatih/color"
)

var (
	version       string = "0.1"
	versionString string = "lsgs " + version
	usage         string = versionString + `

Usage:
	lsgs [<path>] [options]

Options:
	<path>  Path to working tree, which you want to list status
	--max-depth <level>  maximum recursion depth [default: 1]
`

	red   = color.New(color.FgHiRed).SprintFunc()
	green = color.New(color.FgHiBlue).SprintFunc()
)

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

	path, err := newPath(workingDir)
	if err != nil {
		log.Fatal(err)
	}
	err = listGitStatuses(path, 1, maxDepth)
	if err != nil {
		log.Fatal(err)
	}
}

func listGitStatuses(path Path, depth, maxDepth int) error {
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
			fmt.Printf("  %s\n", path)
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
				err := listGitStatuses(
					filePath, depth+1, maxDepth,
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
			fmt.Printf("  %s\n", path)
		}
		return nil
	case err == nil:
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = path.linkTo
		cmd.Stderr = os.Stderr

		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error running 'git status': %s", err)
		}

		if len(out) > 0 {
			fmt.Printf("%s %s\n", red("✗"), path)
		} else {
			fmt.Printf("%s %s\n", green("•"), path)
		}
		return nil
	default:
		return fmt.Errorf(
			"can't stat '%s': %s", filepath.Join(path.path, ".git"),
		)
	}

}
