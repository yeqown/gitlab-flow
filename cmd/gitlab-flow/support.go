package main

import (
	"path"
	"path/filepath"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/urfave/cli/v2"
	"github.com/yeqown/log"
)

// _cliGlobalFlags should be used like this:
// flow --debug -c path/to/config SUB-COMMAND [...options]
var _cliGlobalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "conf",
		Aliases:     []string{"c"},
		Value:       conf.DefaultConfPath(),
		DefaultText: "~/.gitlab-flow",
		Usage:       "choose which `path/to/file` to load",
		Required:    false,
	},
	&cli.StringFlag{
		Name:        "cwd",
		Value:       conf.DefaultCWD(),
		DefaultText: conf.DefaultCWD(),
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
		DefaultText: path.Base(conf.DefaultCWD()),
		Usage:       "input `projectName` to locate which project should be operate.",
		Required:    false,
	},
	&cli.BoolFlag{
		Name:        "force-remote",
		Value:       false,
		DefaultText: "false",
		Usage: "query project from remote not from local. This should be used when project " +
			"name is duplicated, and could not found from local.",
		Required: false,
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
	ForceRemote bool
	CWD         string
}

func parseGlobalFlags(c *cli.Context) globalFlags {
	return globalFlags{
		ConfPath:    c.String("conf"),
		DebugMode:   c.Bool("debug"),
		ProjectName: c.String("project"),
		OpenBrowser: c.Bool("web"),
		ForceRemote: c.Bool("force-remote") || c.Bool("sync-project"),
		CWD:         c.String("cwd"),
	}
}

func getOpFeatureContext(c *cli.Context) *types.OpFeatureContext {
	return &types.OpFeatureContext{
		ForceCreateMergeRequest: c.Bool("force-create-mr"),
		FeatureBranchName:       c.String("feature-branch-name"),
		ParseIssueCompatible:    c.Bool("parse-issue-compatible"),
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
// DONE(@yeqown): CWD could be configured from CLI.
func setEnviron(flags globalFlags) *types.FlowContext {
	log.
		WithField("flags", flags).
		Debugf("setEnviron called")

	// get absolute path of current working directory.
	cwd, err := filepath.Abs(flags.CWD)
	if err != nil {
		log.
			WithField("cwd", flags.CWD).
			Fatalf("get absolute path of CWD failed: %v", err)
	}

	// prepare configuration
	cfg, err := conf.Load(flags.ConfPath, nil)
	if err != nil {
		log.
			WithField("path", flags.ConfPath).
			Fatalf("could not load config file: %v", err)
	}

	// pass flags parameters into configuration
	if flags.DebugMode {
		cfg.DebugMode = flags.DebugMode
	}
	if flags.OpenBrowser {
		cfg.OpenBrowser = flags.OpenBrowser
	}

	if err = cfg.Valid(); err != nil {
		log.
			WithField("cfg", cfg).
			Fatalf("config is invalid")
	}

	return types.NewContext(cwd, flags.ConfPath, flags.ProjectName, cfg, flags.ForceRemote)
}
