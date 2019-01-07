package binlog

import (
	"context"
	"fmt"

	"github.com/sah4ez/xsync/pkg/config"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

type Binlog struct {
	cfg      replication.BinlogSyncerConfig
	Tables   map[string][]config.Table
	gtid     string
	position string
}

func (b *Binlog) Run() {
	syncer := replication.NewBinlogSyncer(b.cfg)
	gtid, _ := mysql.ParseGTIDSet("mysql", b.gtid+":"+b.position)
	streamer, _ := syncer.StartSyncGTID(gtid)
	for {
		ev, _ := streamer.GetEvent(context.Background())
		// Dump event
		fmt.Printf(">>> %s\n", ev.Header.EventType)
		if e, ok := ev.Event.(*replication.RowsEvent); ok {
			fmt.Printf(">>> %s.", B2S(e.Table.Schema))
			fmt.Printf("%s\n", B2S(e.Table.Table))
			fmt.Printf(">>> %#v\n", e.Rows)
		}
	}
}

func B2S(bs []uint8) string {
	ba := make([]byte, 0, len(bs))
	for _, b := range bs {
		ba = append(ba, byte(b))
	}
	return string(ba)
}

func NewBinlog(serverId uint32, host string, port uint16, user, password string, t map[string][]config.Table, gtid, position string) *Binlog {
	return &Binlog{
		cfg: replication.BinlogSyncerConfig{
			ServerID: serverId,
			Flavor:   "mysql",
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
		},
		Tables:   t,
		gtid:     gtid,
		position: position,
	}
}
