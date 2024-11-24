package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	cli "github.com/urfave/cli/v2"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/gitlab-flow/pkg"
)

// _cliGlobalFlags should be used like this:
// flow --debug -c path/to/config SUB-COMMAND [...options]
var _cliGlobalFlags = []cli.Flag{
	// &cli.StringFlag{
	// 	Name:        "conf",
	// 	Aliases:     []string{"c"},
	// 	Value:       conf.ConfigPath(),
	// 	DefaultText: "~/.gitlab-flow",
	// 	Usage:       "choose which `path/to/file` to load",
	// 	Required:    false,
	// },
	&cli.BoolFlag{
		Name:        "debug",
		Value:       false,
		Usage:       "verbose mode",
		DefaultText: "false",
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
	&cli.StringFlag{
		Name:        "cwd",
		Value:       "",
		DefaultText: "",
		Usage:       "choose which `path/to/file` to load",
		Required:    false,
	},
	&cli.StringFlag{
		Name:        "project",
		Aliases:     []string{"p"},
		Value:       "",
		DefaultText: "",
		Usage:       "input `projectName` to locate which project should be operate.",
		Required:    false,
	},
}

type globalFlags struct {
	DebugMode   bool // verbose mode
	OpenBrowser bool // open web browser automatically or not
	ForceRemote bool // DO NOT query from local, or create remote resource even if local has the same name.

	// ProjectName is the name of project which should be operated.
	// if not set, will use the project name from current git repository.
	ProjectName string
	// CWD is the current working directory,
	// if not set, will use the current git repository root path.
	CWD string
}

func parseGlobalFlags(c *cli.Context) globalFlags {
	return globalFlags{
		DebugMode:   c.Bool("debug"),
		OpenBrowser: c.Bool("web"),
		ForceRemote: c.Bool("force-remote"),
		ProjectName: c.String("project"),
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
	ctx, ch := resolveFlags(flags)
	return internal.NewFlow(ctx, ch)
}

func getDash(c *cli.Context) internal.IDash {
	flags := parseGlobalFlags(c)
	ctx, ch := resolveFlags(flags)
	return internal.NewDash(ctx, ch)
}

func getConfigHelper(flags globalFlags) (internal.IConfigHelper, error) {
	ctx := &internal.ConfigHelperContext{
		CWD: flags.CWD,
	}
	ch := internal.NewConfigHelper(ctx)
	err := ch.Preload()
	return ch, err
}

// resolveFlags set global environment of debug mode.
// DONE(@yeqown): apply project name from CLI and CWD.
// DONE(@yeqown): CWD could be configured from CLI.
func resolveFlags(flags globalFlags) (*types.FlowContext, internal.IConfigHelper) {
	log.
		WithField("flags", flags).
		Debugf("resolveFlags called")

	var (
		cwd         string = flags.CWD
		projectName string = flags.ProjectName
	)

	// if current working directory is not set, use default cwd.
	if cwd == "" {
		cwd = defaultCWD()
	}
	// get absolute path of current working directory.
	cwd, err := filepath.Abs(flags.CWD)
	if err != nil {
		log.Fatalf("get absolute path of CWD(%s) failed: %v", flags.CWD, err)
	}

	// if project name is not set, use the base name of current working directory.
	if projectName == "" {
		projectName = path.Base(cwd)
	}

	helper, err := getConfigHelper(flags)
	if err != nil {
		log.Fatalf("could not preload configuration: %v", err)
		return nil, nil
	}
	c, err := helper.Project(true)
	if err != nil {
		log.Fatalf("could not get merged configuration: %v", err)
		return nil, nil
	}

	// pass flags parameters into configuration
	if flags.DebugMode {
		c.DebugMode = flags.DebugMode
	}
	if flags.OpenBrowser {
		c.OpenBrowser = flags.OpenBrowser
	}

	if err = helper.ValidateConfig(c, true); err != nil {
		log.
			WithFields(log.Fields{
				"config":          c,
				"config.branches": c.Branch,
				"oauth":           c.OAuth2,
			}).
			Fatalf("config is invalid: %v", err)
	}

	types.SetBranchSetting(c.Branch.Master, c.Branch.Dev, c.Branch.Test)
	types.SetBranchPrefix(
		c.Branch.FeatureBranchPrefix,
		c.Branch.HotfixBranchPrefix,
		c.Branch.ConflictResolveBranchPrefix,
		c.Branch.IssueBranchPrefix,
	)

	return types.NewContext(cwd, flags.ProjectName, c, flags.ForceRemote), helper
}

var (
	_defaultCWD     string
	_defaultCwdOnce sync.Once
)

var tips = `HINT: current working directory maybe not a git repository, please make sure you 
HINT: are in a git repository, retry after you are in a git repository.
HINT:
HINT: If you're in a git repository, please make sure you have installed git command.
HINT:
HINT: You are able to submit an issue to: 
HINT: https://github.com/yeqown/gitlab-flow/issues
`

// defaultCWD returns the working directory of current project, default cwd is from
// git rev-parse --show-toplevel command, but if the command could not execute successfully,
// `pwd` command will be used instead.
//
// NOTICE that: if current working directory is not a git repository, the function shutdown and
// print tips to user.
func defaultCWD() string {
	_defaultCwdOnce.Do(func() {
		w := bytes.NewBuffer(nil)
		if err := pkg.RunOutput("git rev-parse --show-toplevel", w); err != nil {
			fmt.Printf(tips)
			log.Fatalf("executing 'git rev-parse --show-toplevel' failed: %v", err)
		}

		if s := w.String(); s != "" {
			_defaultCWD = s
		}

		if _defaultCWD == "" {
			_defaultCWD, _ = os.Getwd()
		}

		_defaultCWD = strings.Trim(_defaultCWD, "\n")
		_defaultCWD = strings.Trim(_defaultCWD, "\t")
	})

	return _defaultCWD
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
