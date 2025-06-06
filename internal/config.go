package internal

import (
	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/conf"
	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	"github.com/yeqown/gitlab-flow/internal/types"
)

type IConfigHelper interface {
	// Context returns the context of configuration helper.
	Context() *ConfigHelperContext

	Config(typ types.ConfigType) types.ConfigHolder

	SaveTo(configType types.ConfigType) string
}

type ConfigHelperContext struct {
	CWD string

	ProjectConfPath string // project configuration file path
	GlobalConfPath  string // global configuration file path
}

func NewConfigHelper(helperContext *ConfigHelperContext) (IConfigHelper, error) {
	ch := &fileConfigImpl{
		helperContext: helperContext,

		globalConfig:  new(types.Config),
		projectConfig: new(types.ProjectConfig),

		gitOp: gitop.NewBasedCmd(helperContext.CWD),
	}

	err := ch.preload()
	if err != nil {
		log.Warnf("NewConfigHelper failed to preload: %v", err)
		return ch, errors.Wrap(err, "preload configuration failed")
	}

	return ch, nil
}

// fileConfigImpl is an implementation of IConfigHelper which is used to load and save configuration from file.
// It's searching for a configuration file in the following order:
// 1. Current git repository root directory.
// 2. User home directory.
// And merge them. Since the configuration in the current git repository has higher priority on branch setting.
// And the current git repository configuration can only change the branch setting yet.
type fileConfigImpl struct {
	helperContext *ConfigHelperContext

	globalConfig  *types.Config
	projectConfig *types.ProjectConfig

	gitOp gitop.IGitOperator
}

func (f *fileConfigImpl) preload() (err error) {
	err = conf.Load(f.helperContext.ProjectConfPath, f.projectConfig, false)
	if err != nil {
		log.Debugf("load project config file failed: %v", err)
	}

	err = conf.Load(f.helperContext.GlobalConfPath, f.globalConfig, true)
	if err != nil {
		log.Debugf("load global config file failed: %v", err)
		return errors.Wrap(err, "load global config file failed")
	}

	return nil
}

func (f *fileConfigImpl) Context() *ConfigHelperContext { return f.helperContext }

func (f *fileConfigImpl) Config(typ types.ConfigType) types.ConfigHolder {
	if typ == types.ConfigType_Global {
		return f.globalConfig
	}

	// typ == types.ConfigType_Project
	render := &types.ProjectConfig{
		ProjectName: f.projectConfig.ProjectName,
		Branch:      f.projectConfig.Branch,
		DebugMode:   f.projectConfig.DebugMode,
		OpenBrowser: f.projectConfig.OpenBrowser,
	}

	if f.projectConfig.Branch == nil {
		render.Branch = f.globalConfig.Branch
	}
	if f.projectConfig.DebugMode == nil {
		v := f.globalConfig.DebugMode
		render.DebugMode = &v
	}
	if f.projectConfig.OpenBrowser == nil {
		v := f.globalConfig.OpenBrowser
		render.OpenBrowser = &v
	}

	return render
}

func (f *fileConfigImpl) SaveTo(configType types.ConfigType) string {
	target := f.helperContext.ProjectConfPath
	if configType == types.ConfigType_Global {
		target = f.helperContext.GlobalConfPath
	}

	log.Debugf("SaveTo(%s) with config: %s", configType, target)

	return target
}
