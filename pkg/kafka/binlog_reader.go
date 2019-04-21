package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/sah4ez/xsync/pkg/builder"
	stdkafka "github.com/segmentio/kafka-go"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"go.uber.org/zap"
)

type BinlogReader struct {
	reader  *stdkafka.Reader
	Target  *client.Conn
	columns map[string][]string
	logger  *zap.Logger
}

func (b *BinlogReader) Run() {
	for {
		ctx := context.Background()
		m, err := b.reader.FetchMessage(ctx)
		if err != nil {
			b.logger.Error("Get event ",
				zap.String("cmd", string(m.Key)),
				zap.String("err", err.Error()),
			)
			continue
		}
		data := m.Value
		cmd := CommandSQL(string(m.Key))
		switch cmd {
		case Transaction, Rollback, Commit:
			b.Target.Execute(string(data))
		case Insert:
			insert := builder.Insert()
			err = insert.UnmarshallBinary(data)
			if err != nil {
				break
			}
			columns, err := b.LoadColumnsForTable(insert.Tables)
			if err != nil {
				break
			}
			insert = insert.Column(columns...)
			insertStr, err := insert.ToSql()
			if err != nil {
				break
			}

			_, err = b.Target.Execute(insertStr)
			if err != nil {
				if strings.Contains(err.Error(), fmt.Sprintf("%d", mysql.ER_DUP_ENTRY)) {
					continue
				}
				b.logger.Error("execute insert query",
					zap.String("query", insertStr),
					zap.String("err", err.Error()),
				)
				b.Target.Execute("ROLLBACK;")
				continue
			}
			b.logger.Info("successful insert query", zap.String("query", insertStr))
		case Delete:
			del := builder.Delete()
			err = del.UnmarshallBinary(data)
			if err != nil {
				break
			}
			columns, err := b.LoadColumnsForTable(del.Tables)
			if err != nil {
				break
			}
			for i, c := range columns {
				del.Wheres[i] = c + "=" + del.Wheres[i]
				if columns[i] != columns[len(columns)-1] {
					del.Wheres[i] = del.Wheres[i] + " AND "
				}
			}
			delStr, err := del.ToSql()
			if err != nil {
				b.logger.Error("build delete query",
					zap.String("err", err.Error()),
				)
				b.Target.Execute("ROLLBACK;")
				continue
			}
			_, err = b.Target.Execute(delStr)
			if err != nil {
				b.logger.Error("execute delete query",
					zap.String("query", delStr),
					zap.String("err", err.Error()),
				)
				b.Target.Execute("ROLLBACK;")
				continue
			}
			b.logger.Info("successful delete query", zap.String("query", delStr))

		case Update:
			update := builder.Update()
			err = update.UnmarshallBinary(data)
			if err != nil {
				break
			}
			columns, err := b.LoadColumnsForTable(update.Tables)
			if err != nil {
				break
			}
			for i, c := range columns {
				update.Wheres[i] = c + "=" + update.Wheres[i]
				if columns[i] != columns[len(columns)-1] {
					update.Wheres[i] = update.Wheres[i] + " AND "
				}
			}
			updateStr, err := update.ToSql()
			if err != nil {
				b.logger.Error("build update query",
					zap.String("err", err.Error()),
				)
				b.Target.Execute("ROLLBACK;")
				continue
			}
			_, err = b.Target.Execute(updateStr)
			if err != nil {
				b.logger.Error("execute update query",
					zap.String("query", updateStr),
					zap.String("err", err.Error()),
				)
				b.Target.Execute("ROLLBACK;")
				continue
			}
			b.logger.Info("successful update query",
				zap.String("query", updateStr))
		case Select:
		default:
		}
		if err != nil {
			b.logger.Error("Execute",
				zap.String("cmd", string(m.Key)),
				zap.String("err", err.Error()),
			)
			continue
		}

		b.reader.CommitMessages(ctx, m)

	}
	b.reader.Close()
}

func (b *BinlogReader) LoadColumnsForTable(fullTable string) ([]string, error) {
	if c, ok := b.columns[fullTable]; ok {
		return c, nil
	}

	parts := strings.Split(fullTable, ".")

	columnSelect := builder.Select().
		Column("COLUMN_NAME").
		From("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_NAME='" + parts[1] + "' AND TABLE_SCHEMA='" + parts[0] + "'")
	columnStr, err := columnSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build column select query %s", err.Error())
	}

	v, err := b.Target.Execute(columnStr)
	if err != nil {
		return nil, fmt.Errorf("execute select column query %s", err.Error())
	}
	var vv []string

	if v.RowNumber() < 1 {
		return nil, fmt.Errorf("invalid count column %d", v.RowNumber())
	}

	for i := 0; i < v.RowNumber(); i++ {
		str, _ := v.GetString(i, 0)
		vv = append(vv, str)
	}
	b.columns[fullTable] = vv
	return vv, nil
}

func NewBinlogReader(tgt *client.Conn, reader *stdkafka.Reader, logger *zap.Logger) *BinlogReader {
	return &BinlogReader{
		Target:  tgt,
		reader:  reader,
		columns: make(map[string][]string),
		logger:  logger,
	}
}
