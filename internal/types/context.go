package types

import (
	"os"
	"path"

	"github.com/yeqown/log"
)

// FlowContext contains all necessary parameters of flow command to execute.
type FlowContext struct {
	// Conf of flow CLI.
	Conf *Config
	// CWD current working directory.
	CWD string
	// project of current working directory, normally, get from current working directory.
	Project *ProjectBasics

	//// Branch current branch name.
	//ProjectBranch string

	ConfPath string
}

// NewContext be generated with non project information.
// Do not use Project directly!!!
func NewContext(cwd, confPath string, c *Config) *FlowContext {
	ctx := &FlowContext{
		Conf:    c,
		CWD:     cwd,
		Project: nil,
	}

	ctx.calcConfPath(confPath)
	return ctx
}

func (c *FlowContext) calcConfPath(confPath string) {
	fi, err := os.Stat(confPath)
	if err != nil {
		log.Errorf("could not stat confPath=%s", confPath)
		return
	}

	if fi.IsDir() {
		c.ConfPath = confPath
		return
	}

	c.ConfPath = path.Dir(confPath)
	return
}

// ProjectBasics contains basic attributes of project.
type ProjectBasics struct {
	ID     int
	Name   string
	WebURL string
}
