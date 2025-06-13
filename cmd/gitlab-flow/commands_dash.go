package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/yeqown/gitlab-flow/internal/types"
)

func getDashSubCommands() cli.Commands {
	return cli.Commands{
		getDashFeatureDetailSubCommand(),
		getDashProjectDetailSubCommand(),
		getDashMilestoneOverviewSubCommand(),
	}
}

// gitlab-flow dash feature -b featureBranchName
func getDashFeatureDetailSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "feature",
		Aliases:   []string{"f"},
		Usage:     "overview of the feature of current project.",
		ArgsUsage: "-b, --branch_name `BranchName`",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "branch_name",
				Aliases:     []string{"b"},
				Usage:       "input the target `BranchName`",
				DefaultText: "current branch",
				Required:    false,
			},
		},
		Action: func(c *cli.Context) error {
			featureBranchName := c.String("branch_name")
			data, err := getDash(c).FeatureDetail(featureBranchName)
			if err != nil {
				fmt.Printf("\nIf could not parse branch name by default, you can try:\n" +
					"1. specify a branch name by `-b YOUR-BRANCH-NAME`\n" +
					"2. switch to feature branch by `git checkout feature/YOUR-BRANCH-NAME`\n")
				return err
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
}

// gitlab-flow dash milestone -m milestoneName
func getDashMilestoneOverviewSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "milestone",
		Aliases:   []string{"m"},
		Usage:     "overview of one milestone, includes: merges, issues, branch",
		ArgsUsage: "-m, --milestone_name -b --branch_name",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "milestone_name",
				Aliases: []string{"m"},
				Usage: "input `milestoneName` which you want to get its information, " +
					"default current branch milestone",
				Required: false,
			},
			&cli.StringFlag{
				Name:        "branch_name",
				Aliases:     []string{"b"},
				Usage:       "filter `branchName`",
				DefaultText: types.MasterBranch.String(),
			},
		},
		Action: func(c *cli.Context) error {
			milestoneName := c.String("milestone_name")
			filterBranchName := c.String("branch_name")
			data, err := getDash(c).MilestoneOverview(milestoneName, filterBranchName)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
}

// gitlab-flow dash project
func getDashProjectDetailSubCommand() *cli.Command {
	return &cli.Command{
		Name:    "project",
		Aliases: []string{"p"},
		Usage:   "do something of current project.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "module",
				Aliases:     []string{"m"},
				Usage:       "project `module`, module is one of (home, tag, branch, commit)",
				DefaultText: "all",
				Value:       "all",
				Required:    false,
			},
		},
		Action: func(c *cli.Context) error {
			module := c.String("module")
			data, err := getDash(c).ProjectDetail(module)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
}
