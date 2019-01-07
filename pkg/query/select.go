package query

import (
	"fmt"
	"strconv"
	"strings"
)

type SelectBuilder struct {
	Columns []string
	Froms   []string
	Wheres  []string
	LimitV  uint
}

func Select() SelectBuilder {
	return SelectBuilder{}
}

func (s SelectBuilder) Column(columns ...string) SelectBuilder {
	s.Columns = append(s.Columns, columns...)
	return s
}

func (s SelectBuilder) From(tabels ...string) SelectBuilder {
	s.Froms = append(s.Froms, tabels...)
	return s
}

func (s SelectBuilder) Where(cond ...string) SelectBuilder {
	s.Wheres = append(s.Wheres, cond...)
	return s
}

func (s SelectBuilder) Limit(limit string) SelectBuilder {
	v, err := strconv.Atoi(limit)
	if err != nil {
		s.LimitV = 0
	} else {
		s.LimitV = uint(v)
	}
	return s
}

func (s SelectBuilder) ToSql() (string, error) {
	ss := "SELECT "
	if len(s.Columns) > 0 {
		ss += strings.Join(s.Columns, ",")
	} else {
		ss += " * "
	}

	if len(s.Froms) > 0 {
		ss += " FROM "
		ss += strings.Join(s.Froms, ",")
	} else {
		return "", fmt.Errorf("missing FROM statement in select query")
	}

	if len(s.Wheres) > 0 {
		ss += " WHERE "
		ss += strings.Join(s.Wheres, " ")
	}

	if s.LimitV > 0 {
		ss += " LIMIT " + fmt.Sprintf("%d", s.LimitV)
	}

	return ss, nil
}
