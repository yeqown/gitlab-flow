package types

import (
	"os"
	"path"

	"github.com/yeqown/log"
)

// ProjectBasics contains basic attributes of project.
type ProjectBasics struct {
	ID     int
	Name   string
	WebURL string
}

// FlowContext contains all necessary parameters of flow command to execute.
// SHOULD NOT export context attributes to another package.
type FlowContext struct {
	oauth *OAuth
	// project of current working directory, normally,
	// get from current working directory.
	project *ProjectBasics
	// cwd represents current working directory.
	cwd string
	// gitlabAPIUrl represents gitlab API endpoint.
	gitlabAPIUrl string
	// projectName the actual name of project.
	projectName string
	// confPath of configuration file path.
	confPath string
	// forceRemote force choose load project from remote(gitlab) rather than local.
	forceRemote bool
	// debug indicates whether gitlab-flow print more detail logs.
	debug       bool
	openBrowser bool
}

// NewContext be generated with non project information.
// Do not use Project directly!!!
func NewContext(cwd, confPath, projectName string, c *Config, forceRemote bool) *FlowContext {
	if cwd == "" {
		panic("cwd could not be empty")
	}

	ctx := &FlowContext{
		oauth:        c.OAuth2,
		gitlabAPIUrl: c.GitlabAPIURL,
		cwd:          cwd,
		project:      nil,
		projectName:  "",
		confPath:     "",
		forceRemote:  forceRemote,
		debug:        c.DebugMode,
		openBrowser:  c.OpenBrowser,
	}

	ctx.applyConfPath(confPath)
	ctx.applyProjectName(projectName)

	return ctx
}

func (c *FlowContext) GetOAuth() *OAuth {
	if c == nil {
		return &OAuth{}
	}
	return c.oauth
}

func (c *FlowContext) InjectProject(p *ProjectBasics) {
	c.project = p
}

func (c *FlowContext) Project() *ProjectBasics {
	if c == nil {
		return &ProjectBasics{}
	}

	return c.project
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

func (c *FlowContext) CWD() string {
	if c == nil {
		return ""
	}

	return c.cwd
}

func (c *FlowContext) IsDebug() bool {
	if c == nil {
		return false
	}

	return c.debug
}

func (c *FlowContext) ShouldOpenBrowser() bool {
	if c == nil {
		return false
	}

	return c.openBrowser
}

func (c *FlowContext) APIEndpoint() string {
	if c == nil {
		return ""
	}

	return c.gitlabAPIUrl
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

// applyProjectName to judge which project name should be taken.
func (c *FlowContext) applyProjectName(projectName string) {
	if projectName != "" {
		c.projectName = projectName
		return
	}

	c.projectName = path.Base(c.cwd)
	return
}

// OpFeatureContext contains all parameters of features' operations in common.
type OpFeatureContext struct {
	// ForceCreateMergeRequest if this is true, means merge request would be create no matter whether
	// merge request has been created or merged.
	ForceCreateMergeRequest bool
	// FeatureBranchName specify which branch name to use in the lifecycle of feature operations.
	FeatureBranchName string

	// ParseIssueCompatible if this is true, means parse issueName to feature in compatible way.
	ParseIssueCompatible bool
}
