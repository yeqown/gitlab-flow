// Package conf implements the configuration of the application.
// The configuration is loaded from the configuration file. The configuration
// file is a TOML file.
package conf

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"github.com/yeqown/log"

	"github.com/yeqown/gitlab-flow/internal/types"
)

const (
	DefaultScopes       = "api read_user read_repository"
	DefaultCallbackHost = "localhost:2333"

	defaultConfigDirectoryName = ".gitlab-flow" // under user home path.
	defaultConfigFilename      = "config.toml"
)

var (
	defaultConf = &types.Config{
		Branch: &types.BranchSetting{
			Master:                      types.MasterBranch,
			Dev:                         types.DevBranch,
			Test:                        types.TestBranch,
			FeatureBranchPrefix:         types.FeatureBranchPrefix,
			HotfixBranchPrefix:          types.HotfixBranchPrefix,
			ConflictResolveBranchPrefix: types.ConflictResolveBranchPrefix,
			IssueBranchPrefix:           types.IssueBranchPrefix,
		},
		OAuth2: &types.OAuth{
			Scopes:       DefaultScopes,
			CallbackHost: DefaultCallbackHost,
			AccessToken:  "",
			RefreshToken: "",
		},
		GitlabAPIURL: "https://YOUR_HOSTNAME/api/v4",
		GitlabHost:   "https://YOUR_HOSTNAME",
		DebugMode:    false,
		OpenBrowser:  true,
	}

	//go:embed config.tpl
	configTemplateContent string
	configTpl             = template.Must(template.New("config").Parse(configTemplateContent))

	//go:embed config.project.tpl
	projectConfigTemplateContent string
	projectConfigTpl             = template.Must(template.New("project_config").Parse(projectConfigTemplateContent))
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
func Load(confPath string, parser Parser, must bool) (cfg *types.Config, err error) {
	if parser == nil {
		parser = NewTOMLParser()
	}

	cfg = &types.Config{
		OAuth2:       new(types.OAuth),
		Branch:       new(types.BranchSetting),
		GitlabAPIURL: "",
		GitlabHost:   "",
		DebugMode:    false,
		OpenBrowser:  false,
	}
	p := precheckConfigDirectory(confPath)
	var r io.Reader
	if r, err = os.Open(p); err != nil {
		if !must && os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "conf.Load")
	}

	err = parser.Unmarshal(r, cfg)

	return cfg, err
}

func Save(confPath string, cfg *types.Config, global bool) error {
	p := precheckConfigDirectory(confPath)
	w, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "open config file")
	}
	defer func() {
		_ = w.Close()
	}()

	var tpl *template.Template
	if global {
		tpl = configTpl
	} else {
		tpl = projectConfigTpl
	}

	if err = tpl.Execute(w, cfg); err != nil {
		return errors.Wrap(err, "execute template")
	}

	return err
}

// Default get default config which is embedded in the source file, so that
// this program could run without any configuration file.
func Default() *types.Config {
	return defaultConf
}

func ConfigPath(parent string) string {
	if parent == "" {
		// generate default config directory
		home, err := os.UserHomeDir()
		if err != nil {
			log.Errorf("get user home failed: %v", err)
		}
		parent = home
	}

	var err error
	configDirectory := filepath.Join(parent, defaultConfigDirectoryName)
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

// precheckConfigDirectory could parse filename and path from configDirectory.
func precheckConfigDirectory(configDirectory string) string {
	fi, err := os.Stat(configDirectory)
	if err != nil {
		log.Fatalf("could not stat config file: %v", err)
		panic("could not reach")
	}

	if fi.IsDir() {
		return filepath.Join(configDirectory, defaultConfigFilename)
	}

	return configDirectory
}
