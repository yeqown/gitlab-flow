package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/yeqown/log"
)

// _cliGlobalFlags should be used like this:
// flow --debug -c path/to/config SUB-COMMAND [...options]
var _cliGlobalFlags = []cli.Flag{
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
		Usage:       "verbose mode",
		DefaultText: "false",
		Required:    false,
	},
	&cli.StringFlag{
		Name:        "project",
		Aliases:     []string{"p"},
		Value:       "",
		DefaultText: "extract working directory",
		Usage:       "input `projectName` to locate which project should be operate.",
		Required:    false,
	},
	&cli.BoolFlag{
		Name:        "web",
		Value:       false,
		Usage:       "open web browser automatically or not",
		DefaultText: "false",
		Required:    false,
	},
}

type globalFlags struct {
	ConfPath    string
	DebugMode   bool
	ProjectName string
	OpenBrowser bool
}

func parseGlobalFlags(c *cli.Context) globalFlags {
	return globalFlags{
		ConfPath:    c.String("conf_path"),
		DebugMode:   c.Bool("debug"),
		ProjectName: c.String("project"),
		OpenBrowser: c.Bool("web"),
	}
}

func getFlow(c *cli.Context) internal.IFlow {
	flags := parseGlobalFlags(c)
	ctx := setEnviron(flags)
	return internal.NewFlow(ctx)
}

func getDash(c *cli.Context) internal.IDash {
	flags := parseGlobalFlags(c)
	ctx := setEnviron(flags)
	return internal.NewDash(ctx)
}

// setEnviron set global environment of debug mode.
// DONE(@yeqown): apply project name from CLI and CWD.
// TODO(@yeqown): CWD could be configured from CLI.
func setEnviron(flags globalFlags) *types.FlowContext {
	if !flags.DebugMode {
		log.SetLogLevel(log.LevelInfo)
	} else {
		// open caller report
		log.SetCallerReporter(true)
		log.SetLogLevel(log.LevelDebug)
	}
	log.
		WithField("flags", flags).
		Debugf("setEnviron called")

	// prepare configuration
	cfg, err := conf.Load(flags.ConfPath, nil)
	if err != nil {
		log.
			WithField("path", flags.ConfPath).
			Fatalf("could not load config file")
		panic("could not reach")
	}
	if err = cfg.
		Apply(flags.DebugMode, flags.OpenBrowser).
		Valid(); err != nil {
		log.
			WithField("cfg", cfg).
			Fatalf("config is invalid")
		panic("could not reach")
	}

	// generate a FlowContext
	cwd, _ := os.Getwd()
	return types.NewContext(cwd, flags.ConfPath, flags.ProjectName, cfg)
}
