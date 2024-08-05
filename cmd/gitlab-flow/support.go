package main

import (
	"fmt"
	"path"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"
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
		ForceRemote: c.Bool("force-remote"),
		CWD:         c.String("cwd"),
	}
}

func getOpFeatureContext(c *cli.Context) *types.OpFeatureContext {
	return &types.OpFeatureContext{
		ForceCreateMergeRequest: c.Bool("force-create-mr"),
		FeatureBranchName:       c.String("feature-branch-name"),
		AutoMergeRequest:        c.Bool("auto-merge"),
		ParseIssueCompatible:    c.Bool("parse-issue-compatible"),
	}
}

func getOpHotfixContext(c *cli.Context) *types.OpHotfixContext {
	return &types.OpHotfixContext{
		ForceCreateMergeRequest: c.Bool("force-create-mr"),
	}
}

func getFlow(c *cli.Context) internal.IFlow {
	flags := parseGlobalFlags(c)
	ctx := resolveFlags(flags)
	return internal.NewFlow(ctx)
}

func getDash(c *cli.Context) internal.IDash {
	flags := parseGlobalFlags(c)
	ctx := resolveFlags(flags)
	return internal.NewDash(ctx)
}

func getConfig(c *cli.Context) internal.IConfig {
	flags := parseGlobalFlags(c)
	ctx := resolveFlags(flags)
	return internal.NewConfig(ctx)
}

// resolveFlags set global environment of debug mode.
// DONE(@yeqown): apply project name from CLI and CWD.
// DONE(@yeqown): CWD could be configured from CLI.
func resolveFlags(flags globalFlags) *types.FlowContext {
	log.
		WithField("flags", flags).
		Debugf("resolveFlags called")

	// get absolute path of current working directory.
	cwd, err := filepath.Abs(flags.CWD)
	if err != nil {
		log.
			WithField("cwd", flags.CWD).
			Fatalf("get absolute path of CWD failed: %v", err)
	}

	// prepare configuration
	var c *types.Config
	if c, err = conf.Load(flags.ConfPath, nil); err != nil {
		log.
			WithField("path", flags.ConfPath).
			Fatalf("could not load config file: %v", err)
	}

	// pass flags parameters into configuration
	if flags.DebugMode {
		c.DebugMode = flags.DebugMode
	}
	if flags.OpenBrowser {
		c.OpenBrowser = flags.OpenBrowser
	}

	if err = c.Valid(); err != nil {
		log.
			WithField("config", c).
			Fatalf("config is invalid")
	}

	types.SetBranchSetting(c.Branch.Master, c.Branch.Dev, c.Branch.Test)
	types.SetBranchPrefix(
		c.Branch.FeatureBranchPrefix,
		c.Branch.HotfixBranchPrefix,
		c.Branch.ConflictResolveBranchPrefix,
		c.Branch.IssueBranchPrefix,
	)

	return types.NewContext(cwd, flags.ConfPath, flags.ProjectName, c, flags.ForceRemote)
}

var _cliHelpTemplate = fmt.Sprintf(`
%s
%s
`, logoASCII, cli.AppHelpTemplate)

var logoASCII = `
 ________  ___  _________  ___       ________  ________          ________ ___       ________  ___       __      
|\   ____\|\  \|\___   ___\\  \     |\   __  \|\   __  \        |\  _____\\  \     |\   __  \|\  \     |\  \    
\ \  \___|\ \  \|___ \  \_\ \  \    \ \  \|\  \ \  \|\ /_       \ \  \__/\ \  \    \ \  \|\  \ \  \    \ \  \   
 \ \  \  __\ \  \   \ \  \ \ \  \    \ \   __  \ \   __  \       \ \   __\\ \  \    \ \  \\\  \ \  \  __\ \  \  
  \ \  \|\  \ \  \   \ \  \ \ \  \____\ \  \ \  \ \  \|\  \       \ \  \_| \ \  \____\ \  \\\  \ \  \|\__\_\  \ 
   \ \_______\ \__\   \ \__\ \ \_______\ \__\ \__\ \_______\       \ \__\   \ \_______\ \_______\ \____________\
    \|_______|\|__|    \|__|  \|_______|\|__|\|__|\|_______|        \|__|    \|_______|\|_______|\|____________|
`
