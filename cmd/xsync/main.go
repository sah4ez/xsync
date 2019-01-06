package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

var (
	Revision = ""
	Version  = ""
)

func main() {
	logger, err := zap.NewProduction()
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
	err = app.Run(os.Args)
	if err != nil {
		logger.Error("Couldn't start application", zap.String("err", err.Error()))
	}
}
