package kafka

import (
	"context"

	"go.uber.org/zap"
)

type MessagesLogger struct {
	log  *zap.Logger
	next Msg
}

func (ml *MessagesLogger) Push(ctx context.Context, cmd CommandSQL, query []byte) {
	ml.log.Debug("receive command",
		zap.Any("cmd", cmd),
	)
	ml.next.Push(ctx, cmd, query)
}

func NewMessagesLogger(msg Msg, log *zap.Logger) *MessagesLogger {
	return &MessagesLogger{
		next: msg,
		log:  log,
	}
}
