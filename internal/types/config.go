package types

import "errors"

var (
	ErrEmptyAccessToken = errors.New("empty access token")
	ErrEmptyGitlabAPI   = errors.New("empty gitlab API URL")
)

//// GitlabUser contains necessary information of current gitlab user.
//type GitlabUser struct {
//	ID        int    `toml:"id"`
//	UserName  string `toml:"user_name"`
//	Email     string `toml:"email"`
//	AvatarURL string `toml:"avatar_url"`
//}

// Config contains all fields can be specified by user.
type Config struct {
	AccessToken  string `toml:"access_token"`
	DebugMode    bool   `toml:"debug"`
	GitlabAPIURL string `toml:"gitlab_api_url"`
}

func (cfg Config) Valid() error {
	if cfg.AccessToken == "" {
		return ErrEmptyAccessToken
	}

	if cfg.GitlabAPIURL == "" {
		return ErrEmptyGitlabAPI
	}

	return nil
}
