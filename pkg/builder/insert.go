package builder

import (
	"fmt"
	"strings"
)

type InsertBuilder struct {
	Tabels      string
	Columns     []string
	Values      [][]string
	OnDuplicate []string
}

func Insert() InsertBuilder {
	return InsertBuilder{}
}

func (i InsertBuilder) Table(table string) InsertBuilder {
	i.Tabels = table
	return i
}

func (i InsertBuilder) Column(columns ...string) InsertBuilder {
	i.Columns = append(i.Columns, columns...)
	return i
}

func (i InsertBuilder) Value(vals ...string) InsertBuilder {
	var val []string
	val = append(val, vals...)
	i.Values = append(i.Values, val)
	return i
}

func (i InsertBuilder) OnDuplicateKeyUpdate(columns ...string) InsertBuilder {
	for _, c := range columns {
		i.OnDuplicate = append(i.OnDuplicate, fmt.Sprintf("%s=VALUES(%s)", c, c))
	}
	return i
}

func (i InsertBuilder) ToSql() (string, error) {
	ii := "INSERT INTO "
	if i.Tabels == "" {
		return ii, fmt.Errorf("missing insert table value")
	}

	ii += i.Tabels
	if len(i.Columns) != 0 {
		ii += " (" + strings.Join(i.Columns, ",") + ") "
	}

	if len(i.Values) > 0 {
		pv := []string{}
		for _, v := range i.Values {
			if len(i.Columns) == 0 || len(v) == len(i.Columns) {
				pv = append(pv, " ("+strings.Join(v, ",")+") ")
			} else {
				return ii, fmt.Errorf("invalid count of values %v for insert %v", v, i.Columns)
			}
		}
		ii += " VALUES " + strings.Join(pv, ",")
	} else {
		return ii, fmt.Errorf("missing insert values")
	}

	if len(i.OnDuplicate) > 0 {
		ii += " ON DUPLICATE KEY UPDATE " + strings.Join(i.OnDuplicate, ",")
	}
	return ii, nil
}
