package main

import (
	"os"

	"github.com/yeqown/gitlab-flow/internal"
	"github.com/yeqown/gitlab-flow/internal/conf"
	"github.com/yeqown/gitlab-flow/internal/types"

	"github.com/yeqown/log"
)

func getFlow(confPath string, debug bool) internal.IFlow {
	setDebugEnviron(debug)
	log.
		WithFields(log.Fields{
			"confPath": confPath,
			"debug":    debug,
		}).
		Debugf("getFlow called")

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
	ctx := types.NewContext(cwd, confPath, cfg)
	return internal.NewFlow(ctx)
}

func getDash(confPath string, debug bool) internal.IDash {
	setDebugEnviron(debug)
	log.
		WithFields(log.Fields{
			"confPath": confPath,
			"debug":    debug,
		}).
		Debugf("getFlow called")

	return internal.NewDash(confPath, debug)
}

// setDebugEnviron set global environment of debug mode.
func setDebugEnviron(debug bool) {
	log.SetLogLevel(log.LevelInfo)
	if !debug {
		return
	}

	// open caller report
	log.SetCallerReporter(true)
	// open debug log
	log.SetLogLevel(log.LevelDebug)
}
