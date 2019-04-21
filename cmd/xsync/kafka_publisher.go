package main

import (
	"strings"
	"time"

	"github.com/sah4ez/xsync/pkg/config"
	"github.com/sah4ez/xsync/pkg/kafka"
	stdkafka "github.com/segmentio/kafka-go"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

// Publisher return a command with kafka subscriber
func Publisher(cfg *config.Config, log *zap.Logger) cli.Command {
	return cli.Command{
		Name:        "kafka-publisher",
		Aliases:     []string{"kp"},
		Description: "publish on topic in kafka",
		Action: func(c *cli.Context) error {
			log.Info("start publishing", zap.String("kafka-addr", cfg.Kafka.Addr))

			w := stdkafka.NewWriter(stdkafka.WriterConfig{
				Brokers:           strings.Split(cfg.Kafka.Addr, ";"),
				Topic:             cfg.Kafka.Topic,
				Balancer:          &stdkafka.RoundRobin{},
				BatchSize:         50,
				RebalanceInterval: 1 * time.Second,
			})

			b := kafka.NewBinlogWriter(
				w,
				cfg.Binlog.ServerID,
				cfg.Binlog.Host,
				cfg.Binlog.Port,
				cfg.Binlog.User,
				cfg.Binlog.Password,
				cfg.Schemas,
				cfg.Binlog.GTID,
				cfg.Binlog.Position,
				log,
			)
			b.Run()
			return nil
		},
	}
}
