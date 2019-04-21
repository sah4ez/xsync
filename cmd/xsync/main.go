package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sah4ez/xsync/pkg/config"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

var (
	Revision = ""
	Version  = ""
)

var (
	configFlag = flag.String("config", "./config.yaml", "Loading configuration from source")
)

func init() {
	flag.Parse()
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("couldn't create logger %s", err.Error()))
	}
	defer logger.Sync()

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version: %s\nrevision: %s\n", c.App.Version, Revision)
	}

	app := cli.NewApp()
	app.Name = "xsync"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
			Value: "/home/sah4ez/go/src/github.com/sah4ez/xsync/config.yaml",
		},
	}
	var cfgLoader config.Cfg

	if *configFlag != "" {
		val := *configFlag
		if strings.Contains(val, "yaml") {
			cfgLoader = &config.ConfigYAML{Path: val}
		}
	}

	cfg, err := cfgLoader.Load()
	if err != nil {
		logger.Panic("couldn't load config", zap.String("err", err.Error()))
		os.Exit(1)
	}

	app.Commands = []cli.Command{
		Batching(cfg, logger),
		Binlog(cfg, logger),
		Publisher(cfg, logger),
		Subscriber(cfg, logger),
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Error("Couldn't start application", zap.String("err", err.Error()))
	}
}
