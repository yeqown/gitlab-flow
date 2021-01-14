package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

// feature 管理下的 子命令
func getFeatureSubCommands() cli.Commands {
	return cli.Commands{
		getFeatureBeginSubCommand(),
		getFeatureBeginIssueSubCommand(),
		getFeatureFinishIssueSubCommand(),
		getFeatureDebugSubCommand(),
		getFeatureTestSubCommand(),
		getFeatureReleaseSubCommand(),
		getSyncMilestoneSubCommand(),
	}
}

// getFeatureBeginSubCommand to start feature command
// command: gitlab-flow feature start @title @desc
// @title will be used as branchName
// @desc will be used as milestone information
func getFeatureBeginSubCommand() *cli.Command {
	return &cli.Command{
		Name:        "begin",
		Usage:       "create a milestone and branch name, feature name would be same to milestone",
		ArgsUsage:   "begin @title @desc",
		Description: "@title title of milestone \n\t @desc description of milestone",
		Category:    "feature",
		Action: func(c *cli.Context) error {
			log.
				WithFields(log.Fields{"args": c.Args().Slice()}).
				Debug("create milestone and branch")

			title := c.Args().Get(0)
			desc := c.Args().Get(1)
			if title == "" {
				return errors.New("title could not be empty")
			}
			if desc == "" {
				return errors.New("description could not be empty")
			}
			debug := c.Bool("debug")

			confPath := c.String("conf_path")
			return getFlow(confPath, debug).FeatureBegin(title, desc)
		},
	}
}

// getFeatureBeginIssueSubCommand
// gitlab-flow feature start-issue -b @branchName [title] [desc]
func getFeatureBeginIssueSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "open-issue",
		Usage:     "open an issue then create issue branch from feature branch, also merge request",
		ArgsUsage: "open-issue -fb @featureBranchName @title @desc",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"fb"},
				Usage:    "-fb, --feature_branch_name",
				Value:    "",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			log.
				WithFields(log.Fields{"args": c.Args().Slice()}).
				Debug("open Issue")

			issueTitle := c.Args().Get(0)
			issueDesc := c.Args().Get(1)

			confPath := c.String("conf_path")
			featureBranchName := c.String("feature_branch_name")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).FeatureBeginIssue(featureBranchName, issueTitle, issueDesc)
		},
	}
}

// getFeatureFinishIssueSubCommand
// gitlab-flow feature close-issue -b @branchName -i @branchName
func getFeatureFinishIssueSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "close-issue",
		Usage:     "close an issue it's merge request.",
		ArgsUsage: "close-issue -ib @issueBranchName -fb @featureBranchName",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"fb"},
				Usage:    "-fb, --feature_branch_name",
				Value:    "",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "issue_branch_name",
				Aliases:  []string{"ib"},
				Value:    "",                         // default current branch
				Usage:    "-ib, --issue_branch_name", // be be overwritten
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("feature_branch_name")
			issueBranchName := c.String("issue_branch_name")
			confPath := c.String("conf_path")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).FeatureFinishIssue(featureBranchName, issueBranchName)
		},
	}
}

// gitlab-flow feature debug -b @branchName
func getFeatureDebugSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "open a merge request from feature branch into DevBranch",
		ArgsUsage: "close-issue @issueBranchName",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"-fb"},
				Usage:    "-fb, --feature_branch_name",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("feature_branch_name")
			confPath := c.String("conf_path")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).FeatureDebugging(featureBranchName)
		},
	}
}

// gitlab-flow feature test -b @branchName
func getFeatureTestSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "open a merge request from feature branch into TestBranch",
		ArgsUsage: "close-issue @issueBranchName",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"-fb"},
				Usage:    "-fb, --feature_branch_name",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			confPath := c.String("conf_path")
			featureBranchName := c.String("feature_branch_name")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).FeatureTest(featureBranchName)
		},
	}
}

// gitlab-flow feature release -b  @branchName
func getFeatureReleaseSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "open a merge request from feature branch into MasterBranch",
		ArgsUsage: "close-issue @issueBranchName",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"-fb"},
				Usage:    "-fb, --feature_branch_name",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			confPath := c.String("conf_path")
			featureBranchName := c.String("feature_branch_name")
			debug := c.Bool("debug")
			return getFlow(confPath, debug).FeatureRelease(featureBranchName)
		},
	}
}

// getSyncMilestoneSubCommand
func getSyncMilestoneSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "sync",
		Usage:     "rebuild local data from remote gitlab repository",
		ArgsUsage: "sync -m --milestoneID @id",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "milestone_id",
				Aliases:  []string{"m"},
				Usage:    "-m, --milestone_id",
				Required: false,
			},
			&cli.BoolFlag{
				Name:    "interact",
				Aliases: []string{"i"},
				Usage:   "-i, --interact",
			},
		},
		Action: func(c *cli.Context) error {
			confPath := c.String("conf_path")
			milestoneID := c.Int("milestoneID")
			debug := c.Bool("debug")
			interact := c.Bool("interact")

			f := getFlow(confPath, debug)
			return f.SyncMilestone(milestoneID, interact)
		},
	}
}
