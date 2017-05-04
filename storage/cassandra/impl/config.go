package impl

import (
	"time"
)

// CassandraConn describes the properties to manage a Cassandra connection.
type CassandraConn struct {
	ContactPoints      []string      `yaml:"contactPoints"`
	Port               int           `yaml:"port"`
	Username           string        `yaml:"username"`
	Password           string        `yaml:"password"`
	Consistency        string        `yaml:"consistency"`
	ConnectionsPerHost int           `yaml:"connectionsPerHost"`
	Timeout            time.Duration `yaml:"timeout"`
	SocketKeepalive    time.Duration `yaml:"socketKeepalive"`
	ProtoVersion       int           `yaml:"protoVersion"`
	TTL                time.Duration `yaml:"ttl"`
	LocalDCOnly        bool          `yaml:"localDCOnly"` // deprecated
	DataCenter         string        `yaml:"dataCenter"`  // data center filter
	PageSize           int           `yaml:"pageSize"`
	RetryCount         int           `yaml:"retryCount"`
	HostPolicy         string        `yaml:"hostPolicy"`
	TimeoutLimit       int           `yaml:"timeoutLimit"`  // number of timeouts allowed
	CQLVersion         string        `yaml:"cqlVersion"`    // set only on C* 3.x
	MaxGoRoutines      int           `yaml:"maxGoroutines"` // a capacity limit
}