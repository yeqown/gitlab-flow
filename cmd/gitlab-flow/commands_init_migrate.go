package main

import (
	"net/url"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/types"
)

func buildGitlabQuestions(cfg *types.Config) []*survey.Question {
	return []*survey.Question{
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
			Name: "callbackHost",
			Prompt: &survey.Input{
				Message: "Input your callback host, DO NOT edit if you are not sure",
				Default: cfg.OAuth2.CallbackHost,
				Help:    "default: localhost:2333",
			},
			Validate: survey.Required,
		},
	}
}

func buildFlagsQuestions(cfg *types.Config) []*survey.Question {
	return []*survey.Question{
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
}

func buildBranchQuestions(cfg *types.Config) []*survey.Question {
	return []*survey.Question{
		{
			Name: "masterBranch",
			Prompt: &survey.Input{
				Message: "Input your master branch name",
				Default: string(cfg.Branch.Master),
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "devBranch",
			Prompt: &survey.Input{
				Message: "Input your dev branch name",
				Default: string(cfg.Branch.Dev),
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "testBranch",
			Prompt: &survey.Input{
				Message: "Input your test branch name",
				Default: string(cfg.Branch.Test),
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "featureBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your feature branch prefix",
				Default: cfg.Branch.FeatureBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "hotfixBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your hotfix branch prefix",
				Default: cfg.Branch.HotfixBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "conflictResolveBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your conflict resolve branch prefix",
				Default: cfg.Branch.ConflictResolveBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "issueBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your issue branch prefix",
				Default: cfg.Branch.IssueBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
	}
}

// surveyConfig initialize configuration in an interactive session.
// DONE(@yeqown): init flow2 in survey method.
func surveyConfig(cfg *types.Config) error {
	log.
		WithField("config", cfg).
		Debug("surveyConfig called")

	if cfg.OAuth2 == nil {
		cfg.OAuth2 = new(types.OAuth)
	}

	questions := make([]*survey.Question, 0, 8)
	questions = append(questions, buildGitlabQuestions(cfg)...)
	questions = append(questions, buildFlagsQuestions(cfg)...)
	questions = append(questions, buildBranchQuestions(cfg)...)

	ans := new(answer)
	err := survey.Ask(questions, ans)
	if err == nil {
		log.
			WithField("answer", ans).
			Debug("surveyConfig done")
	}

	u, err := url.Parse(ans.APIUrl)
	if err != nil {
		return errors.Wrap(err, "gitlab API URL is invalid")
	}
	cfg.GitlabAPIURL = ans.APIUrl
	// only save the scheme and host
	cfg.GitlabHost = u.Scheme + "://" + u.Host
	cfg.OAuth2.CallbackHost = ans.CallbackHost

	cfg.DebugMode = ans.DebugMode
	cfg.OpenBrowser = ans.OpenBrowser

	cfg.Branch.Master = types.BranchTyp(ans.MasterBranch)
	cfg.Branch.Dev = types.BranchTyp(ans.DevBranch)
	cfg.Branch.Test = types.BranchTyp(ans.TestBranch)
	cfg.Branch.FeatureBranchPrefix = ans.FeatureBranchPrefix
	cfg.Branch.HotfixBranchPrefix = ans.HotfixBranchPrefix
	cfg.Branch.ConflictResolveBranchPrefix = ans.ConflictResolveBranchPrefix
	cfg.Branch.IssueBranchPrefix = ans.IssueBranchPrefix

	return err
}

type answer struct {
	APIUrl       string
	CallbackHost string

	OpenBrowser bool
	DebugMode   bool

	MasterBranch                string
	DevBranch                   string
	TestBranch                  string
	FeatureBranchPrefix         string
	HotfixBranchPrefix          string
	ConflictResolveBranchPrefix string
	IssueBranchPrefix           string
}
