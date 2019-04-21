package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
)

type OrderType string

var (
	Asc  OrderType = "ASC"
	Desc OrderType = "DESC"
)

type SelectBuilder struct {
	Columns  []string  `json:"columns"`
	Froms    []string  `json:"froms"`
	Wheres   []string  `json:"wheres"`
	LimitV   uint      `json:"limit_v"`
	OrderByV OrderType `json:"order_by_v"`
	OrderByF []string  `json:"order_by_f"`
}

func Select() SelectBuilder {
	return SelectBuilder{}
}

func (s SelectBuilder) Column(columns ...string) SelectBuilder {
	s.Columns = append(s.Columns, columns...)
	return s
}

func (s SelectBuilder) From(tables ...string) SelectBuilder {
	s.Froms = append(s.Froms, tables...)
	return s
}

func (s SelectBuilder) Where(cond ...string) SelectBuilder {
	s.Wheres = append(s.Wheres, cond...)
	return s
}

func (s SelectBuilder) OrderBy(t OrderType, cond ...string) SelectBuilder {
	s.OrderByV = t
	s.OrderByF = append(s.OrderByF, cond...)
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

func (b SelectBuilder) MarshallBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(b)
	return buf.Bytes(), nil
}

func (b *SelectBuilder) UnmarshallBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(b)
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

	if len(s.OrderByF) > 0 {
		ss += " ORDER BY "
		ss += strings.Join(s.OrderByF, ",")
		ss += " " + string(s.OrderByV)
	}

	if s.LimitV > 0 {
		ss += " LIMIT " + fmt.Sprintf("%d", s.LimitV)
	}

	return ss, nil
}
