package query

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sah4ez/xsync/pkg/config"
	"github.com/sah4ez/xsync/pkg/pool"
	"github.com/siddontang/go-mysql/client"
)

type Querier struct {
	sl     sync.Mutex
	Source *client.Conn
	tl     sync.Mutex
	Target *client.Conn
	Pool   *pool.Pool
	Tables map[string][]config.Table
}

func (q *Querier) Run() {
	for schema, tables := range q.Tables {
		for _, t := range tables {
			table := t
			go func() {
				interval := 1 * time.Second
				if table.Interval > time.Duration(0) {
					interval = table.Interval
				}
				for {
					select {
					case <-time.After(interval):
						task := func() error {
							t := table
							fmt.Println(">>>>>", schema, t)
							selectStr := Select().
								Column("*").
								From(schema + "." + t.Table).
								Where(t.FieldID + ">" + t.Latest)

							if t.Batch != "0" {
								selectStr = selectStr.Limit(t.Batch)
							}

							str, err := selectStr.ToSql()
							if err != nil {
								return fmt.Errorf("build query error: %s", err.Error())
							}

							q.sl.Lock()
							v, err := q.Source.Execute(str)
							q.sl.Unlock()
							if err != nil {
								return fmt.Errorf("source query: %s has error: %s", str, err.Error())
							}

							if v != nil && v.Resultset != nil {

								insert := Insert().Table(schema + "." + t.Table)

								fields := make([]string, len(v.Resultset.FieldNames))
								for field, i := range v.Resultset.FieldNames {
									fields[i] = field
								}

								insert = insert.Column(fields...)

								if len(v.Resultset.RowDatas) == 0 {
									return nil
								}
								for _, vvv := range v.Resultset.RowDatas {
									vv, err := vvv.ParseText(v.Resultset.Fields)
									if err != nil {
										return fmt.Errorf("parse query: %s has error: %s", str, err.Error())
									}

									var vals []string
									for _, val := range vv {
										valUint8, ok := val.([]uint8)
										if ok {
											valUint8str := B2S(valUint8)
											intValue := ""
											if _, err := strconv.Atoi(valUint8str); err != nil {
												intValue = "'" + valUint8str + "'"
											} else {
												intValue = valUint8str
											}
											vals = append(vals, intValue)
										}
									}
									insert = insert.Value(vals...)
								}

								var onDuplicate []string
								for _, field := range fields {
									if field != t.FieldID {
										onDuplicate = append(fields, field)
									}
								}
								insert = insert.OnDuplicateKeyUpdate(onDuplicate...)

								insertStr, err := insert.ToSql()
								if err != nil {
									return fmt.Errorf("build insert query has error: %s", err.Error())
								}

								q.tl.Lock()
								_, err = q.Target.Execute(insertStr)
								q.tl.Lock()
								if err != nil {
									return fmt.Errorf("insert query: %s has error: %s", insertStr, err.Error())
								}
							}
							return nil
						}
						q.Pool.Schedule(task)
					}
				}
			}()
		}
	}
}

func B2S(bs []uint8) string {
	ba := make([]byte, 0, len(bs))
	for _, b := range bs {
		ba = append(ba, byte(b))
	}
	return string(ba)
}

func NewQuerier(src, tgt *client.Conn, pool *pool.Pool, tables map[string][]config.Table) *Querier {
	return &Querier{
		Source: src,
		Target: tgt,
		Pool:   pool,
		Tables: tables,
	}
}
