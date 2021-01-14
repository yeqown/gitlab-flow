package gitop

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/yeqown/log"
)

var _ IGitOperator = &operatorBasedCmd{}

// 使用本地的 git 客户端完成操作
type operatorBasedCmd struct {
	Cmd              string // git command
	Dir              string // repo dir
	fetchCmd         string // fetch command
	checkoutCmd      string // checkout command
	currentBranchCmd string
}

// NewBasedCmd generate a git operator based command line.
func NewBasedCmd(dir string) IGitOperator {
	return operatorBasedCmd{
		Cmd:              "git",
		Dir:              dir,
		fetchCmd:         "fetch {arg}",
		checkoutCmd:      "checkout {branch}",
		currentBranchCmd: "rev-parse --abbrev-ref HEAD",
	}
}

// reference to https://golang.org/x/tools/go/vcs
func (c operatorBasedCmd) run(dir string, cmd string, keyval ...string) error {
	_, err := c.run1(dir, cmd, keyval, true)
	return err
}

func (c operatorBasedCmd) run1(dir string, cmdline string, keyval []string, verbose bool) ([]byte, error) {
	m := make(map[string]string)
	for i := 0; i < len(keyval); i += 2 {
		m[keyval[i]] = keyval[i+1]
	}
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = expand(m, arg)
	}

	_, err := exec.LookPath(c.Cmd)
	if err != nil {
		log.Errorf("go: missing %s command.", c.Cmd)
		return nil, err
	}

	cmd := exec.Command(c.Cmd, args...)
	cmd.Dir = dir
	cmd.Env = envForDir(cmd.Dir)

	log.Debugf("cd %s", dir)
	log.Debugf("%s %s", c.Cmd, strings.Join(args, " "))

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		if verbose {
			log.Errorf("# cd %s; %s %s", dir, c.Cmd, strings.Join(args, " "))
			log.Errorf("%s", out)
		}
		return nil, err
	}
	return out, nil
}

// Checkout local branch and control whether create a new branch or not.
// TODO(@yeqown) use `b` parameter.
func (c operatorBasedCmd) Checkout(branchName string, b bool) error {
	return c.run(c.Dir, c.checkoutCmd, "branch", branchName)
}

// FetchOrigin fetch origin branched.
func (c operatorBasedCmd) FetchOrigin() error {
	return c.run(c.Dir, c.fetchCmd, "arg", "--all")
}

// Current get current repository info, includes:
// * current branch name
// * repository name ?
func (c operatorBasedCmd) CurrentBranch() (string, error) {
	branchName, err := c.run1(c.Dir, c.currentBranchCmd, nil, false)
	if err != nil {
		return "", errors.Wrapf(err, "get current branch name failed")
	}

	return strings.TrimSpace(string(branchName)), nil
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
