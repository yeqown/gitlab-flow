package types

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

// Config contains all fields can be specified by user.
type Config struct {
	OAuth2       *OAuth         `toml:"oauth"`
	Branch       *BranchSetting `toml:"branch"`
	GitlabAPIURL string         `toml:"gitlab_api_url"`
	GitlabHost   string         `toml:"gitlab_host"`
	DebugMode    bool           `toml:"debug"`
	OpenBrowser  bool           `toml:"open_browser"`
}
