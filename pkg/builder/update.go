package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
)

type UpdateBuilder struct {
	Tables  string   `json:"tables"`
	Columns []string `json:"columns"`
	Values  []string `json:"values"`
	Wheres  []string `json:"wheres"`
}

func Update() UpdateBuilder {
	return UpdateBuilder{}
}

func (b UpdateBuilder) Table(table string) UpdateBuilder {
	b.Tables = table
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

func (b UpdateBuilder) MarshallBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(b)
	return buf.Bytes(), nil
}

func (b *UpdateBuilder) UnmarshallBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(b)
}

func (b UpdateBuilder) ToSql() (string, error) {
	u := "UPDATE "
	if b.Tables == "" {
		return u, fmt.Errorf("missing update table value")
	}

	u += b.Tables + " SET "

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
