package main

import (
	"github.com/yeqown/gitlab-flow/internal/conf"

	"github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

func getInitCommand() *cli.Command {
	return &cli.Command{
		Name: "init",
		Usage: "initialize gitlab-flow, generate default config file and sqlite DB " +
			"related to the path",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "access-token",
				Aliases:  []string{"s"},
				Required: true,
				Usage:    "access_token is `secret` for user to access gitlab API.",
			},
			&cli.StringFlag{
				Name:     "host",
				Aliases:  []string{"d"},
				Required: true,
				Usage:    "gitlab API host is the host of YOUR gitlab server. https://gitlab.example.com/api/v4/",
			},
		},
		ArgsUsage: "-s ACCESS_TOKEN -h GITLAB_HOST",
		Action: func(c *cli.Context) error {
			accessToken := c.String("access-token")
			host := c.String("gitlab-host")
			confPath := c.String("conf")

			cfg := conf.Default()
			cfg.AccessToken = accessToken
			cfg.GitlabAPIURL = host

			if err := conf.Save(confPath, cfg, nil); err != nil {
				log.Errorf("gitlab-flow initialize failed: %v", err)
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
				Usage:    "switch to parse issue name in comptaible mode",
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
