package main

import (
	"awesomeProject1/Log_transfer/Kafka_and_task_mgr"
	"awesomeProject1/Log_transfer/config"
	"awesomeProject1/Log_transfer/elasticsearch"
	"awesomeProject1/Log_transfer/etcd"
	"awesomeProject1/Log_transfer/ip"
	"fmt"
	"gopkg.in/ini.v1"
	"sync"
	"time"
)

var (
	cfg = new(config.Config)
)

func main() {

	//加载引用配置文件
	//将配置文件赋值给结构体的key
	err := ini.MapTo(cfg, "Log_transfer/config/config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}
	//初始化es
	elasticsearch.EsInit(cfg.ElasticSearch.Address)
	//初始化etcd
	err = etcd.Init(cfg.Etcd.Address, time.Duration(cfg.Etcd.Timeout)*time.Second)
	if err != nil {
		fmt.Println("Init etcd faild")
		return
	}
	//初始化kafka
	Kafka_and_task_mgr.KafkaInit([]string{cfg.Kafka.Address})
	//--------------------------------初始化完毕-------------------------
	//拟合得到要拉取etcd信息的key
	//拉取本机对外的	ip 不同的ip拟合到etcd.Getconf的key中对应不同的etcd用于拉取配置的key
	ip, err := ip.GetOutboundIP()
	if err != nil {
		fmt.Println("GetOutboundIP faild,err:", err)
	}
	etcdGetconfkey := fmt.Sprintf(cfg.Etcd.Key, ip)
	fmt.Println("the address of etcdgetconfkey: ", etcdGetconfkey)
	//从etcd中拉取日志收集项的配置信息
	Topicconfs, err := etcd.GetConf(etcdGetconfkey)
	if err != nil {
		fmt.Println("etcd GetConf faild")
		return
	}

	//初始化topic任务
	Kafka_and_task_mgr.TaskInit(Topicconfs)
	//派一个哨兵去watch etcd 中key的变化,有变化实时通知logtransfer，实现热加载
	var wg sync.WaitGroup
	wg.Add(1)

	go etcd.WatchConf(etcdGetconfkey, Kafka_and_task_mgr.ExposenewConfChan()) //哨兵发现更新了信息发送到通道里面

	wg.Wait()

}
