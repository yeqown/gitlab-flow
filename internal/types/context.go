package types

import (
	"path/filepath"

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
	mergedConfig *Config

	// project of the current working directory, normally, get from the current working directory.
	project *ProjectBasics
	// cwd represents the current working directory.
	cwd string
	// projectName the actual name of a project.
	projectName string

	// forceRemote force chooses load project from remote(gitlab) rather than local.
	forceRemote bool
	// debug indicates whether gitlab-flow prints more detail logs.
	debug       bool
	openBrowser bool
}

// NewContext be generated with non-project information.
// Do not use Project directly!!!
func NewContext(cwd, projectName string, c *Config, forceRemote bool) *FlowContext {
	if cwd == "" {
		panic("cwd could not be empty")
	}

	ctx := &FlowContext{
		// oauth:        c.OAuth2,
		// gitlabAPIUrl: c.GitlabAPIURL,
		mergedConfig: c,
		cwd:          cwd,
		project:      nil, // set later by InjectProject
		projectName:  "",  // set later by applyProjectName
		forceRemote:  forceRemote,
		debug:        c.DebugMode,
		openBrowser:  c.OpenBrowser,
	}

	ctx.applyProjectName(projectName)

	return ctx
}

func (c *FlowContext) Config() *Config {
	if c == nil || c.mergedConfig == nil {
		log.Fatal("invalid context, mergedConfig is nil")
	}

	return c.mergedConfig
}

func (c *FlowContext) GetOAuth() *OAuth {
	if c == nil || c.mergedConfig == nil || c.mergedConfig.OAuth2 == nil {
		return &OAuth{}
	}
	return c.mergedConfig.OAuth2
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
	if c == nil || c.mergedConfig == nil {
		return ""
	}

	return c.mergedConfig.GitlabAPIURL
}

// applyProjectName to judge which project name should be taken.
func (c *FlowContext) applyProjectName(projectName string) {
	if projectName != "" {
		c.projectName = projectName
		return
	}

	// FIXED: this split file path to get project name is not correct.
	// we used path.Base to get project name before, but it's not correct.
	c.projectName = filepath.Base(c.cwd)

	return
}

// OpFeatureContext contains all parameters of features' operations in common.
type OpFeatureContext struct {
	// ForceCreateMergeRequest if this is true, means merge request would be create no matter whether
	// merge request has been created or merged.
	ForceCreateMergeRequest bool
	// FeatureBranchName specify which branch name to use in the lifecycle of feature operations.
	FeatureBranchName string
	// AutoMergeRequest if this is true, means merge request would be merged automatically.
	AutoMergeRequest bool

	// ParseIssueCompatible if this is true, means parse issueName to feature in compatible way.
	ParseIssueCompatible bool
}

type OpHotfixContext struct {
	// ForceCreateMergeRequest if this is true, means merge request would be create no matter whether
	// merge request has been created or merged.
	ForceCreateMergeRequest bool
}
