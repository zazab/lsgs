package main

import (
	"regexp"
	"strconv"

	"github.com/docopt/docopt-go"
	"github.com/reconquest/ser-go"
)

//noinspection ProblematicWhitespace
var (
	version       = "1.0"
	versionString = "lsgs " + version
	usage         = versionString + `

Usage:
    lsgs [<path>...] [options]
    lsgs -R [<path>...] [options]
    lsgs -B [<path>...] [options]

Options:
    -R                   Checks if repo is pushed to origin. Repo is marked as
                         dirity if repo is not in detached HEAD state and if:
                          * branch has not pushed commits
                          * repo is in dirty state (marked as "(dirty)")
                          * current branch has no tracking information
                            (marked as "(not tracked)")
    -B                   Show repo branch. Repo marked as dirty if branch
                         differs from master.
    <path>               Path to working tree, which you want to list status.
                         Supports multiple paths. [default: .]
    --max-depth <level>  Maximum recursion depth [default: 1]
    -r                   Alias for --max-depth 7
    -x <pattern>         Exlude directories matching pattern
    -d --dirty           Show only dirty repos
    -q --quiet           Be quiet
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, versionString, false, true)
	if err != nil {
		panic(err)
	}

	var (
		maxDepth, _    = strconv.Atoi(args["--max-depth"].(string))
		workingDirs, _ = args["<path>"].([]string)

		remote = args["-R"].(bool)
		branch = args["-B"].(bool)

		exclude, _ = args["-x"].(string)
		dirRegexp *regexp.Regexp

		onlyDirty = args["--dirty"].(bool)
		quiet     = args["--quiet"].(bool)
		recursive = args["-r"].(bool)
	)

	if recursive && maxDepth == 1 {
		maxDepth = 7
	}

	if len(workingDirs) == 0 {
		workingDirs = []string{"."}
	}

	if exclude != "" {
		dirRegexp, err = regexp.Compile(exclude)
		if err != nil {
			logger.Fatal(ser.Errorf(err, "can't compile regexp '%s'", exclude))
		}
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

	for _, workingDir := range workingDirs {
		path, err := newPath(workingDir)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = walkDirs(path, 1, maxDepth, onlyDirty, quiet, checker, dirRegexp)
		if err != nil {
			logger.Error(err)
		}
	}
}
