package main

import (
	cli "github.com/urfave/cli/v2"
)

// getFeatureCommand
// gitlab-flow feature [command options] -c --conf
func getFeatureCommand() *cli.Command {
	return &cli.Command{
		Name:  "feature",
		Usage: "managing the works in developing.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force-create-mr",
				Value:       false,
				Usage:       "force to create Merge Request",
				DefaultText: "false",
				Required:    false,
			},
			&cli.StringFlag{
				Name:     "feature-branch-name",
				Aliases:  []string{"f"},
				Usage:    "input the `featureBranchName`",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "parse-issue-compatible",
				Usage:    "switch to parse issue name in compatible mode [deprecated]",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "auto-merge",
				Usage:    "auto merge request when feature is done",
				Required: false,
			},
		},
		Subcommands: getFeatureSubCommands(),
	}
}

// gitlab-flow hotfix [command options] -p --specProject
func getHotfixCommand() *cli.Command {
	return &cli.Command{
		Name:  "hotfix",
		Usage: "managing the works in hotfix.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force-create-mr",
				Value:       false,
				Usage:       "force to create Merge Request",
				DefaultText: "false",
				Required:    false,
			},
		},
		Subcommands: getHotfixSubCommands(),
	}
}

// getDashCommand
func getDashCommand() *cli.Command {
	return &cli.Command{
		Name:        "dash",
		Usage:       "overview of local development",
		Subcommands: getDashSubCommands(),
	}
}

// getSyncCommand
func getSyncCommand() *cli.Command {
	return &cli.Command{
		Name:        "sync",
		Usage:       "synchronize resource from remote gitlab server",
		Subcommands: getSyncSubCommands(),
	}
}

// getConfigCommand
// configure current project branch settings, which would override global settings.
// show print current project settings, if not set, use global setting as project setting
// flow2 config show/edit/init [--global] [--project]
func getConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "show current configuration",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "global",
				Aliases:     []string{"g"},
				Value:       false,
				DefaultText: "false",
				Usage:       "show global configuration",
				Required:    false,
			},
		},
		Subcommands: getConfigSubCommands(),
	}
}
