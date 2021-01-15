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
		Category: "tools",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "access_token",
				Aliases:  []string{"s"},
				Required: true,
				Usage:    "access_token is secret for user to access gitlab API.",
			},
			&cli.StringFlag{
				Name:     "gitlab_host",
				Aliases:  []string{"hh"},
				Required: true,
				Usage:    "gitlab_host is the domain of YOUR gitlab server.",
			},
		},
		ArgsUsage: "-s ACCESS_TOKEN -h GITLAB_HOST [-c, --conf_path CONF_PATH]",
		Action: func(c *cli.Context) error {
			accessToken := c.String("access_token")
			host := c.String("gitlab_host")
			confPath := c.String("conf_path")

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
// gitlab-flow feature [command options] -c --conf_path
func getFeatureCommand() *cli.Command {
	return &cli.Command{
		Name:        "feature",
		Usage:       "managing the works in developing.",
		Category:    "feature",
		Subcommands: getFeatureSubCommands(),
	}
}

// gitlab-flow hotfix [command options] -p --specProject
func getHotfixCommand() *cli.Command {
	return &cli.Command{
		Name:        "hotfix",
		Usage:       "managing the works in hotfix.",
		Category:    "hotfix",
		Subcommands: getHotfixSubCommands(),
	}
}

// getDashCommand
func getDashCommand() *cli.Command {
	return &cli.Command{
		Name:        "dash",
		Usage:       "overview of local development",
		Category:    "dash",
		Subcommands: getDashSubCommands(),
	}
}
