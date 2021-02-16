package traft

import (
	"encoding/json"

	"go.uber.org/zap"
)

var (
	llg = zap.NewNop()
	lg  *zap.SugaredLogger
)

func initLogging() {
	// if os.Getenv("CLUSTER_DEBUG") != "" {
	// }
	var err error
	// llg, err = zap.NewProduction()
	llg, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	lg = llg.Sugar()

	// initZap()

}

func initZap() {
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout", "/tmp/logs"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"foo": "bar"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	var err error
	llg, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	defer llg.Sync()

	llg.Info("logger construction succeeded")

	lg = llg.Sugar()

}
