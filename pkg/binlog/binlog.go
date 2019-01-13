package binlog

import (
	"context"
	"fmt"
	"strings"

	"github.com/sah4ez/xsync/pkg/builder"
	"github.com/sah4ez/xsync/pkg/config"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"go.uber.org/zap"
)

type Binlog struct {
	cfg      replication.BinlogSyncerConfig
	Target   *client.Conn
	Tables   map[string][]config.Table
	gtid     string
	position string
	columns  map[string][]string
	logger   *zap.Logger
}

func (b *Binlog) Run() {
	syncer := replication.NewBinlogSyncer(b.cfg)
	gtid, _ := mysql.ParseGTIDSet("mysql", b.gtid+":"+b.position)
	streamer, _ := syncer.StartSyncGTID(gtid)
	for {
		ev, err := streamer.GetEvent(context.Background())
		if err != nil {
			b.logger.Error("Get event ",
				zap.String("err", err.Error()),
			)
			continue
		}
		// Debug
		// fmt.Printf(">>> %s\n", ev.Header.EventType)
		// fmt.Printf(">>>> %#v\n", B2S(ev.RawData))
		// fmt.Printf(">>>>> %#v\n", ev)
		// ev.Dump(os.Stdout)
		if e, ok := ev.Event.(*replication.RowsEvent); ok {
			cur := config.Table{}

			schema := B2S(e.Table.Schema)
			table := B2S(e.Table.Table)
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
			columns, err := b.LoadColumnsForTable(schema, table)
			if err != nil {
				b.logger.Error("load columns for table",
					zap.String("binlog", fmt.Sprintf("%+v", e)),
					zap.String("err", err.Error()),
				)
				continue
			}

			switch ev.Header.EventType {
			case replication.WRITE_ROWS_EVENTv2:
				if err != nil {
					b.logger.Error("execute insert query",
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}
				insert := builder.Insert().
					Table(fullTable).
					Column(columns...)
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
				insertStr, err := insert.ToSql()
				if err != nil {
					b.logger.Error("build insert query",
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}

				_, err = b.Target.Execute(insertStr)
				if err != nil {
					if strings.Contains(err.Error(), fmt.Sprintf("%d", mysql.ER_DUP_ENTRY)) {
						continue
					}
					b.logger.Error("execute insert query",
						zap.String("query", insertStr),
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}
				b.logger.Info("successful insert query",
					zap.String("query", insertStr))
			case replication.UPDATE_ROWS_EVENTv2:
				if len(e.Rows) != 2 {
					b.logger.Error("invalid cound rows for update",
						zap.String("binlog", fmt.Sprintf("%+v", e)),
					)
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

				wheres := []string{}
				for i, c := range columns {
					wheres = append(wheres, c+"="+oldRow[i])
				}

				update := builder.Update().
					Table(fullTable).
					Column(columns...).
					Value(newRow...).
					Where(strings.Join(wheres, " AND "))

				updateStr, err := update.ToSql()
				if err != nil {
					b.logger.Error("build update query",
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}
				_, err = b.Target.Execute(updateStr)
				if err != nil {
					b.logger.Error("execute update query",
						zap.String("query", updateStr),
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}
				b.logger.Info("successful update query",
					zap.String("query", updateStr))

			case replication.DELETE_ROWS_EVENTv2:
				b.logger.Debug("delete bin log event",
					//	zap.String("query", updateStr),
					zap.String("binlog", fmt.Sprintf("%+v", e)),
				//	zap.String("err", err.Error()),
				)
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

				wheres := []string{}
				for i, c := range columns {
					wheres = append(wheres, c+"="+strRows[i])
				}
				del := builder.Delete().
					Table(fullTable).
					Where(strings.Join(wheres, " AND "))

				delStr, err := del.ToSql()
				if err != nil {
					b.logger.Error("build delete query",
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}
				_, err = b.Target.Execute(delStr)
				if err != nil {
					b.logger.Error("execute delete query",
						zap.String("query", delStr),
						zap.String("binlog", fmt.Sprintf("%+v", e)),
						zap.String("err", err.Error()),
					)
					continue
				}
				b.logger.Info("successful delete query",
					zap.String("query", delStr))

			default:
				b.logger.Debug("unsupported type", zap.String("event_type", string(ev.Header.EventType)))
			}
		}
	}
}

func (b *Binlog) LoadColumnsForTable(schema, table string) ([]string, error) {
	if c, ok := b.columns[schema+"."+table]; ok {
		return c, nil
	}

	columnSelect := builder.Select().
		Column("COLUMN_NAME").
		From("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_NAME='" + table + "' AND TABLE_SCHEMA='" + schema + "'")
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
	b.columns[schema+"."+table] = vv
	return vv, nil
}

func B2S(bs []uint8) string {
	ba := make([]byte, 0, len(bs))
	for _, b := range bs {
		ba = append(ba, byte(b))
	}
	return string(ba)
}

func NewBinlog(tgt *client.Conn, serverId uint32, host string, port uint16, user, password string, t map[string][]config.Table, gtid, position string, logger *zap.Logger) *Binlog {
	return &Binlog{
		Target: tgt,
		cfg: replication.BinlogSyncerConfig{
			ServerID: serverId,
			Flavor:   "mysql",
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
		},
		Tables:   t,
		gtid:     gtid,
		position: position,
		columns:  make(map[string][]string),
		logger:   logger,
	}
}
