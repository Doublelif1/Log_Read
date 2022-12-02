package config

type Config struct {
	Kafka         Kafka         `ini:"kafka"`
	ElasticSearch ElasticSearch `ini:"elasticsearch"`
	Etcd          Etcd          `ini:"etcd"`
}

type Kafka struct {
	Address string `ini:"address"`
}

type ElasticSearch struct {
	Address string `ini:"address"`
}
type Etcd struct {
	Address string `ini:"address"`
	Key     string `ini:"collect_log_key"`
	Timeout int    `ini:"timeout"`
}
