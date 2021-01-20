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

// Load to load config from confPath with specified parser.
func Load(confPath string, parser ConfigParser) (cfg *types.Config, err error) {
	if parser == nil {
		parser = NewTOMLParser()
	}

	var (
		r io.Reader
	)
	cfg = new(types.Config)
	p := precheckConfigDirectory(confPath)
	r, err = os.OpenFile(p, os.O_RDONLY, 0777)
	err = parser.Unmarshal(r, cfg)

	return cfg, err
}

// Save to save config with specified parser.
func Save(confPath string, cfg *types.Config, parser ConfigParser) error {
	if parser == nil {
		parser = NewTOMLParser()
	}

	data, err := parser.Marshal(cfg)
	if err != nil {
		return err
	}

	p := precheckConfigDirectory(confPath)
	w, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, string(data))
	return err
}

var (
	defaultConf = &types.Config{
		AccessToken:  "",
		DebugMode:    false,
		GitlabAPIURL: "https://YOUR_HOSTNAME/api/v4",
		OpenBrowser:  true,
	}
)

// Default .
func Default() *types.Config {
	return defaultConf
}

// DefaultConfPath
func DefaultConfPath() string {
	// generate default config directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("get user home failed: %v", err)
	}

	configDirectory := filepath.Join(home, _defaultConfigDirectory)
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

func DefaultCWD() string {
	cwd, _ := os.Getwd()
	return cwd
}

const (
	_defaultConfigDirectory = ".gitlab-flow" // under user home path.
	_configFilename         = "config.toml"
)

// precheckConfigDirectory could parse filename and path from configDirectory.
func precheckConfigDirectory(configDirectory string) string {
	fi, err := os.Stat(configDirectory)
	if err != nil {
		log.Fatalf("could not stat config file: %v", err)
		panic("could not reach")
	}

	if fi.IsDir() {
		return filepath.Join(configDirectory, _configFilename)
	}

	return configDirectory
}
