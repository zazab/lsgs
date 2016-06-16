# lsgs
If you have to work with tons of different git repositories, regulary switching
context from one to another, then lsgs will help to track it state with ease.

### Istallation
`go get github.com/zazab/lsgs`

### Usage
```
lsgs 1.0

Usage:
    lsgs [<path>...] [options]
    lsgs -R [<path>...] [options]
    lsgs -B [<path>...] [options]

Options:
    -R                   Checks if repo is pushed to origin. Repo is marked as
                         dirity if repo is not in detached HEAD state and if :
                          * branch has not pushed commits
                          * repo is in dirty state (marked as "(dirty)")
                          * current branch has no tracking information
                            (marked as "(not tracked)")
    -b                   Show repo branch. Repo marked as dirty if branch
                         differs from master.
    <path>               Path to working tree, which you want to list status.
                         Supports multiple paths. [default: .]
    --max-depth <level>  Maximum recursion depth [default: 1]
    -r                   Alias for --max-depth 7
    -d --dirty           Show only dirty repos
    -q --quiet           Be quiet
```
