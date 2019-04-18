package main

import (
	"github.com/sah4ez/xsync/pkg/binlog"
	"github.com/sah4ez/xsync/pkg/config"
	"github.com/siddontang/go-mysql/client"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func Binlog(cfg *config.Config, logger *zap.Logger) cli.Command {
	return cli.Command{
		Name:        "binlog",
		Aliases:     []string{"bl"},
		Description: "synchronization through binlog",
		Action: func(c *cli.Context) error {
			targetConn, err := client.Connect(
				cfg.Target.Addr,
				cfg.Target.User,
				cfg.Target.Password,
				cfg.Target.DB,
			)
			if err != nil {
				return err
			}
			defer targetConn.Close()
			targetConn.Ping()

			b := binlog.NewBinlog(
				targetConn,
				cfg.Binlog.ServerID,
				cfg.Binlog.Host,
				cfg.Binlog.Port,
				cfg.Binlog.User,
				cfg.Binlog.Password,
				cfg.Schemas,
				cfg.Binlog.GTID,
				cfg.Binlog.Position,
				logger,
			)
			b.Run()
			return nil
		},
	}
}
