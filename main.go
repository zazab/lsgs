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

	err = listGitStatuses(workingDir, 1, maxDepth)
	if err != nil {
		log.Fatal(err)
	}
}

func listGitStatuses(path string, depth, maxDepth int) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("can't stat '%s': %s", path, err)
	}

	if !info.IsDir() {
		return nil
	}

	info, err = os.Stat(filepath.Join(path, ".git"))
	switch {
	case os.IsNotExist(err): // not a git repo
		if depth > maxDepth {
			fmt.Printf("%s\n", path)
			return nil
		}

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return fmt.Errorf("can't read dir '%s': %s", path, err)
		}

		failed := false
		for _, file := range files {
			if file.IsDir() {
				err := listGitStatuses(
					filepath.Join(path, file.Name()), depth+1, maxDepth,
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
		return nil
	case err == nil:
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = path
		cmd.Stderr = os.Stderr

		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error running 'git status': %s", err)
		}

		if len(out) > 0 {
			fmt.Printf("%s %s\n", path, red("✗"))
		} else {
			fmt.Printf("%s %s\n", path, green("•"))
		}
		return nil
	default:
		return fmt.Errorf("can't stat '%s': %s", filepath.Join(path, ".git"))
	}

}
