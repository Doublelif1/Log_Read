package config

type Config struct {
	Kafka Kafka `ini:"kafka"`
	Etcd  Etcd  `ini:"etcd"`
}

type Kafka struct {
	Address string `ini:"address"`
	MaxChan int    `ini:"chan_max_size"`
}

type Etcd struct {
	Address string `ini:"address"`
	Key     string `ini:"collect_log_key"`
	Timeout int    `ini:"timeout"`
}
