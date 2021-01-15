package main

import (
	"os"

	"github.com/yeqown/gitlab-flow/internal/conf"

	"github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "gitlab-flow"
	app.Usage = "CLI tool"
	app.Authors = []*cli.Author{
		{
			Name:  "yeqown",
			Email: "yeqown@gmail.com",
		},
	}
	app.Version = "v1.5.1"
	app.Description = `A tool for managing gitlab Feature/Milestone/Issue/MergeRequest as gitlab-flow.`
	app.Flags = _globalFlags

	mountCommands(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mountCommands(app *cli.App) {
	app.Commands = []*cli.Command{
		getInitCommand(),
		getFeatureCommand(),
		getHotfixCommand(),
		getDashCommand(),
	}
}

var _globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "conf_path",
		Aliases:     []string{"c"},
		Value:       conf.DefaultConfPath(),
		DefaultText: "~/.gitlab-flow",
		Usage:       "choose which `path/to/file` to load",
		Required:    false,
	},
	&cli.BoolFlag{
		Name:        "debug",
		Value:       false,
		Usage:       "--debug",
		DefaultText: "false",
		Required:    false,
	},
}
