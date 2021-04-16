package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/yeqown/gitlab-flow/internal/types"
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
		getFeatureResolveConflictCommand(),
		// getSyncMilestoneSubCommand(), moved into `sync` commands
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
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureBegin(opc, title, desc)
		},
	}
}

// getFeatureBeginIssueSubCommand
// gitlab-flow feature start-issue -b @branchName [title] [desc]
func getFeatureBeginIssueSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "open-issue",
		Usage:     "open an issue then create issue branch from feature branch, also merge request",
		ArgsUsage: "open-issue -f @featureBranchName @title @desc",
		Category:  "feature",
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			defer func() {
				log.
					WithFields(log.Fields{"args": c.Args().Slice()}).
					Debug("open Issue")
			}()

			issueTitle := c.Args().Get(0)
			issueDesc := c.Args().Get(1)
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureBeginIssue(opc, issueTitle, issueDesc)
		},
	}
}

// getFeatureFinishIssueSubCommand
// gitlab-flow feature close-issue -b @branchName -i @branchName
func getFeatureFinishIssueSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "close-issue",
		Usage:     "close an issue it's merge request.",
		ArgsUsage: "close-issue -i @issueBranchName -f @featureBranchName",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "issue_branch_name",
				Aliases:  []string{"i"},
				Value:    "",                            // default current branch
				Usage:    "input the `issueBranchName`", // be be overwritten
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			issueBranchName := c.String("issue_branch_name")
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureFinishIssue(opc, issueBranchName)
		},
	}
}

// gitlab-flow feature debug -b @branchName
func getFeatureDebugSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "open a merge request from feature branch into DevBranch",
		ArgsUsage: "-f, --feature_branch_name `featureBranchName`",
		Category:  "feature",
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureDebugging(opc)
		},
	}
}

// gitlab-flow feature test -b @branchName
func getFeatureTestSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "open a merge request from feature branch into TestBranch",
		ArgsUsage: "-f, --feature_branch_name `featureBranchName`",
		Category:  "feature",
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureTest(opc)
		},
	}
}

// gitlab-flow feature release -b  @branchName
func getFeatureReleaseSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "open a merge request from feature branch into MasterBranch",
		ArgsUsage: "-f, --feature_branch_name `featureBranchName`",
		Category:  "feature",
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureRelease(opc)
		},
	}
}

func getFeatureResolveConflictCommand() *cli.Command {
	return &cli.Command{
		Name:      "resolve-conflict",
		Usage:     "if there is a conflict of your merge request indicates conflicts(feature => target branch)",
		ArgsUsage: "-f, --feature_branch_name `featureBranchName`, -t, --target_branch `targetBranch`",
		Category:  "feature",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target_branch",
				Aliases:     []string{"t"},
				Usage:       "input the `targetBranch`",
				Value:       "master",
				DefaultText: "master",
				Required:    false,
			},
		},
		Action: func(c *cli.Context) error {
			targetBranchName := c.String("target_branch")
			opc := getOpFeatureContext(c)
			return getFlow(c).FeatureResolveConflict(opc, types.BranchTyp(targetBranchName))
		},
	}
}
