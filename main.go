package main

import (
	"strconv"

	"github.com/docopt/docopt-go"
)

var (
	version       = "0.1"
	versionString = "lsgs " + version
	usage         = versionString + `

Usage:
	lsgs [<path>] [options]
	lsgs -R [<path>] [options]
	lsgs -b [<path>] [options]

Options:
	<path>               Path to working tree, which you want to list status [default: .]
	--max-depth <level>  Maximum recursion depth [default: 1]
	-d --dirty           Show only dirty repos
	-R --remote          Checks if repo is pushed to origin
	-q --quiet           Be quiet
	-b --branch          Show repo branch
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, versionString, false, true)
	if err != nil {
		panic(err)
	}

	var (
		maxDepth, _   = strconv.Atoi(args["--max-depth"].(string))
		workingDir, _ = args["<path>"].(string)

		onlyDirty = args["--dirty"].(bool)
		remote    = args["--remote"].(bool)
		branch    = args["--branch"].(bool)
		quiet     = args["--quiet"].(bool)
	)

	if workingDir == "" {
		workingDir = "."
	}

	path, err := newPath(workingDir)
	if err != nil {
		logger.Fatal(err)
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
		logger.Fatal(err)
	}
}
