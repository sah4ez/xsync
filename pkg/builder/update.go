package builder

import (
	"fmt"
	"strings"
)

type UpdateBuilder struct {
	Tabels  string
	Columns []string
	Values  []string
	Wheres  []string
}

func Update() UpdateBuilder {
	return UpdateBuilder{}
}

func (b UpdateBuilder) Table(table string) UpdateBuilder {
	b.Tabels = table
	return b
}

func (b UpdateBuilder) Column(columns ...string) UpdateBuilder {
	b.Columns = append(b.Columns, columns...)
	return b
}

func (b UpdateBuilder) Where(cond ...string) UpdateBuilder {
	b.Wheres = append(b.Wheres, cond...)
	return b
}

func (b UpdateBuilder) Value(vals ...string) UpdateBuilder {
	b.Values = append(b.Values, vals...)
	return b
}

func (b UpdateBuilder) ToSql() (string, error) {
	u := "UPDATE "
	if b.Tabels == "" {
		return u, fmt.Errorf("missing insert table value")
	}

	u += b.Tabels + " SET "

	sets := []string{}
	if len(b.Columns) != len(b.Values) {
		return u, fmt.Errorf("invalid count of values %v for update %v", b.Values, b.Columns)
	}
	for i, column := range b.Columns {
		sets = append(sets, column+"="+b.Values[i])
	}
	u += strings.Join(sets, ",")

	if len(b.Wheres) == 0 {
		return u, fmt.Errorf("empty where statement")
	}
	u += " WHERE " + strings.Join(b.Wheres, " ")

	return u, nil
}
