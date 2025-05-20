package main

import (
	"fmt"
	"net/url"
	"os"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
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

func explainConfigFlags(c *cli.Context) types.ConfigType {
	global := c.Bool("global")
	if global {
		return types.ConfigType_Global
	}

	return types.ConfigType_Project
}

// getConfigInitCommand initialize configuration gitlab-flow, generate a default config file and sqlite DB
// related to the path. This command interacts with user to get configuration.
// Usage: gitlab-flow [flags] config init
func getConfigInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "initialize configuration gitlab-flow, generate default config file and sqlite DB",
		Action: func(c *cli.Context) error {
			configType := explainConfigFlags(c)
			flags := parseGlobalFlags(c)
			ch, err := getConfigHelper(flags)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					panic("load config file failed: " + err.Error())
				}
			}

			var configHolder types.ConfigHolder

			switch configType {
			case types.ConfigType_Project:
				configHolder = ch.Config(types.ConfigType_Project)
				err = surveyProjectConfig(configHolder.AsProject())
			default:
				configHolder = conf.Default()
				err = surveyConfig(configHolder.AsGlobal())
			}
			if err != nil {
				log.Warnf("failed to survey config: %v", err)
				return err
			}

			if err = configHolder.ValidateConfig(); err != nil {
				log.Errorf("config is invalid: %v", err)
				return err
			}

			if configType == types.ConfigType_Global {
				cfg := configHolder.AsGlobal()
				support := gitlabop.NewOAuth2Support(gitlabop.NewOAuth2ConfigFrom(cfg))
				if err = support.Enter(""); err != nil {
					log.
						WithFields(log.Fields{"config": cfg}).
						Error("gitlab-flow initialize.oauth failed:", err)
					return err
				}
				cfg.OAuth2.AccessToken, cfg.OAuth2.RefreshToken = support.Load()
			}

			target := ch.SaveTo(configType)
			if !surveySaveChoice(target) {
				log.Info("Aborted to save configuration")
				return nil
			}

			if err = conf.Save(target, configHolder); err != nil {
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
// Default print the project configuration, if it does not exist
// print the default(global) configuration.
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
			configType := explainConfigFlags(c)
			flags := parseGlobalFlags(c)
			ch, err := getConfigHelper(flags)
			if err != nil {
				log.Errorf("preload configuration failed: %v", err)
				return nil
			}

			// Display project configuration by default.
			configHolder := ch.Config(configType)
			if configHolder == nil {
				log.Error("could not get configuration")
				return errors.New("could not get configuration")
			}

			// Branch settings
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Module", "Setting", "Value"})
			table.SetAutoMergeCells(true)
			table.SetColumnColor(
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgWhiteColor},
			)

			switch configHolder.Type() {
			case types.ConfigType_Project:
				cfg := configHolder.AsProject()
				data := fillConfigRenderData(
					cfg.Branch,
					nil,
					"",
					"",
					cfg.DebugMode,
					cfg.OpenBrowser,
					cfg.ProjectName,
				)
				table.AppendBulk(data)
			case types.ConfigType_Global:
				cfg := configHolder.AsGlobal()
				data := fillConfigRenderData(
					cfg.Branch,
					cfg.OAuth2,
					cfg.GitlabAPIURL,
					cfg.GitlabHost,
					&cfg.DebugMode,
					&cfg.OpenBrowser,
					"",
				)
				table.AppendBulk(data)
			}

			table.Render()

			return nil
		},
	}
}

func fillConfigRenderData(
	branch *types.BranchSetting,
	oauth2 *types.OAuth,
	gitlabAPIURL, gitlabHost string,
	debug, openBrowser *bool,
	projectName string,
) (data [][]string) {
	data = make([][]string, 0, 10)

	if projectName != "" {
		data = append(data, []string{"Project", "Name", projectName})
	}

	if branch != nil {
		data = append(data, []string{"Branch Settings", "Master", branch.Master.String()})
		data = append(data, []string{"Branch Settings", "Dev", branch.Dev.String()})
		data = append(data, []string{"Branch Settings", "Test", branch.Test.String()})
		data = append(data, []string{"Branch Settings", "Feature Branch Prefix", branch.FeatureBranchPrefix})
		data = append(data, []string{"Branch Settings", "Hotfix Branch Prefix", branch.HotfixBranchPrefix})
		data = append(data, []string{"Branch Settings", "Conflict Branch Prefix", branch.ConflictResolveBranchPrefix})
	}

	if oauth2 != nil {
		data = append(data, []string{"Gitlab OAuth2", "Callback Host", oauth2.CallbackHost})
		data = append(data, []string{"Gitlab OAuth2", "Access Token", oauth2.AccessToken})
		data = append(data, []string{"Gitlab OAuth2", "Refresh Token", oauth2.RefreshToken})
	}

	if gitlabAPIURL != "" {
		data = append(data, []string{"Gitlab", "API Endpoint", gitlabAPIURL})
	}
	if gitlabHost != "" {
		data = append(data, []string{"Gitlab", "Host", gitlabHost})
	}

	if debug != nil {
		data = append(data, []string{"Flags", "Debug", fmt.Sprintf("%v", *debug)})
	}
	if openBrowser != nil {
		data = append(data, []string{"Flags", "Auto Open Browser", fmt.Sprintf("%v", *openBrowser)})
	}

	return data
}

// getConfigEditCommand edit current configuration in the terminal, interact with user to get configuration.
// Usage: gitlab-flow [flags] config edit
// Currently, only support to edit the project configuration, and branch setting.
func getConfigEditCommand() *cli.Command {
	return &cli.Command{
		Name:  "edit",
		Usage: "edit current configuration",
		Action: func(c *cli.Context) error {
			configType := explainConfigFlags(c)
			flags := parseGlobalFlags(c)
			helper, err := getConfigHelper(flags)
			if err != nil {
				log.Errorf("preload configuration failed: %v", err)
				return errors.Wrap(err, "preload configuration failed")
			}

			configHolder := helper.Config(configType)

			switch configType {
			case types.ConfigType_Project:
				err = surveyProjectConfig(configHolder.AsProject())
			default:
				err = surveyConfig(configHolder.AsGlobal())
			}
			if err != nil {
				log.Warnf("failed to survey config: %v", err)
				return err
			}
			if err = configHolder.ValidateConfig(); err != nil {
				log.Errorf("config is invalid: %v", err)
				return err
			}

			select {
			case <-c.Context.Done():
				log.Warn("user canceled the operation")
				return nil
			default:
			}

			target := helper.SaveTo(configType)
			if !surveySaveChoice(target) {
				log.Info("Aborted to save configuration")
				return nil
			}

			if err = conf.Save(target, configHolder); err != nil {
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
		{
			Name: "appID",
			Prompt: &survey.Input{
				Message: "Input your gitlab AppID",
				Default: "",
				Help:    "DES + base64 encoded, default DES secret: `aflowcli`",
			},
			Validate: survey.Required,
		},
		{
			Name: "appSecret",
			Prompt: &survey.Input{
				Message: "Input your gitlab AppSecret",
				Default: "",
				Help:    "DES + base64 encoded, default DES secret: `aflowcli`",
			},
			Validate: survey.Required,
		},
	}
}

func buildFlagsQuestions(debugMode, openBrowser bool, withOAuthMode bool) []*survey.Question {
	qs := []*survey.Question{
		{
			Name: "debugMode",
			Prompt: &survey.Confirm{
				Message: "Would you like to use gitlab in debug mode?",
				Default: debugMode,
			},
			Validate:  nil,
			Transform: nil,
		},
		{
			Name: "openBrowser",
			Prompt: &survey.Confirm{
				Message: "Would you let gitlab-flow open browser automatically when needed",
				Default: openBrowser,
			},
			Validate:  nil,
			Transform: nil,
		},
	}

	if withOAuthMode {
		qs = append(qs, &survey.Question{
			Name: "oauthMode",
			Prompt: &survey.Select{
				Message: "Select your OAuth2 mode. If you are not in desktop environment, please select manual",
				Options: []string{
					"auto",
					"manual",
				},
				Default: "auto",
			},
			Validate:  nil,
			Transform: nil,
		})
	}

	return qs
}

func buildProjectNameQuestions(name string) []*survey.Question {
	return []*survey.Question{
		{
			Name: "projectName",
			Prompt: &survey.Input{
				Message: "Input your project name",
				Default: name,
			},
		},
	}
}

func buildBranchQuestions(cfg *types.BranchSetting) []*survey.Question {
	return []*survey.Question{
		{
			Name: "masterBranch",
			Prompt: &survey.Input{
				Message: "Input your master branch name",
				Default: string(cfg.Master),
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "devBranch",
			Prompt: &survey.Input{
				Message: "Input your dev branch name",
				Default: string(cfg.Dev),
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "testBranch",
			Prompt: &survey.Input{
				Message: "Input your test branch name",
				Default: string(cfg.Test),
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "featureBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your feature branch prefix",
				Default: cfg.FeatureBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "hotfixBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your hotfix branch prefix",
				Default: cfg.HotfixBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "conflictResolveBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your conflict resolve branch prefix",
				Default: cfg.ConflictResolveBranchPrefix,
			},
			Validate:  survey.Required,
			Transform: nil,
		},
		{
			Name: "issueBranchPrefix",
			Prompt: &survey.Input{
				Message: "Input your issue branch prefix",
				Default: cfg.IssueBranchPrefix,
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
	questions = append(questions, buildFlagsQuestions(cfg.DebugMode, cfg.OpenBrowser, true)...)
	questions = append(questions, buildBranchQuestions(cfg.Branch)...)

	ans := new(configSurveyAns)
	if err := survey.Ask(questions, ans); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			log.Warnf("user canceled the operation")
		}
		return errors.Wrap(err, "survey.Ask failed")
	}
	log.
		WithField("configSurveyAns", ans).
		Debug("surveyConfig done")

	u, err := url.Parse(ans.APIUrl)
	if err != nil {
		return errors.Wrap(err, "gitlab API URL is invalid")
	}
	cfg.GitlabAPIURL = ans.APIUrl
	// only save the scheme and host
	cfg.GitlabHost = u.Scheme + "://" + u.Host
	cfg.OAuth2.CallbackHost = ans.CallbackHost
	cfg.OAuth2.Mode = func(a string) types.OAuth2Mode {
		switch a {
		case "auto":
			return types.OAuth2Mode_Auto
		case "manual":
			return types.OAuth2Mode_Manual
		}
		return types.OAuth2Mode_Auto
	}(ans.OAuthMode)
	cfg.OAuth2.AppID = ans.AppID
	cfg.OAuth2.AppSecret = ans.AppSecret

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

type configSurveyAns struct {
	APIUrl       string
	CallbackHost string

	OpenBrowser bool
	DebugMode   bool
	OAuthMode   string
	AppID       string
	AppSecret   string

	MasterBranch                string
	DevBranch                   string
	TestBranch                  string
	FeatureBranchPrefix         string
	HotfixBranchPrefix          string
	ConflictResolveBranchPrefix string
	IssueBranchPrefix           string
}

func surveyProjectConfig(cfg *types.ProjectConfig) error {
	questions := make([]*survey.Question, 0, 4)
	questions = append(questions, buildProjectNameQuestions(cfg.ProjectName)...)
	questions = append(questions, buildBranchQuestions(cfg.Branch)...)
	questions = append(questions, buildFlagsQuestions(*cfg.DebugMode, *cfg.OpenBrowser, false)...)

	ans := new(projectSurveyAns)
	if err := survey.Ask(questions, ans); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			log.Warnf("user canceled the operation")
		}

		return errors.Wrap(err, "survey.Ask failed")
	}

	cfg.ProjectName = ans.ProjectName

	cfg.DebugMode = &ans.DebugMode
	cfg.OpenBrowser = &ans.OpenBrowser

	cfg.Branch.Master = types.BranchTyp(ans.MasterBranch)
	cfg.Branch.Dev = types.BranchTyp(ans.DevBranch)
	cfg.Branch.Test = types.BranchTyp(ans.TestBranch)
	cfg.Branch.FeatureBranchPrefix = ans.FeatureBranchPrefix
	cfg.Branch.HotfixBranchPrefix = ans.HotfixBranchPrefix
	cfg.Branch.ConflictResolveBranchPrefix = ans.ConflictResolveBranchPrefix
	cfg.Branch.IssueBranchPrefix = ans.IssueBranchPrefix

	return nil
}

type projectSurveyAns struct {
	ProjectName string

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

func surveySaveChoice(target string) bool {
	ans := new(bool)
	if err := survey.AskOne(&survey.Confirm{
		Message: "The configuration would saved to " + target + ", continue?",
		Default: true,
	}, ans); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			log.Warnf("user canceled the operation")
		}
	}

	return *ans
}
