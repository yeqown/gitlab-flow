package main

import (
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/AlecAivazis/survey/v2"
	"github.com/yeqown/log"
)

// surveyConfig initialize configuration in an interactive session.
// DONE(@yeqown): init flow2 in survey method.
func surveyConfig(cfg *types.Config) error {
	log.
		WithField("config", cfg).
		Debug("surveyConfig called")

	if cfg.OAuth == nil {
		cfg.OAuth = new(types.OAuth)
	}

	questions := []*survey.Question{
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

	return err
}

type answer struct {
	AppID, AppSecret, Host, APIURL string
	OpenBrowser                    bool
	DebugMode                      bool
}
