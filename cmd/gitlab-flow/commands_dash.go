package main

import (
	"fmt"

	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/urfave/cli/v2"
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
		Usage:     "查看一个迭代分支的详情（MR清单，Issue清单，Branch清单，关联的Milestone）",
		ArgsUsage: "@branchName",
		Category:  "dash",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "branchName",
				Aliases:  []string{"b"},
				Usage:    "迭代分支名",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			confPath := c.String("conf_path")
			debug := c.Bool("debug")

			featureBranchName := c.String("branchName")
			data, err := getDash(confPath, debug).FeatureDetail(featureBranchName)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", data)
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
		ArgsUsage: "@milestoneName",
		Category:  "dash",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "milestone_name",
				Aliases:  []string{"m"},
				Usage:    "-m, --milestone_name",
				Required: true,
			},
			&cli.StringFlag{
				Name:        "filter_branch",
				Aliases:     []string{"f"},
				Usage:       "-f, --filter_branch @branchName default: master",
				DefaultText: types.MasterBranch.String(),
			},
		},
		Action: func(c *cli.Context) error {
			milestoneName := c.String("milestone_name")
			filterBranchName := c.String("filter_branch")
			confPath := c.String("conf_path")
			debug := c.Bool("debug")
			data, err := getDash(confPath, debug).MilestoneOverview(milestoneName, filterBranchName)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", data)
			return nil
		},
	}
}

// gitlab-flow dash project
func getDashProjectDetailSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "project",
		Aliases:   []string{"p"},
		Usage:     "-p, --project",
		ArgsUsage: "-p @projectName, default current project",
		Category:  "dash",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "open_web",
				Aliases:  []string{"o"},
				Usage:    "-o, --open_web",
				Required: false,
				Value:    false,
			},
		},
		Action: func(c *cli.Context) error {
			confPath := c.String("conf_path")
			debug := c.Bool("debug")
			open := c.Bool("open_web")
			data, err := getDash(confPath, debug).ProjectDetail(open)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", data)
			return nil
		},
	}
}
