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
			Mode:         types.OAuth2Mode_Auto,
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

// Load to load config from confPath with specified parser.
func Load(confPath string, cfg types.ConfigHolder, must bool) (err error) {
	p := precheckConfigDirectory(confPath)
	var r io.Reader
	if r, err = os.Open(p); err != nil {
		if !must && os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "conf.Load")
	}

	err = NewTOMLParser().Unmarshal(r, cfg)

	return err
}

func Save(confPath string, c types.ConfigHolder) error {
	p := precheckConfigDirectory(confPath)
	w, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "open config file")
	}
	defer func() {
		_ = w.Close()
	}()

	var tpl *template.Template
	switch c.Type() {
	case types.ConfigType_Project:
		tpl = projectConfigTpl
	default:
		tpl = configTpl
	}

	if err = tpl.Execute(w, c); err != nil {
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
