package internal

import (
	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/conf"
	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	"github.com/yeqown/gitlab-flow/internal/types"
)

type IConfigHelper interface {
	// Preload loads configuration from file system. If the configuration
	// is not found, return error.
	// NOTE: this method also refill the context of configuration helper.
	Preload() error

	// Context returns the context of configuration helper.
	Context() *ConfigHelperContext

	// Global returns the global configuration, if the merge is true,
	// it will merge the project configuration into global configuration.
	Global() (*types.Config, error)

	// Project returns the configuration of current project, if the merge is true,
	// it will merge the global configuration into project configuration.
	Project(merge bool) (*types.Config, error)

	// Save saves configuration to file system. If the file is not exist,
	// create it. If the file is existed, override it.
	// NOTE: the target is the file path.
	Save(config *types.Config, global bool) (string, error)
}

type ConfigHelperContext struct {
	CWD string

	ProjectConfPath string // project configuration file path
	GlobalConfPath  string // global configuration file path
}

func NewConfigHelper(ctx *ConfigHelperContext) IConfigHelper {
	ch := &fileConfigImpl{
		ctx: ctx,

		globalConfig:  nil,
		projectConfig: nil,

		gitOp: gitop.NewBasedCmd(ctx.CWD),
	}

	return ch
}

// fileConfigImpl is an implementation of IConfigHelper which is used to load and save configuration from file.
// It's searching for configuration file in the following order:
// 1. current git repository root directory.
// 2. user home directory.
// and merge them. since the configuration in current git repository has higher priority on branch setting.
// And the current git repository configuration can only change the branch setting yet.
type fileConfigImpl struct {
	ctx *ConfigHelperContext

	globalConfig  *types.Config
	projectConfig *types.Config

	gitOp gitop.IGitOperator
}

func (f *fileConfigImpl) Preload() (err error) {
	f.ctx.ProjectConfPath = conf.ConfigPath(f.ctx.CWD)
	f.ctx.GlobalConfPath = conf.ConfigPath("")

	f.projectConfig, err = conf.Load(f.ctx.ProjectConfPath, nil)
	if err != nil {
		log.Debugf("load project configuration failed: %v", err)
	}

	f.globalConfig, err = conf.Load(f.ctx.GlobalConfPath, nil)
	if err != nil {
		return errors.Wrap(err, "load global configuration failed")
	}

	return nil
}

func (f *fileConfigImpl) Context() *ConfigHelperContext {
	return f.ctx
}

func (f *fileConfigImpl) Global() (*types.Config, error) {
	return f.globalConfig, nil
}

func (f *fileConfigImpl) Project(merge bool) (*types.Config, error) {
	// if !merge {
	// 	if f.projectConfig == nil {
	// 		return nil, errors.New("project not found")
	// 	}
	//
	// 	return f.projectConfig, nil
	// }

	// merge global configuration into project configuration.
	if f.projectConfig == nil {
		log.Debugf("project configuration not found, use global configuration")
		return f.globalConfig, nil
	}

	// merge global configuration(except branch settings) into project configuration.
	return &types.Config{
		OAuth2:       f.globalConfig.OAuth2,
		Branch:       f.projectConfig.Branch,
		GitlabAPIURL: f.globalConfig.GitlabAPIURL,
		GitlabHost:   f.globalConfig.GitlabHost,
		DebugMode:    f.globalConfig.DebugMode,
		OpenBrowser:  f.globalConfig.OpenBrowser,
	}, nil
}

func (f *fileConfigImpl) Save(config *types.Config, global bool) (target string, err error) {
	target = f.ctx.ProjectConfPath
	if global {
		target = f.ctx.GlobalConfPath
	}

	err = conf.Save(target, config, global)
	if err != nil {
		return target, errors.Wrap(err, "save configuration failed")
	}

	return target, nil
}
