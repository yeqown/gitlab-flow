package internal

import (
	gitop "github.com/yeqown/gitlab-flow/internal/git-operator"
	"github.com/yeqown/gitlab-flow/internal/types"
)

type IConfig interface {
	// SearchAndMerge searches configuration from file system. If found,
	// return the configuration and true, otherwise return false.
	// If current git repository is found, it will search configuration
	SearchAndMerge() (*types.Config, bool, error)

	// Save saves configuration to file system. If the file is not exist,
	// create it. If the file is existed, override it.
	// NOTE: the target is the file path.
	Save(target string, config *types.Config) error
}

func NewConfig(ctx *types.FlowContext) IConfig {
	return &fileConfigImpl{
		gitOp: gitop.NewBasedCmd(ctx.CWD()),
	}
}

// fileConfigImpl is an implementation of IConfig which is used to load and save configuration from file.
// It's searching for configuration file in the following order:
// 1. current git repository root directory.
// 2. user home directory.
// and merge them. since the configuration in current git repository has higher priority on branch setting.
// And the current git repository configuration can only change the branch setting yet.
type fileConfigImpl struct {
	gitOp gitop.IGitOperator
}

func (f *fileConfigImpl) SearchAndMerge() (*types.Config, bool, error) {
	// TODO: implement this
	return nil, false, nil
}

func (f *fileConfigImpl) Save(target string, config *types.Config) error {
	// TODO: implement this
	return nil
}
