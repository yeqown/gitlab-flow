package types

import (
	"github.com/pkg/errors"
)

// OAuth2Mode represents the mode of OAuth2 authorization, there are two modes:
// 1. open browser to authorize automatically, then the gitlab server will redirect to callbackURI.
// 2. print a link to authorize, user can copy and paste to browser to authorize, and then the gitlab server
// will redirect to callbackURI as well, but user need to copy the code from browser to terminal.
//   - this mode is useful when the application is running in a headless environment.
type OAuth2Mode int

func (m OAuth2Mode) String() string {
	switch m {
	case OAuth2Mode_Auto:
		return "auto"
	case OAuth2Mode_Manual:
		return "manual"
	default:
		return "unknown"
	}
}

const (
	// OAuth2Mode_Auto means that the application will open a browser to authorize automatically.
	// This mode is default.
	OAuth2Mode_Auto OAuth2Mode = 1

	// OAuth2Mode_Manual means that the application will print a link to authorize,
	// user can copy and paste to browser to authorize.
	OAuth2Mode_Manual OAuth2Mode = 2
)

type OAuth struct {
	Scopes       string     `toml:"scopes"`
	CallbackHost string     `toml:"callback_host"` // Notice: callback host for oauth2 without scheme
	AccessToken  string     `toml:"access_token"`
	RefreshToken string     `toml:"refresh_token"`
	AppID        string     `toml:"app_id"`
	AppSecret    string     `toml:"app_secret"`
	Mode         OAuth2Mode `toml:"mode"`
}

// BranchSetting contains some personal setting of git branch.
type BranchSetting struct {
	Master, Dev, Test BranchTyp

	FeatureBranchPrefix         string `toml:"feature_branch_prefix"`
	HotfixBranchPrefix          string `toml:"hotfix_branch_prefix"`
	ConflictResolveBranchPrefix string `toml:"conflict_resolve_branch_prefix"`
	IssueBranchPrefix           string `toml:"issue_branch_prefix"`
}

var (
	errEmptyBranch    = errors.New("invalid branch setting")
	errEmptyOAuth     = errors.New("invalid gitlab OAuth setting")
	errEmptyGitlabAPI = errors.New("empty gitlab API/HOST URL")
)

type ConfigType string

const (
	ConfigType_Global  ConfigType = "global"
	ConfigType_Project ConfigType = "project"
)

type ConfigHolder interface {
	Type() ConfigType

	AsGlobal() *Config
	AsProject() *ProjectConfig
	ValidateConfig() error
}

// Config contains all fields can be specified by user.
type Config struct {
	OAuth2       *OAuth         `toml:"oauth"`
	Branch       *BranchSetting `toml:"branch"`
	GitlabAPIURL string         `toml:"gitlab_api_url"`
	GitlabHost   string         `toml:"gitlab_host"`
	DebugMode    bool           `toml:"debug"`
	OpenBrowser  bool           `toml:"open_browser"`
}

func (c *Config) Type() ConfigType {
	return ConfigType_Global
}

func (c *Config) AsGlobal() *Config {
	return c
}

func (c *Config) AsProject() *ProjectConfig {
	panic("global config should not be converted to project config")
}

func (c *Config) ValidateConfig() error {
	// check branch settings first.
	if c.Branch == nil {
		return errEmptyBranch
	}

	if c.Branch.Master == "" || c.Branch.Dev == "" || c.Branch.Test == "" ||
		c.Branch.FeatureBranchPrefix == "" || c.Branch.HotfixBranchPrefix == "" ||
		c.Branch.ConflictResolveBranchPrefix == "" || c.Branch.IssueBranchPrefix == "" {
		return errors.Wrap(errEmptyBranch, "some branch(s) is not set")
	}

	// if global configuration, check OAuth2 settings and GitlabAPIURL, GitlabHost.
	if c.GitlabAPIURL == "" || c.GitlabHost == "" {
		return errEmptyGitlabAPI
	}

	if c.OAuth2 == nil || c.OAuth2.Scopes == "" || c.OAuth2.CallbackHost == "" {
		return errEmptyOAuth
	}

	return nil
}

// ProjectConfig contains some fields can be specified by user,
// but they have higher priority than global config.
type ProjectConfig struct {
	// ProjectName is the name of project, it is used to identify the project.
	ProjectName string `toml:"project_name,omitempty"`

	// The following fields are not necessary to be specified by user,
	// but they are used to store the project information.
	// They have higher priority than global config.
	Branch      *BranchSetting `toml:"branch,omitempty"`
	DebugMode   *bool          `toml:"debug,omitempty"`
	OpenBrowser *bool          `toml:"open_browser,omitempty"`
}

func (c *ProjectConfig) Type() ConfigType {
	return ConfigType_Project
}

func (c *ProjectConfig) AsGlobal() *Config {
	panic("project config should not be converted to global config")
}

func (c *ProjectConfig) AsProject() *ProjectConfig {
	return c
}

func (c *ProjectConfig) ValidateConfig() error {
	// check branch settings first.
	if c.Branch == nil {
		return errEmptyBranch
	}

	if c.Branch.Master == "" || c.Branch.Dev == "" || c.Branch.Test == "" ||
		c.Branch.FeatureBranchPrefix == "" || c.Branch.HotfixBranchPrefix == "" ||
		c.Branch.ConflictResolveBranchPrefix == "" || c.Branch.IssueBranchPrefix == "" {
		return errors.Wrap(errEmptyBranch, "some branch(s) is not set")
	}

	return nil
}
