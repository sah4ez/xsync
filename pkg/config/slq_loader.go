package config

import (
	"fmt"

	"github.com/siddontang/go-mysql/client"
)

var (
	SettingsTable = "transaction_base.xsync_settings"
)

type ConfigSQL struct {
	Settings *client.Conn
	Schemas  map[string][]Table
	Tables   map[string]uint64
}

var getSettings = "SELECT value FROM %s WHERE key_id='%s'"
var setSettings = "INSERT INTO %s (key_id, value) VALUES (%s, %s) ON DUPLICATE KEY UPDATE value=VALUES(value)"

func (cs *ConfigSQL) Load() (*Config, error) {
	c := &Config{}
	for schema, table := range cs.Schemas {
		for _, t := range table {
			q := fmt.Sprintf(getSettings, SettingsTable, schema+"."+t.Table)
			fmt.Println(">>>", schema, t, q)
			v, err := cs.Settings.Execute(q)
			fmt.Println(">>>>>", schema, t, q, err.Error())
			if err != nil {
				return nil, err
			}

			if v.ColumnNumber() != 1 || v.RowNumber() != 1 {
				return nil,
					fmt.Errorf("invalid cound column %d or rows %d for settings %s",
						v.ColumnNumber(), v.RowNumber(), schema+"."+t.Table)
			}
			vv, err := v.GetUint(0, 0)
			if err != nil {
				return nil, err
			}
			cs.Tables[schema+"."+t.Table] = vv
		}
	}
	return c, nil
}

func (cs *ConfigSQL) SetSQL(schema string, table string, value string) error {
	q := fmt.Sprintf(setSettings, SettingsTable, "'"+schema+"."+table+"'", value)
	_, err := cs.Settings.Execute(q)
	if err != nil {
		return err
	}
	return nil
}
