package main

import (
	"errors"
	"os"

	"github.com/yeqown/gitlab-flow/internal/conf"
	gitlabop "github.com/yeqown/gitlab-flow/internal/gitlab-operator"

	"github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

func getInitCommand() *cli.Command {
	return &cli.Command{
		Name: "init",
		Usage: "initialize or migrate configuration gitlab-flow, generate default config file and sqlite DB " +
			"related to the path",
		Action: func(c *cli.Context) error {
			confPath := c.String("conf")
			cfg, err := conf.Load(confPath, nil)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					panic(err)
				}

				cfg = conf.Default()
			}

			if err = surveyConfig(cfg); err != nil {
				log.Errorf("failed to survey config: %v", err)
				return err
			}

			// DONE(@yeqown): refresh user's access token
			support := gitlabop.NewOAuth2Support(&gitlabop.OAuth2Config{
				Host:         cfg.GitlabHost,
				ServeAddr:    "", // use default
				AccessToken:  "", // empty
				RefreshToken: "", // empty
			})
			if err = support.Enter(""); err != nil {
				log.
					WithFields(log.Fields{"config": cfg}).
					Error("gitlab-flow initialize.oauth failed:", err)
				return err
			}
			cfg.OAuth.AccessToken, cfg.OAuth.RefreshToken = support.Load()

			if err = conf.Save(confPath, cfg, nil); err != nil {
				log.Errorf("gitlab-flow initialize.saveConfig failed: %v", err)
				return err
			}

			log.Infof("gitlab-flow has initialized. conf path is %s", confPath)
			return nil
		},
	}
}

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
				Usage:    "switch to parse issue name in compatible mode",
				Required: false,
			},
		},
		Subcommands: getFeatureSubCommands(),
	}
}

// gitlab-flow hotfix [command options] -p --specProject
func getHotfixCommand() *cli.Command {
	return &cli.Command{
		Name:        "hotfix",
		Usage:       "managing the works in hotfix.",
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
