package main

import (
	"os"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/yeqown/log"
)

func getFlow(confPath string, debug bool) internal.IFlow {
	cfg, err := conf.Load(confPath, nil)
	if err != nil {
		log.
			Fatalf("could not load config file from %s", confPath)
		panic("could not reach")
	}

	if err = cfg.Debug(debug).Valid(); err != nil {
		log.
			WithField("config", cfg).
			Fatalf("config is invalid: %s", confPath)
		panic("could not reach")
	}

	// DONE(@yeqown) get cwd correctly.
	cwd, _ := os.Getwd()
	return internal.NewFlow(types.NewContext(cwd, confPath, cfg))
}

func getDash(confPath string, debug bool) internal.IDash {
	return internal.NewDash(confPath, debug)
}
