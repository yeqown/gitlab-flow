package types

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
