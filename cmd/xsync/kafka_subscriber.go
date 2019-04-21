package main

import (
	"strings"

	"github.com/sah4ez/xsync/pkg/config"
	"github.com/sah4ez/xsync/pkg/kafka"
	stdkafka "github.com/segmentio/kafka-go"
	"github.com/siddontang/go-mysql/client"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

// Subscriber return a command with kafka subscriber
func Subscriber(cfg *config.Config, log *zap.Logger) cli.Command {
	return cli.Command{
		Name:        "kafka-subscriber",
		Aliases:     []string{"ks"},
		Description: "subscribe on topic in kafka",
		Action: func(c *cli.Context) error {
			log.Debug("start kafka subscriber", zap.String("kafka-addr", cfg.Kafka.Addr))
			r := stdkafka.NewReader(stdkafka.ReaderConfig{
				Brokers:   strings.Split(cfg.Kafka.Addr, ";"),
				Partition: cfg.Kafka.Partition,
				Topic:     cfg.Kafka.Topic,
				MaxWait:   cfg.Kafka.MaxWait,
				MinBytes:  cfg.Kafka.MinBytes,
				MaxBytes:  cfg.Kafka.MaxBytes,

				CommitInterval: cfg.Kafka.CommitInterval,
			})
			r.SetOffset(cfg.Kafka.Offset)

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

			b := kafka.NewBinlogReader(targetConn, r, log)
			b.Run()

			return nil
		},
	}
}
