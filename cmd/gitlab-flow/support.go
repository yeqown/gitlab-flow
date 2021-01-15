package main

import (
	"os"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/yeqown/log"
)

func getFlow(confPath string, debug bool) internal.IFlow {
	cfg := setEnviron(confPath, debug)
	// DONE(@yeqown) get cwd correctly.
	cwd, _ := os.Getwd()
	ctx := types.NewContext(cwd, confPath, cfg)
	return internal.NewFlow(ctx)
}

func getDash(confPath string, debug bool) internal.IDash {
	cfg := setEnviron(confPath, debug)
	cwd, _ := os.Getwd()
	ctx := types.NewContext(cwd, confPath, cfg)
	return internal.NewDash(ctx)
}

// setEnviron set global environment of debug mode.
func setEnviron(confPath string, debug bool) *types.Config {
	if !debug {
		log.SetLogLevel(log.LevelInfo)
	} else {
		// open caller report
		log.SetCallerReporter(true)
		log.SetLogLevel(log.LevelDebug)
	}

	log.
		WithFields(log.Fields{
			"confPath": confPath,
			"debug":    debug,
		}).
		Debugf("setEnviron called")

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

	return cfg
}
