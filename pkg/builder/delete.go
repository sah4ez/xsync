package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
)

type DeleteBuilder struct {
	Tables string   `json:"tables"`
	Wheres []string `json:"wheres"`
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

func (b DeleteBuilder) MarshallBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(b)
	return buf.Bytes(), nil
}

func (b *DeleteBuilder) UnmarshallBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(b)
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
