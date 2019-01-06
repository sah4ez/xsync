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
						q.Pool.Schedule(
							func() error {
								t := table
								fmt.Println(">>>>>", schema, t)
								selectStr := "SELECT * FROM " + schema + "." + t.Table +
									" WHERE " + t.FieldID + ">" + t.Latest +
									" LIMIT " + t.Batch

								q.sl.Lock()
								v, err := q.Source.Execute(selectStr)
								q.sl.Unlock()
								if err != nil {
									return fmt.Errorf("source query: %s has error: %s", selectStr, err.Error())
								}

								if v != nil && v.Resultset != nil {
									insertStr := "INSERT INTO " + schema + "." + t.Table
									insertStr += "("

									fields := make([]string, len(v.Resultset.FieldNames))
									for field, i := range v.Resultset.FieldNames {
										fields[i] = field
									}

									count := 0
									for _, field := range fields {
										insertStr += field
										if len(v.Resultset.FieldNames)-1 != count {
											insertStr += ","
										}
										count += 1
									}

									insertStr += ") VALUES "

									for _, vvv := range v.Resultset.RowDatas {
										newLine := " " + insertStr
										vv, err := vvv.ParseText(v.Resultset.Fields)
										if err != nil {
											return fmt.Errorf("parse query: %s has error: %s", selectStr, err.Error())
										}

										newLine += "("
										onDuplicate := " ON DUPLICATE KEY UPDATE  "
										for ii, val := range vv {
											valUint8, ok := val.([]uint8)
											if ok {
												valUint8str := B2S(valUint8)
												intValue := ""
												if _, err := strconv.Atoi(valUint8str); err != nil {
													intValue = "'" + valUint8str + "'"
												} else {
													intValue = valUint8str
												}
												newLine += intValue
												if fields[ii] != t.FieldID {
													onDuplicate += fields[ii] + "=" + intValue

													if ii < len(vv)-1 {
														onDuplicate += ","
													}
												}
											}
											if ii < len(vv)-1 {
												newLine += ","
											}
										}
										newLine += ")" + onDuplicate
										q.tl.Lock()
										_, err = q.Target.Execute(newLine)
										q.tl.Lock()
										if err != nil {
											return fmt.Errorf("insert query: %s has error: %s", newLine, err.Error())
										}
									}
								}
								return nil
							},
						)
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
