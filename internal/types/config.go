package types

import "errors"

var (
	ErrEmptyAccessToken = errors.New("empty access token")
	ErrEmptyGitlabAPI   = errors.New("empty gitlab API URL")
)

// Config contains all fields can be specified by user.
type Config struct {
	AccessToken  string `toml:"access_token"`
	DebugMode    bool   `toml:"debug"`
	GitlabAPIURL string `toml:"gitlab_api_url"`
	OpenBrowser  bool   `toml:"open_browser"`
}

// Apply open debug in Config if debug is true, otherwise do nothing.
func (cfg *Config) Apply(debug, openBrowser bool) *Config {
	if debug {
		cfg.DebugMode = debug
	}

	if openBrowser {
		cfg.OpenBrowser = openBrowser
	}

	return cfg
}

// Valid validates config is valid to use.
func (cfg Config) Valid() error {
	if cfg.AccessToken == "" {
		return ErrEmptyAccessToken
	}

	if cfg.GitlabAPIURL == "" {
		return ErrEmptyGitlabAPI
	}

	return nil
}
