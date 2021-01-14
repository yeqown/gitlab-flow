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
			"related to the path (default path is `~/.gitlab-flow/`)",
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
				Aliases:  []string{"h"},
				Required: true,
				Usage:    "gitlab_host is the domain of YOUR gitlab server.",
			},
			&cli.StringFlag{
				Name:     "conf_path",
				Aliases:  []string{"c"},
				Value:    conf.DefaultConfPath(),
				Required: true,
				Usage:    "conf_path is the directory which contains your config and local database.",
			},
		},
		ArgsUsage: "gitlab-flow init -s ACCESS_TOKEN -h GITLAB_HOST [-c CONF_PATH]",
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
		Name:      "feature",
		Usage:     "managing the works in developing.",
		ArgsUsage: "gitlab-flow hotfix [-c, --conf_path] [-v, --debug]",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "conf_path",
				Aliases:  []string{"c"},
				Value:    conf.DefaultConfPath(),
				Usage:    "-c, --conf_path",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  []string{"v"},
				Value:    false,
				Usage:    "-v, --debug ",
				Required: false,
			},
		},
		Subcommands: getFeatureSubCommands(),
	}
}

// gitlab-flow hotfix [command options] -p --specProject
func getHotfixCommand() *cli.Command {
	return &cli.Command{
		Name:      "hotfix",
		Usage:     "managing the works in hotfix.",
		ArgsUsage: "gitlab-flow hotfix [-c, --conf_path] [-v, --debug]",
		Category:  "hotfix",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "conf_path",
				Aliases:  []string{"c"},
				Value:    conf.DefaultConfPath(),
				Usage:    "-c, --conf_path",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  []string{"v"},
				Value:    false,
				Usage:    "-v, --debug ",
				Required: false,
			},
		},
		Subcommands: getHotfixSubCommands(),
	}
}

// getDashCommand
func getDashCommand() *cli.Command {
	return &cli.Command{
		Name:        "dash",
		Usage:       "gitlab-flow dash",
		Description: "overview of local development",
		Category:    "dash",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "conf_path",
				Aliases:  []string{"c"},
				Value:    conf.DefaultConfPath(),
				Usage:    "-c, --conf_path",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  []string{"v"},
				Value:    false,
				Usage:    "-v, --debug ",
				Required: false,
			},
		},
		Subcommands: getDashSubCommands(),
	}
}
