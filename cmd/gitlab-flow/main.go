package main

import (
	"os"

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
	app.Version = "v1.7.2"
	app.Description = `A tool for managing gitlab Feature/Milestone/Issue/MergeRequest as gitlab-flow.`
	app.Flags = _cliGlobalFlags

	setupLogger()
	setupCommands(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setupLogger() {
	log.SetTimeFormat(true, "")
	log.SetLogLevel(log.LevelInfo)
}

func setupCommands(app *cli.App) {
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetCallerReporter(true)
			log.SetLogLevel(log.LevelDebug)
		}

		return nil
	}

	app.Commands = []*cli.Command{
		getInitCommand(),
		getFeatureCommand(),
		getHotfixCommand(),
		getDashCommand(),
		getSyncCommand(),
	}
}
