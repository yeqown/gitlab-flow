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
		Name:        "open",
		Usage:       "open a milestone and branch name, feature name would be same to milestone",
		ArgsUsage:   "open @title @desc",
		Description: "@title title of milestone \n\t @desc description of milestone",
		Category:    "feature",
		Action: func(c *cli.Context) error {
			title := c.Args().Get(0)
			desc := c.Args().Get(1)
			if title == "" {
				return errors.New("'Title' could not be empty")
			}
			if desc == "" {
				return errors.New("'Description' could not be empty")
			}
			return getFlow(c).FeatureBegin(title, desc)
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
				Usage:    "input the target branch name",
				Value:    "",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			defer func() {
				log.
					WithFields(log.Fields{"args": c.Args().Slice()}).
					Debug("open Issue")
			}()

			issueTitle := c.Args().Get(0)
			issueDesc := c.Args().Get(1)
			featureBranchName := c.String("feature_branch_name")
			return getFlow(c).FeatureBeginIssue(featureBranchName, issueTitle, issueDesc)
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
				Usage:    "input the target branch name",
				Value:    "",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "issue_branch_name",
				Aliases:  []string{"ib"},
				Value:    "",                         // default current branch
				Usage:    "-ib, --issue_branch_name", // be be overwritten
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("feature_branch_name")
			issueBranchName := c.String("issue_branch_name")
			return getFlow(c).FeatureFinishIssue(featureBranchName, issueBranchName)
		},
	}
}

// gitlab-flow feature debug -b @branchName
func getFeatureDebugSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "open a merge request from feature branch into DevBranch",
		ArgsUsage: "-fb, --feature_branch_name `BranchName`",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"-fb"},
				Usage:    "input the target branch name",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("feature_branch_name")
			return getFlow(c).FeatureDebugging(featureBranchName)
		},
	}
}

// gitlab-flow feature test -b @branchName
func getFeatureTestSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "open a merge request from feature branch into TestBranch",
		ArgsUsage: "-fb, --feature_branch_name `BranchName`",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"-fb"},
				Usage:    "input the target branch name",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("feature_branch_name")
			return getFlow(c).FeatureTest(featureBranchName)
		},
	}
}

// gitlab-flow feature release -b  @branchName
func getFeatureReleaseSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "open a merge request from feature branch into MasterBranch",
		ArgsUsage: "-fb, --feature_branch_name `BranchName`",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "feature_branch_name",
				Aliases:  []string{"-fb"},
				Usage:    "input the target branch name",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("feature_branch_name")
			return getFlow(c).FeatureRelease(featureBranchName)
		},
	}
}

// getSyncMilestoneSubCommand
func getSyncMilestoneSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "sync",
		Usage:     "rebuild local data from remote gitlab repository",
		ArgsUsage: "",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "milestone_id",
				Aliases:  []string{"m"},
				Usage:    "choose milestone manually",
				Required: false,
			},
			&cli.BoolFlag{
				Name:    "interact",
				Aliases: []string{"i"},
				Usage:   "choose milestone in the list load from remote repository",
				Value:   false,
			},
		},
		Action: func(c *cli.Context) error {
			milestoneID := c.Int("milestoneID")
			interact := c.Bool("interact")
			return getFlow(c).SyncMilestone(milestoneID, interact)
		},
	}
}
