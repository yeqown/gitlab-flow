package gitop

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/yeqown/log"
)

var _ IGitOperator = &operatorBasedCmd{}

// operatorBasedCmd implements IGitOperator based local `git` executable file.
type operatorBasedCmd struct {
	// cmd indicates which executable command to use.
	cmd string
	// dir is the directory of git repository in local filesystem.
	dir string
	// verbose mode indicates print more information into os.Stdout
	verbose bool

	// commands
	fetchCmd         string // fetch command
	checkoutCmd      string // checkout command
	currentBranchCmd string
	mergeCmd         string
}

// NewBasedCmd generate a git operator based command line.
func NewBasedCmd(dir string) IGitOperator {
	return operatorBasedCmd{
		cmd:              "git",
		dir:              dir,
		verbose:          true,
		fetchCmd:         "fetch {arg}",
		checkoutCmd:      "checkout {createFlag}{branch}",
		currentBranchCmd: "rev-parse --abbrev-ref HEAD",
		mergeCmd:         "merge --no-ff {branch}",
	}
}

// reference to https://golang.org/x/tools/go/vcs
func (c operatorBasedCmd) run(dir string, cmd string, keyval ...string) error {
	_, err := c.run1(dir, cmd, keyval)
	return err
}

func (c operatorBasedCmd) run1(dir string, cmdline string, keyvalPairs []string) ([]byte, error) {
	m := make(map[string]string)
	for i := 0; i < len(keyvalPairs); i += 2 {
		m[keyvalPairs[i]] = keyvalPairs[i+1]
	}
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = expand(m, arg)
	}

	_, err := exec.LookPath(c.cmd)
	if err != nil {
		log.Errorf("go: missing %s command.", c.cmd)
		return nil, err
	}

	cmd := exec.Command(c.cmd, args...)
	cmd.Dir = dir
	cmd.Env = envForDir(cmd.Dir)

	log.Debugf("cd %s", dir)
	log.Debugf("%s %s", c.cmd, strings.Join(args, " "))

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		log.
			WithFields(log.Fields{
				"dir":   dir,
				"cmd":   fmt.Sprintf("%s %s", c.cmd, strings.Join(args, " ")),
				"out":   string(out),
				"error": err,
			}).
			Error("command execute failed")

		return out, err
	}

	return out, nil
}

// Checkout local branch and control whether create a new branch or not.
// DONE(@yeqown) use `b` parameter.
func (c operatorBasedCmd) Checkout(branchName string, create bool) error {
	createFlag := ""
	if create {
		createFlag = "-b"
	}
	return c.run(c.dir, c.checkoutCmd, "createFlag", createFlag, "branch", branchName)
}

// FetchOrigin fetch origin branched.
func (c operatorBasedCmd) FetchOrigin() error {
	return c.run(c.dir, c.fetchCmd, "arg", "--all")
}

// CurrentBranch get current repository info, includes:
// * current branch name
// * repository name ?
func (c operatorBasedCmd) CurrentBranch() (string, error) {
	branchName, err := c.run1(c.dir, c.currentBranchCmd, nil)
	if err != nil {
		return "", errors.Wrapf(err, "get current branch name failed")
	}

	return strings.TrimSpace(string(branchName)), nil
}

// Merge would merge source into target with --no-ff flag.
// if current branch is not target branch, it checkouts to target automatically.
// FIXED(@yeqown): conflict output should be formatted.
func (c operatorBasedCmd) Merge(source, target string) error {
	if source == "" || target == "" {
		return errors.New("invalid branch parameter of Merge")
	}

	b, err := c.CurrentBranch()
	if err != nil {
		return errors.Wrapf(err, "Merge => c.CurrentBranch() failed")
	}
	if b == "" {
		return errors.New("Merge => c.CurrentBranch() failed: empty branch got")
	}

	if strings.Compare(b, target) != 0 {
		if err = c.Checkout(target, false); err != nil {
			return errors.Wrap(err, "automatic checkout failed")
		}
	}

	pairs := []string{"branch", source}
	output, err := c.run1(c.dir, c.mergeCmd, pairs)
	if len(output) != 0 {
		title := fmt.Sprintf("\nMerge Output (%s => %s):\n", source, target)
		_, _ = fmt.Fprintf(os.Stdout, title+string(output))
	}

	return err
}

// expand rewrites s to replace {k} with match[k] for each key k in match.
func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

// envForDir returns a copy of the environment
// suitable for running in the given directory.
// The environment is the current process's environment
// but with an updated $PWD, so that an os.Getwd in the
// child will be faster.
func envForDir(dir string) []string {
	env := os.Environ()
	// Internally we only use rooted paths, so dir is rooted.
	// Even if dir is not rooted, no harm done.
	return mergeEnvLists([]string{"PWD=" + dir}, env)
}

// mergeEnvLists merges the two environment lists such that
// variables with the same name in "in" replace those in "out".
func mergeEnvLists(in, out []string) []string {
NextVar:
	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue NextVar
			}
		}
		out = append(out, inkv)
	}
	return out
}
