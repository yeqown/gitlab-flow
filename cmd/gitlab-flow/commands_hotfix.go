package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

// hotfix 管理下的子命令
func getHotfixSubCommands() cli.Commands {
	return cli.Commands{
		getHotfixStartSubCommand(),
		getHotfixFinishSubCommand(),
	}
}

// getHotfixStartSubCommand to start hotfix
// gitlab-flow hotfix start @title @desc
func getHotfixStartSubCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "open a hotfix branch and merge request to master",
		ArgsUsage:   "@title @desc",
		Description: "@title title \n\t @desc description",
		Category:    "hotfix",
		Action: func(c *cli.Context) error {
			log.
				WithFields(log.Fields{"args": c.Args().Slice()}).
				Debug("start hotfix")

			title := c.Args().Get(0)
			desc := c.Args().Get(1)
			if title == "" {
				return errors.New("title could not be empty")
			}

			if desc == "" {
				return errors.New("desc could not be empty")
			}

			confPath := c.String("conf_path")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).HotfixBegin(title, desc)
		},
	}
}

// getHotfixStartSubCommand to finish hotfix
// gitlab-flow hotfix release @title @desc
func getHotfixFinishSubCommand() *cli.Command {
	return &cli.Command{
		Name:        "release",
		Usage:       "finish a hotfix",
		ArgsUsage:   "@branchName",
		Description: "@title title",
		Category:    "hotfix",
		Action: func(c *cli.Context) error {
			log.
				WithFields(log.Fields{"args": c.Args().Slice()}).
				Debug("finish hotfix")

			hotfixBranchName := c.Args().Get(0)
			if hotfixBranchName == "" {
				return errors.New("title could not be empty")
			}

			confPath := c.String("conf_path")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).HotfixFinish(hotfixBranchName)
		},
	}
}
