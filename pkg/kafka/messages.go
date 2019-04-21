package kafka

import (
	"context"

	stdkafka "github.com/segmentio/kafka-go"
)

type CommandSQL string

var (
	Transaction CommandSQL = "transaction"
	Commit      CommandSQL = "commit"
	Rollback    CommandSQL = "rollback"
	Insert      CommandSQL = "insert"
	Delete      CommandSQL = "delete"
	Update      CommandSQL = "update"
	Select      CommandSQL = "select"
)

type Msg interface {
	Push(ctx context.Context, cmd CommandSQL, query []byte)
}

type Messages struct {
	writer *stdkafka.Writer
	msgs   []stdkafka.Message
}

func (m *Messages) Push(ctx context.Context, cmd CommandSQL, query []byte) {
	if m.msgs == nil {
		m.msgs = make([]stdkafka.Message, 0)
	}

	switch cmd {
	case Transaction:
		m.msgs = append(m.msgs, message(cmd, query))
	case Rollback:
		m.msgs = nil
	case Commit:
		m.msgs = append(m.msgs, message(cmd, query))
		m.writer.WriteMessages(ctx, m.msgs...)
		m.msgs = nil
	case Select:
		// just skip select
	default:
		if len(m.msgs) > 0 {
			m.msgs = append(m.msgs, message(cmd, query))
		} else {
			m.writer.WriteMessages(ctx, message(cmd, query))
		}
	}
}

func message(cmd CommandSQL, query []byte) stdkafka.Message {
	return stdkafka.Message{
		Key:   []byte(cmd),
		Value: query,
	}
}

func NewMessages(writer *stdkafka.Writer) *Messages {
	return &Messages{
		writer: writer,
	}
}
