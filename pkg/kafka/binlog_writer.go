package kafka

import (
	"context"
	"fmt"

	"github.com/sah4ez/xsync/pkg/builder"
	"github.com/sah4ez/xsync/pkg/config"
	"github.com/sah4ez/xsync/pkg/utils"
	stdkafka "github.com/segmentio/kafka-go"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"go.uber.org/zap"
)

type BinlogWriter struct {
	cfg      replication.BinlogSyncerConfig
	writer   *stdkafka.Writer
	Tables   map[string][]config.Table
	gtid     string
	position string
	logger   *zap.Logger
}

func (b *BinlogWriter) Run() {
	syncer := replication.NewBinlogSyncer(b.cfg)
	gtid, _ := mysql.ParseGTIDSet("mysql", b.gtid+":"+b.position)
	streamer, _ := syncer.StartSyncGTID(gtid)
	messages := NewMessagesLogger(NewMessages(b.writer), b.logger)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	for {
		ev, err := streamer.GetEvent(ctx)
		if err != nil {
			b.logger.Error("Get event ", zap.String("err", err.Error()))
			continue
		}
		switch ev.Header.EventType {
		case replication.GTID_EVENT:
			messages.Push(ctx, Transaction, []byte("START TRANSACTION;"))
			continue
		case replication.XID_EVENT:
			messages.Push(ctx, Commit, []byte("COMMIT;"))
			continue
		}

		if e, ok := ev.Event.(*replication.RowsEvent); ok {
			cur := config.Table{}

			schema := utils.Byte2Sring(e.Table.Schema)
			table := utils.Byte2Sring(e.Table.Table)
			if ts, ok := b.Tables[schema]; ok {
				for _, t := range ts {
					if t.Table == table {
						cur = t
					}
				}
			}
			if cur == config.NilTable {
				b.logger.Error("missed",
					zap.String("schema", schema),
					zap.String("table", table))
				continue
			}

			fullTable := schema + "." + table

			switch ev.Header.EventType {
			case replication.WRITE_ROWS_EVENTv2:
				insert := builder.Insert().Table(fullTable)
				for _, row := range e.Rows {
					var strRow []string
					for _, i := range row {
						if i == nil {
							strRow = append(strRow, fmt.Sprintf("%v", "NULL"))
							continue
						}

						if str, ok := i.(string); ok {
							strRow = append(strRow, fmt.Sprintf("'%v'", str))
						} else {
							strRow = append(strRow, fmt.Sprintf("%v", i))
						}
					}
					insert = insert.Value(strRow...)
				}
				insertBytes, _ := insert.MarshallBinary()
				messages.Push(ctx, Insert, insertBytes)

			case replication.UPDATE_ROWS_EVENTv2:
				if len(e.Rows) != 2 {
					b.logger.Error("invalid cound rows for update",
						zap.String("binlog", fmt.Sprintf("%+v", e)),
					)
					messages.Push(ctx, Rollback, nil)
					continue
				}

				var oldRow []string
				var newRow []string

				for _, i := range e.Rows[0] {
					if str, ok := i.(string); ok {
						oldRow = append(oldRow, fmt.Sprintf("'%v'", str))
					} else {
						oldRow = append(oldRow, fmt.Sprintf("%v", i))
					}
				}
				for _, i := range e.Rows[1] {
					if str, ok := i.(string); ok {
						newRow = append(newRow, fmt.Sprintf("'%v'", str))
					} else {
						newRow = append(newRow, fmt.Sprintf("%v", i))
					}
				}

				update := builder.Update().
					Table(fullTable).
					Value(newRow...).
					Where(oldRow...)

				updateBytes, _ := update.MarshallBinary()
				messages.Push(ctx, Update, updateBytes)
			case replication.DELETE_ROWS_EVENTv2:
				if len(e.Rows) == 0 {
					continue
				}

				strRows := []string{}

				for _, row := range e.Rows {
					for _, i := range row {
						if str, ok := i.(string); ok {
							strRows = append(strRows, fmt.Sprintf("'%v'", str))
						} else {
							strRows = append(strRows, fmt.Sprintf("%v", i))
						}
					}
				}

				del := builder.Delete().
					Table(fullTable).
					Where(strRows...)

				deleteBytes, _ := del.MarshallBinary()
				messages.Push(ctx, Delete, deleteBytes)
			default:
				b.logger.Debug("unsupported type", zap.String("event_type", string(ev.Header.EventType)))
			}
		}
	}
}

func NewBinlogWriter(writer *stdkafka.Writer, serverId uint32, host string, port uint16, user, password string, t map[string][]config.Table, gtid, position string, logger *zap.Logger) *BinlogWriter {
	return &BinlogWriter{
		cfg: replication.BinlogSyncerConfig{
			ServerID: serverId,
			Flavor:   "mysql",
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
		},
		writer:   writer,
		Tables:   t,
		gtid:     gtid,
		position: position,
		logger:   logger,
	}
}
