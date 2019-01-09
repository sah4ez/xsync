package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sah4ez/xsync/pkg/binlog"
	"github.com/sah4ez/xsync/pkg/config"
	"github.com/sah4ez/xsync/pkg/pool"
	"github.com/sah4ez/xsync/pkg/query"
	"github.com/siddontang/go-mysql/client"
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
		{
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
		},
		{
			Name:    "binlog",
			Aliases: []string{"bl"},
			Usage:   "synchronization through binlog",
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

				//connParams := &sqldb.ConnParams{
				//	Host:  cfg.Binlog.Host,
				//	Port:  int(cfg.Binlog.Port),
				//	Uname: cfg.Binlog.User,
				//	Pass:  cfg.Binlog.Password,
				//}

				//				conn, err := mysql.Connect(context.Background(), &connParams)
				//				if err != nil {
				//					fmt.Println(err.Error())
				//					os.Exit(1)
				//				}
				//
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
				//b := binlog.NewBinlogVitess(
				//	conn,
				//	cfg.Binlog.ServerID,
				//	cfg.Binlog.Host,
				//	cfg.Binlog.Port,
				//	cfg.Binlog.User,
				//	cfg.Binlog.Password,
				//	cfg.Schemas,
				//	cfg.Binlog.GTID,
				//	cfg.Binlog.Position,
				//	logger,
				//)
				//bp, err := binlogplayer.NewBinlogPlayerTables(
				//	binlogplayer.NewDbClient(connParams),
				//	nil,
				//	[]string{"users"},
				//	1,
				//	strings.Split(cfg.Binlog.Position, "-")[0],
				//	strings.Split(cfg.Binlog.Position, "-")[1],
				//	binlogplayer.NewStats())
				//if err != nil {
				//	return err
				//}
				//return bp.ApplyBinlogEvents(context.Background())
				return nil
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Error("Couldn't start application", zap.String("err", err.Error()))
	}
}
