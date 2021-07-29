package main

import (
	"github.com/urfave/cli/v2"
)

func getSyncSubCommands() cli.Commands {
	return cli.Commands{
		getSyncProjectCommand(),
		getSyncMilestoneSubCommand(),
	}
}

// getSyncProjectCommand synchronize project from remote gitlab server into local database.
func getSyncProjectCommand() *cli.Command {
	return &cli.Command{
		Name:      "project",
		Usage:     "synchronize project information from remote gitlab server into local database",
		ArgsUsage: "",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:   "sync-project",
				Value:  true,
				Hidden: true,
			},
		},
		Action: func(c *cli.Context) error {
			return getFlow(c).SyncProject()
		},
	}
}

// getSyncMilestoneSubCommand
func getSyncMilestoneSubCommand() *cli.Command {
	return &cli.Command{
		Name:      "milestone",
		Usage:     "rebuild local features, branches, issues and merges from remote gitlab repository",
		ArgsUsage: "",
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
