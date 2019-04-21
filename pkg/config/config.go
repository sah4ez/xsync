package config

import "time"

type Config struct {
	Threads   int                `yaml:"threads,omitempty"`
	BatchSize int                `yaml:"batchSize,omitempty"`
	Interval  time.Duration      `yaml:"interval,omitempty"`
	Schemas   map[string][]Table `yaml:"schemas,omitempty"`
	Source    MySQLCredentials   `yaml:"source"`
	Target    MySQLCredentials   `yaml:"target"`
	Binlog    BinlogSyncer       `yaml:"binlog,omitempty"`
	Kafka     Kafka              `yaml:"kafka,omitempty"`
}

type Kafka struct {
	Addr           string        `yaml:"addr"`
	Topic          string        `yaml:"topic"`
	Partition      int           `yaml:"partition"`
	MaxWait        time.Duration `yaml:"maxWait"`
	MinBytes       int           `yaml:"minBytes"`
	MaxBytes       int           `yaml:"maxBytes"`
	Offset         int64         `yaml:"offset"`
	CommitInterval time.Duration `yaml:"commitInterval"`
}

var NilTable = Table{}

type Table struct {
	Table      string        `yaml:"table"`
	Batch      string        `yaml:"batch,omitempty"`
	Interval   time.Duration `yaml:"interval,omitempty"`
	CleanAfter time.Duration `yaml:"cleanAfter,omitempty"`
	Latest     string        `yaml:"latest"`
	FieldID    string        `yaml:"fieldId"`
}

type MySQLCredentials struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
	SSL      bool   `yaml:"ssl"`
}

type BinlogSyncer struct {
	ServerID uint32 `yaml:"serverId"`
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	GTID     string `yaml:"gtid"`
	Position string `yaml:"position"`
}

type Cfg interface {
	Load() (*Config, error)
}
