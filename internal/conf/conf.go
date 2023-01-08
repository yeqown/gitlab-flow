// Package conf implements the configuration of the application.
// The configuration is loaded from the configuration file. The configuration
// file is a TOML file.
// TODO(@yeqown): store configuration in the database instead of the file,
// and then we can use CLI application to manage and checkout these configurations.
package conf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/gitlab-flow/pkg"
)

// Parser is an interface to parse config in different ways.
// For example: JSON, TOML and YAML;
type Parser interface {
	// Unmarshal ...
	Unmarshal(r io.Reader, rcv *types.Config) error

	// Marshal ...
	Marshal(cfg *types.Config) ([]byte, error)
}

// Load to load config from confPath with specified parser.
func Load(confPath string, parser Parser) (cfg *types.Config, err error) {
	if parser == nil {
		parser = NewTOMLParser()
	}

	cfg = &types.Config{
		OAuth:        new(types.OAuth),
		Branch:       new(types.BranchSetting),
		GitlabAPIURL: "",
		GitlabHost:   "",
		DebugMode:    false,
		OpenBrowser:  false,
	}
	p := precheckConfigDirectory(confPath)
	var r io.Reader
	r, err = os.OpenFile(p, os.O_RDONLY, 0777)
	if err != nil {
		return nil, errors.Wrap(err, "conf.Load")
	}
	err = parser.Unmarshal(r, cfg)

	return cfg, err
}

// Save to save config with specified parser.
func Save(confPath string, cfg *types.Config, parser Parser) error {
	if parser == nil {
		parser = NewTOMLParser()
	}

	data, err := parser.Marshal(cfg)
	if err != nil {
		return err
	}

	p := precheckConfigDirectory(confPath)
	w, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, string(data))
	return err
}

var (
	defaultConf = &types.Config{
		Branch: &types.BranchSetting{
			Master: types.MasterBranch,
			Dev:    types.DevBranch,
			Test:   types.TestBranch,
		},
		OAuth: &types.OAuth{
			AccessToken:  "",
			RefreshToken: "",
		},
		GitlabAPIURL: "https://YOUR_HOSTNAME/api/v4",
		GitlabHost:   "https://YOUR_HOSTNAME",
		DebugMode:    false,
		OpenBrowser:  true,
	}
)

// Default get default config which is embedded in the source file, so that
// this program could run without any configuration file.
func Default() *types.Config {
	return defaultConf
}

func DefaultConfPath() string {
	// generate default config directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("get user home failed: %v", err)
	}

	configDirectory := filepath.Join(home, _defaultConfigDirectoryName)
	if _, err = os.Stat(configDirectory); err == nil {
		return configDirectory
	}

	// check directory exists or not.
	if os.IsNotExist(err) {
		// could not find the directory, then mkdir
		if err = os.MkdirAll(configDirectory, 0777); err != nil {
			panic(err)
		}
	}

	return configDirectory
}

var (
	_defaultCWD     string
	_defaultCwdOnce sync.Once
)

// DefaultCWD returns the working directory of current project, default cwd is from
// git rev-parse --show-toplevel command, but if the command could not execute successfully,
// `pwd` command will be used instead.
func DefaultCWD() string {
	_defaultCwdOnce.Do(func() {
		w := bytes.NewBuffer(nil)
		if err := pkg.RunOutput("git rev-parse --show-toplevel", w); err != nil {
			log.Debug("pre-exec 'git rev-parse --show-toplevel' failed:")
			log.Debugf("%s\n", err)
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

const (
	_defaultConfigDirectoryName = ".gitlab-flow" // under user home path.
	_defaultConfigFilename      = "config.toml"
)

// precheckConfigDirectory could parse filename and path from configDirectory.
func precheckConfigDirectory(configDirectory string) string {
	fi, err := os.Stat(configDirectory)
	if err != nil {
		log.Fatalf("could not stat config file: %v", err)
		panic("could not reach")
	}

	if fi.IsDir() {
		return filepath.Join(configDirectory, _defaultConfigFilename)
	}

	return configDirectory
}
