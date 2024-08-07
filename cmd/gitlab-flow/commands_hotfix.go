package main

import (
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

// hotfix subcommands
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
		Name:      "open",
		Usage:     "open @title @description",
		ArgsUsage: "@title] [@description",
		Description: "open a hotfix branch and merge request to master. " +
			"\n@title title \n@desc description",
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

			opc := getOpHotfixContext(c)
			return getFlow(c).HotfixBegin(opc, title, desc)
		},
	}
}

// getHotfixStartSubCommand to finish hotfix
// gitlab-flow hotfix release @title @desc
func getHotfixFinishSubCommand() *cli.Command {
	return &cli.Command{
		Name:        "close",
		Usage:       "close [-b, --hotfix_branch_name `hotfixBranchName`]",
		ArgsUsage:   "[-b, --hotfix_branch_name `hotfixBranchName`]",
		Description: "close a hotfix development, then create a merge request into master",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "hotfix_branch_name",
				Aliases:  []string{"b"},
				Value:    "",                   // default current branch
				Usage:    "`hotfixBranchName`", // be overwritten
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			defer func() {
				log.
					WithFields(log.Fields{"args": c.Args().Slice()}).
					Debug("finish hotfix")
			}()
			hotfixBranchName := c.String("hotfix_branch_name")
			opc := getOpHotfixContext(c)
			return getFlow(c).HotfixFinish(opc, hotfixBranchName)
		},
	}
}
