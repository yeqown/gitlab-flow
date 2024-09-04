package main

import (
	"fmt"
	"net/url"
	"os"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"github.com/yeqown/log"

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

func explainConfigFlags(c *cli.Context) (project, global bool, err error) {
	global = c.Bool("global")
	project = c.Bool("project")
	if global && project {
		err = errors.New("only one of global and project could be true")
		return project, global, err
	}

	if global || project {
		return project, global, nil
	}

	// both are false, then set the project to true as default
	project = true
	return project, global, nil
}

// getConfigInitCommand initialize configuration gitlab-flow, generate default config file and sqlite DB
// related to the path. This command interact with user to get configuration.
// Usage: gitlab-flow [flags] config init
func getConfigInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "initialize configuration gitlab-flow, generate default config file and sqlite DB",
		Action: func(c *cli.Context) error {
			_, global, err := explainConfigFlags(c)
			if err != nil {
				log.Errorf("explainConfigFlags failed: %v", err)
				return nil
			}

			flags := parseGlobalFlags(c)
			ch, err := getConfigHelper(flags)
			var cfg *types.Config
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					panic("load config file failed: " + err.Error())
				}
				cfg = conf.Default()
			}

			// Prompt user to input configuration
			if global {
				err = surveyConfig(cfg, true, true, true)
			} else {
				err = surveyConfig(cfg, false, false, true)
			}
			if err != nil {
				log.Errorf("failed to survey config: %v", err)
				return err
			}

			if global {
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
			}

			target, err := ch.Save(cfg, global)
			if err != nil {
				log.Errorf("gitlab-flow initialize.saveConfig failed: %v", err)
				return err
			}

			log.Infof("gitlab-flow has initialized into %s", target)
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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "raw",
				Aliases:     []string{"r"},
				Usage:       "show raw configuration",
				DefaultText: "true",
				Value:       true,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "output configuration to file, support json/toml format. not support yet",
				DefaultText: "none",
				Value:       "none",
			},
		},
		Action: func(c *cli.Context) error {
			_, global, err := explainConfigFlags(c)
			if err != nil {
				log.Errorf("explainConfigFlags failed: %v", err)
				return nil
			}

			flags := parseGlobalFlags(c)
			ch, err := getConfigHelper(flags)
			if err != nil {
				log.Errorf("preload configuration failed: %v", err)
				return nil
			}

			var (
				cfg *types.Config
			)

			// Display project configuration by default.
			cfg, err = ch.Project(true)
			if global {
				cfg, err = ch.Global()
			}
			if cfg == nil || err != nil {
				log.Errorf("could not get configuration with err: %v", err)
				return nil
			}

			// Branch settings
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Module", "Setting", "Value"})
			data := [][]string{
				{"Branch Settings", "Master", cfg.Branch.Master.String()},
				{"Branch Settings", "Dev", cfg.Branch.Dev.String()},
				{"Branch Settings", "Test", cfg.Branch.Test.String()},
				{"Branch Settings", "Feature Branch Prefix", cfg.Branch.FeatureBranchPrefix},
				{"Branch Settings", "Hotfix Branch Prefix", cfg.Branch.HotfixBranchPrefix},
				{"Branch Settings", "Conflict Branch Prefix", cfg.Branch.ConflictResolveBranchPrefix},
				{"Branch Settings", "Issue Branch Prefix", cfg.Branch.IssueBranchPrefix},

				{"Gitlab OAuth2", "Callback Host", cfg.OAuth2.CallbackHost},
				{"Gitlab OAuth2", "Access Token", cfg.OAuth2.AccessToken},
				{"Gitlab OAuth2", "Refresh Token", cfg.OAuth2.RefreshToken},

				{"Gitlab", "API Endpoint", cfg.GitlabAPIURL},
				{"Gitlab", "Host", cfg.GitlabHost},

				{"Flags", "Debug", fmt.Sprintf("%v", cfg.DebugMode)},
				{"Flags", "Auto Open Browser", fmt.Sprintf("%v", cfg.OpenBrowser)},
			}
			table.SetAutoMergeCells(true)
			// table.SetHeaderColor(
			// 	tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
			// 	tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
			// 	tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
			// )

			table.SetColumnColor(
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgWhiteColor},
			)
			table.AppendBulk(data)

			table.Render()

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
			_, global, err := explainConfigFlags(c)
			if err != nil {
				log.Errorf("explainConfigFlags failed: %v", err)
				return nil
			}

			flags := parseGlobalFlags(c)
			ch, err := getConfigHelper(flags)
			if err != nil {
				log.Errorf("preload configuration failed: %v", err)
				return errors.Wrap(err, "preload configuration failed")
			}

			cfg, err := ch.Project(true)
			if global {
				cfg, err = ch.Global()
			}
			if cfg == nil || err != nil {
				log.Error("could not get configuration")
				return errors.New("could not get configuration")
			}

			// Prompt user to input configuration
			if global {
				log.Info("You're editing global configuration !")
				// DO NOT Merge the project configuration into global configuration
				err = surveyConfig(cfg, true, true, true)
			} else {
				log.Info("You're editing project configuration!")
				err = surveyConfig(cfg, false, false, true)
			}
			if err != nil {
				log.Errorf("failed to survey config: %v", err)
				return err
			}

			select {
			case <-c.Context.Done():
				log.Warn("user canceled the operation")
				return nil
			default:
			}

			target, err := ch.Save(cfg, global)
			if err != nil {
				log.Errorf("gitlab-flow initialize.saveConfig failed: %v", err)
				return err
			}

			log.Infof("gitlab-flow configuration has been updated successfully: %s", target)
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
	if askGitlab {
		questions = append(questions, buildGitlabQuestions(cfg)...)
	}
	if askFlags {
		questions = append(questions, buildFlagsQuestions(cfg)...)
	}
	if askBranch {
		questions = append(questions, buildBranchQuestions(cfg)...)
	}

	if len(questions) == 0 {
		return nil
	}

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
