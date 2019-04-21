package main

import (
	"fmt"
	"github.com/sah4ez/xsync/pkg/config"
	"github.com/sah4ez/xsync/pkg/pool"
	"github.com/sah4ez/xsync/pkg/query"
	"github.com/siddontang/go-mysql/client"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func Batching(cfg *config.Config, logger *zap.Logger) cli.Command {
	return cli.Command{
		Name:    "batching",
		Aliases: []string{"b"},
		Usage:   "batching synchronization through sql queries",
		Action: func(c *cli.Context) error {
			isSSL := func(val bool) func(c *client.Conn) {
				if val {
					return func(c *client.Conn) { c.UseSSL(true) }
				}
				return func(c *client.Conn) {}
			}

			sourceConn, err := client.Connect(
				cfg.Source.Addr,
				cfg.Source.User,
				cfg.Source.Password,
				cfg.Source.DB,
				isSSL(cfg.Source.SSL),
			)
			if err != nil {
				return err
			}
			defer sourceConn.Close()
			sourceConn.Ping()

			targetConn, err := client.Connect(
				cfg.Target.Addr,
				cfg.Target.User,
				cfg.Target.Password,
				cfg.Target.DB,
				isSSL(cfg.Target.SSL),
			)
			if err != nil {
				return err
			}
			defer targetConn.Close()
			targetConn.Ping()

			s := make(chan struct{})
			p := pool.New(cfg.Threads, s)
			defer p.Close()

			q := query.NewQuerier(sourceConn, targetConn, p, cfg.Schemas, logger)
			go q.Run()
			<-s
			fmt.Println("batch sync: ", cfg)
			return nil
		},
	}
}
