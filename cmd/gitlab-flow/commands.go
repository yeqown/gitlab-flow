package main

import (
	cli "github.com/urfave/cli/v2"
)

//
//// getInitCommand
//// Deprecated: use `flow config init` instead
//func getInitCommand() *cli.Command {
//	return &cli.Command{
//		Name: "init",
//		Usage: "initialize configuration gitlab-flow, generate default config file and sqlite DB " +
//			"related to the path",
//		Action: func(c *cli.Context) error {
//			confPath := c.String("conf")
//			cfg, err := conf.Load(confPath, nil)
//			if err != nil {
//				if !errors.Is(err, os.ErrNotExist) {
//					panic("load config file failed: " + err.Error())
//				}
//
//				cfg = conf.Default()
//			}
//
//			// Prompt user to input configuration
//			if err = surveyConfig(cfg); err != nil {
//				log.Errorf("failed to survey config: %v", err)
//				return err
//			}
//
//			// DONE(@yeqown): refresh user's access token
//			support := gitlabop.NewOAuth2Support(&gitlabop.OAuth2Config{
//				Host:         cfg.GitlabHost,
//				ServeAddr:    cfg.OAuth2.CallbackHost,
//				AccessToken:  "", // empty
//				RefreshToken: "", // empty
//				Scopes:       cfg.OAuth2.Scopes,
//			})
//			if err = support.Enter(""); err != nil {
//				log.
//					WithFields(log.Fields{"config": cfg}).
//					Error("gitlab-flow initialize.oauth failed:", err)
//				return err
//			}
//			cfg.OAuth2.AccessToken, cfg.OAuth2.RefreshToken = support.Load()
//
//			if err = conf.Save(confPath, cfg, nil); err != nil {
//				log.Errorf("gitlab-flow initialize.saveConfig failed: %v", err)
//				return err
//			}
//
//			log.Infof("gitlab-flow has initialized. conf path is %s", confPath)
//			return nil
//		},
//	}
//}

// getFeatureCommand
// gitlab-flow feature [command options] -c --conf
func getFeatureCommand() *cli.Command {
	return &cli.Command{
		Name:  "feature",
		Usage: "managing the works in developing.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force-create-mr",
				Value:       false,
				Usage:       "force to create Merge Request",
				DefaultText: "false",
				Required:    false,
			},
			&cli.StringFlag{
				Name:     "feature-branch-name",
				Aliases:  []string{"f"},
				Usage:    "input the `featureBranchName`",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "parse-issue-compatible",
				Usage:    "switch to parse issue name in compatible mode [deprecated]",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "auto-merge",
				Usage:    "auto merge request when feature is done",
				Required: false,
			},
		},
		Subcommands: getFeatureSubCommands(),
	}
}

// gitlab-flow hotfix [command options] -p --specProject
func getHotfixCommand() *cli.Command {
	return &cli.Command{
		Name:  "hotfix",
		Usage: "managing the works in hotfix.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force-create-mr",
				Value:       false,
				Usage:       "force to create Merge Request",
				DefaultText: "false",
				Required:    false,
			},
		},
		Subcommands: getHotfixSubCommands(),
	}
}

// getDashCommand
func getDashCommand() *cli.Command {
	return &cli.Command{
		Name:        "dash",
		Usage:       "overview of local development",
		Subcommands: getDashSubCommands(),
	}
}

// getSyncCommand
func getSyncCommand() *cli.Command {
	return &cli.Command{
		Name:        "sync",
		Usage:       "synchronize resource from remote gitlab server",
		Subcommands: getSyncSubCommands(),
	}
}

// getConfigCommand
// configure current project branch settings, which would override global settings.
// show print current project settings, if not set, use global setting as project setting
// flow2 config show/edit/init [--global] [--project]
func getConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "show current configuration",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "global",
				Aliases:     []string{"g"},
				Value:       false,
				DefaultText: "false",
				Usage:       "show global configuration",
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Value:       false,
				DefaultText: "false",
				Usage:       "show project configuration",
				Required:    false,
			},
		},
		Subcommands: getConfigSubCommands(),
	}
}
