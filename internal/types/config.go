package types

import "errors"

var (
	ErrEmptyAccessToken = errors.New("empty access token")
	ErrEmptyGitlabAPI   = errors.New("empty gitlab API URL")
)

type OAuth struct {
	Scopes       string `toml:"scopes"`
	CallbackHost string `toml:"callback_host"` // Notice: callback host for oauth2 without scheme
	AccessToken  string `toml:"access_token"`
	RefreshToken string `toml:"refresh_token"`
}

// BranchSetting contains some personal setting of git branch.
type BranchSetting struct {
	Master, Dev, Test BranchTyp

	FeatureBranchPrefix         string `toml:"feature_branch_prefix"`
	HotfixBranchPrefix          string `toml:"hotfix_branch_prefix"`
	ConflictResolveBranchPrefix string `toml:"conflict_resolve_branch_prefix"`
	IssueBranchPrefix           string `toml:"issue_branch_prefix"`
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

// Valid validates config is valid to use.
func (cfg Config) Valid() error {
	if cfg.GitlabAPIURL == "" {
		return ErrEmptyGitlabAPI
	}

	return nil
}
