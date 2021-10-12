package types

import "errors"

var (
	ErrEmptyAccessToken = errors.New("empty access token")
	ErrEmptyGitlabAPI   = errors.New("empty gitlab API URL")
)

type OAuth struct {
	AccessToken  string `toml:"access_token"`
	RefreshToken string `toml:"refresh_token"`
}

// Config contains all fields can be specified by user.
type Config struct {
	OAuth *OAuth

	// Deprecated: use oauth as instead
	//AccessToken string `toml:"access_token"`

	GitlabAPIURL string `toml:"gitlab_api_url"`
	GitlabHost   string `toml:"gitlab_host"`
	DebugMode    bool   `toml:"debug"`
	OpenBrowser  bool   `toml:"open_browser"`
}

//// Apply open debug in Config if debug is true, otherwise do nothing.
//// Deprecated
//func (cfg *Config) Apply(debug, openBrowser bool) *Config {
//	if debug {
//		cfg.DebugMode = debug
//	}
//
//	if openBrowser {
//		cfg.OpenBrowser = openBrowser
//	}
//
//	return cfg
//}

// Valid validates config is valid to use.
func (cfg Config) Valid() error {
	//if cfg.AccessToken == "" {
	//	return ErrEmptyAccessToken
	//}

	if cfg.GitlabAPIURL == "" {
		return ErrEmptyGitlabAPI
	}

	return nil
}
