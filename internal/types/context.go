package types

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yeqown/log"
)

// ProjectBasics contains basic attributes of project.
type ProjectBasics struct {
	ID     int
	Name   string
	WebURL string
}

// FlowContext contains all necessary parameters of flow command to execute.
type FlowContext struct {
	// Conf of flow CLI.
	Conf *Config

	// CWD current working directory.
	CWD string

	// Project of current working directory, normally,
	// get from current working directory.
	Project *ProjectBasics

	// projectName the actual name of project.
	projectName string

	// confPath of configuration file path.
	confPath string

	// forceRemote force choose load project from remote(gitlab) rather than local.
	forceRemote bool
}

// NewContext be generated with non project information.
// Do not use Project directly!!!
func NewContext(cwd, confPath, projectName string, c *Config, forceRemote bool) *FlowContext {
	if cwd == "" {
		panic("cwd could not be empty")
	}

	ctx := &FlowContext{
		Conf:        c,
		CWD:         cwd,
		Project:     nil,
		forceRemote: forceRemote,
	}

	ctx.applyConfPath(confPath)
	ctx.applyProjectName(projectName)

	return ctx
}

// ProjectName return project name.
func (c *FlowContext) ProjectName() string {
	if c == nil {
		return ""
	}

	return c.projectName
}

// ConfPath return configuration path
func (c *FlowContext) ConfPath() string {
	if c == nil {
		return ""
	}

	return c.confPath
}

// ForceRemote return should module need to locate project by projectName from remote.
// true means locate project from remote, false means do not jump the process of
// locating project from local.
func (c *FlowContext) ForceRemote() bool {
	if c == nil {
		return false
	}

	return c.forceRemote
}

// applyConfPath to get configuration directory path rather than file path.
func (c *FlowContext) applyConfPath(confPath string) {
	fi, err := os.Stat(confPath)
	if err != nil {
		log.Errorf("could not stat confPath=%s", confPath)
		return
	}

	if fi.IsDir() {
		c.confPath = confPath
		return
	}

	c.confPath = path.Dir(confPath)
	return
}

// applyProjectName to judge which project name should be took.
func (c *FlowContext) applyProjectName(projectName string) {
	if projectName != "" {
		c.projectName = projectName
		return
	}

	c.projectName = extractProjectNameFromCWD(c.CWD)

	return
}

// extractProjectNameFromCWD get project name from current working directory.
// input:  /path/to/project
// output: 'project'
func extractProjectNameFromCWD(cwd string) string {
	arr := strings.Split(cwd, string(filepath.Separator))
	return arr[len(arr)-1]
}
