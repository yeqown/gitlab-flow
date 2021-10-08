package main

import (
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/AlecAivazis/survey/v2"
)

// surveyConfig
// DONE(@yeqown): init flow2 in survey method.
// FIXME(@yeqown): make appId and appSecret safety. use build ldflags?
func surveyConfig(cfg *types.Config) error {
	log.
		WithField("config", cfg).
		Debug("surveyConfig called")

	if cfg.OAuth == nil {
		cfg.OAuth = new(types.OAuth)
	}

	questions := []*survey.Question{
		{
			Name: "appId",
			Prompt: &survey.Input{
				Message: "Input your application client ID",
				Default: cfg.OAuth.AppID,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "appSecret",
			Prompt: &survey.Input{
				Message: "Input your application client secret",
				Default: cfg.OAuth.AppSecret,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "host",
			Prompt: &survey.Input{
				Message: "Input your gitlab server host",
				Default: cfg.GitlabHost,
				Help:    "such as: https://gitlab.example.com",
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "apiURL",
			Prompt: &survey.Input{
				Message: "Input your gitlab api url",
				Default: cfg.GitlabAPIURL,
				Help:    "such as: https://gitlab.example.com/api/v4/",
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "debugMode",
			Prompt: &survey.Confirm{
				Message: "Would you like to use gitlab in debug mode?",
				Default: cfg.DebugMode,
			},
			Validate:  nil,
			Transform: nil,
		},
		{
			Name: "openBrowser",
			Prompt: &survey.Confirm{
				Message: "Would you let gitlab-flow open browser automatically when needed",
				Default: cfg.OpenBrowser,
			},
			Validate:  nil,
			Transform: nil,
		},
	}

	ans := new(answer)
	err := survey.Ask(questions, ans)
	if err == nil {
		log.
			WithField("answer", ans).
			Debug("surveyConfig done")
	}

	cfg.DebugMode = ans.DebugMode
	cfg.OpenBrowser = ans.OpenBrowser
	cfg.GitlabHost = ans.Host
	cfg.GitlabAPIURL = ans.APIURL
	cfg.OAuth.AppID = ans.AppID
	cfg.OAuth.AppSecret = ans.AppSecret

	return err
}

type answer struct {
	AppID, AppSecret, Host, APIURL string
	OpenBrowser                    bool
	DebugMode                      bool
}
