package conf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yeqown/gitlab-flow/internal/types"
	"github.com/yeqown/log"
)

// ConfigParser is an interface to parse config in different ways.
// For example: JSON, TOML and YAML;
type ConfigParser interface {
	// Unmarshal ...
	Unmarshal(r io.Reader, rcv *types.Config) error

	// Marshal ...
	Marshal(cfg *types.Config) ([]byte, error)
}

// Load to load config from confpath with specified parser.
func Load(confpath string, parser ConfigParser) (cfg *types.Config, err error) {
	if parser == nil {
		parser = NewTOMLParser()
	}

	var (
		r io.Reader
	)
	cfg = new(types.Config)
	p, create := precheckConfigDirectory(confpath)
	if create {
		cfg = defaultConf
		if err = Save(p, cfg, parser); err != nil {
			return nil, fmt.Errorf("init config file failed: %v", err)
		}
	}

	r, err = os.OpenFile(p, os.O_RDONLY, 0777)
	err = parser.Unmarshal(r, cfg)

	return cfg, err
}

// Save to save config with specified parser.
func Save(confpath string, cfg *types.Config, parser ConfigParser) error {
	if parser == nil {
		parser = NewTOMLParser()
	}

	data, err := parser.Marshal(cfg)
	if err != nil {
		return err
	}
	w, err := os.OpenFile(confpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, data)
	return err
}

var (
	defaultConf = &types.Config{
		AccessToken:       "",
		DebugMode:         false,
		GitlabAPIURL:      "https://YOUR_HOSTNAME/api/v4",
		OpenWebCommandTpl: "open -a Safari %s",
	}
)

const (
	_defaultConfigDirectory = ".gitlab-flow"
	_configFilename         = "config.toml"
)

// precheckConfigDirectory could parse filename and path from configDirectory.
func precheckConfigDirectory(configDirectory string) (s string, create bool) {
	if configDirectory != "" {
		return configDirectory, false
	}

	// generate default config directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("get user home failed: %v", err)
	}

	configDirectory = filepath.Join(home, _defaultConfigDirectory)
	s = filepath.Join(configDirectory, _configFilename)
	if _, err = os.Stat(configDirectory); err == nil {
		return s, false
	}

	// check directory exists or not.
	if os.IsNotExist(err) {
		// could not find the directory, then mkdir
		if err = os.MkdirAll(configDirectory, 0777); err != nil {
			panic(err)
		}
		return s, true
	}

	// other errors while stat config directory.
	panic(err)
}
