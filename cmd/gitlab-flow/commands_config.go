package main

import (
	"net/url"
	"os"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	gitlabop "github.com/yeqown/gitlab-flow/internal/gitlab-operator"
	"github.com/yeqown/gitlab-flow/internal/types"
)

func getConfigSubCommands() []*cli.Command {
	return []*cli.Command{
		getConfigInitCommand(),
		getConfigShowCommand(),
		getConfigEditCommand(),
	}
}

// getConfigInitCommand initialize configuration gitlab-flow, generate default config file and sqlite DB
// related to the path. This command interact with user to get configuration.
// Usage: gitlab-flow [flags] config init
func getConfigInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "initialize configuration gitlab-flow, generate default config file and sqlite DB",
		Action: func(c *cli.Context) error {
			confPath := c.String("conf")
			cfg, err := conf.Load(confPath, nil)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					panic("load config file failed: " + err.Error())
				}

				cfg = conf.Default()
			}

			// Prompt user to input configuration
			if err = surveyConfig(cfg, true, true, true); err != nil {
				log.Errorf("failed to survey config: %v", err)
				return err
			}

			// DONE(@yeqown): refresh user's access token
			support := gitlabop.NewOAuth2Support(&gitlabop.OAuth2Config{
				Host:         cfg.GitlabHost,
				ServeAddr:    cfg.OAuth2.CallbackHost,
				AccessToken:  "", // empty
				RefreshToken: "", // empty
				Scopes:       cfg.OAuth2.Scopes,
			})
			if err = support.Enter(""); err != nil {
				log.
					WithFields(log.Fields{"config": cfg}).
					Error("gitlab-flow initialize.oauth failed:", err)
				return err
			}
			cfg.OAuth2.AccessToken, cfg.OAuth2.RefreshToken = support.Load()

			if err = conf.Save(confPath, cfg, nil); err != nil {
				log.Errorf("gitlab-flow initialize.saveConfig failed: %v", err)
				return err
			}

			log.Infof("gitlab-flow has initialized. conf path is %s", confPath)
			return nil
		},
	}
}

// getConfigShowCommand show current configuration in the terminal.
// Usage: gitlab-flow [flags] config show
// Default print the project configuration, if it is not exist, print the default(global) configuration.
func getConfigShowCommand() *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: "show current configuration",
		Action: func(c *cli.Context) error {

			flags := parseGlobalFlags(c)
			ctx := resolveFlags(flags)

			cc := internal.NewConfig(ctx)
			cfg, ok, err := cc.SearchAndMerge()
			_, _ = cfg, ok

			// TODO: implement this

			if err != nil {
				log.Errorf("search and merge configuration failed: %v", err)
				return errors.Wrap(err, "search and merge configuration failed")
			}

			return nil
		},
	}
}

// getConfigEditCommand edit current configuration in the terminal, interact with user to get configuration.
// Usage: gitlab-flow [flags] config edit
// Currently, only support to edit the project configuration, and branch setting.
func getConfigEditCommand() *cli.Command {
	return &cli.Command{
		Name:  "edit",
		Usage: "edit current configuration",
		Action: func(c *cli.Context) error {
			flags := parseGlobalFlags(c)
			ctx := resolveFlags(flags)

			cc := internal.NewConfig(ctx)
			cfg, ok, err := cc.SearchAndMerge()
			_, _ = cfg, ok
			if err != nil {
				log.Errorf("search and merge configuration failed: %v", err)
				return errors.Wrap(err, "search and merge configuration failed")
			}

			// Prompt user to input configuration
			if err = surveyConfig(cfg, false, false, true); err != nil {
				log.Errorf("failed to survey config: %v", err)
				return err
			}

			if err = cc.Save(ctx.ConfPath(), cfg); err != nil {
				log.Errorf("save configuration failed: %v", err)
				return errors.Wrap(err, "save configuration failed")
			}

			return nil
		},
	}
}

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
func surveyConfig(cfg *types.Config, askGitlab, askFlags, askBranch bool) error {
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
