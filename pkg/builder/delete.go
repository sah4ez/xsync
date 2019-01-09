package builder

import (
	"fmt"
	"strings"
)

type DeleteBuilder struct {
	Tables string
	Wheres []string
}

func Delete() DeleteBuilder {
	return DeleteBuilder{}
}

func (b DeleteBuilder) Table(table string) DeleteBuilder {
	b.Tables = table
	return b
}

func (b DeleteBuilder) Where(cond ...string) DeleteBuilder {
	b.Wheres = append(b.Wheres, cond...)
	return b
}

func (b DeleteBuilder) ToSql() (string, error) {
	d := "DELETE FROM "
	if b.Tables == "" {
		return d, fmt.Errorf("missing delete table value")
	}
	d += b.Tables

	if len(b.Wheres) == 0 {
		return d, fmt.Errorf("empty where statement")
	}
	d += " WHERE " + strings.Join(b.Wheres, " ")

	return d, nil
}
