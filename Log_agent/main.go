package main

import (
	"awesomeProject1/Log_agent/Kakfa"
	"awesomeProject1/Log_agent/Tail_log"
	"awesomeProject1/Log_agent/config"
	"awesomeProject1/Log_agent/etcd"
	"awesomeProject1/Log_agent/ip"
	"fmt"
	"gopkg.in/ini.v1"
	"strings"
	"sync"
	"time"
)

var (
	cfg = new(config.Config)
)

// log_agent的入口
// 先把所有依赖的组件加载完之后 再做业务逻辑
func main() {
	//加载引用配置文件
	//将配置文件赋值给结构体的key
	err := ini.MapTo(cfg, "Log_agent/config/config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}
	//初始化kafka连接
	err = Kafkalog.Init(strings.Split(cfg.Kafka.Address, ";"), cfg.Kafka.MaxChan)
	if err != nil {
		fmt.Println("Init kafka faild")
		return
	}
	//初始化etcd
	err = etcd.Init(cfg.Etcd.Address, time.Duration(cfg.Etcd.Timeout)*time.Second)
	if err != nil {
		fmt.Println("Init etcd faild")
		return
	}

	//拉取本机对外的	ip 不同的ip拟合到etcd.Getconf的key中对应不同的etcd用于拉取配置的key
	ip, err := ip.GetOutboundIP()
	if err != nil {
		fmt.Println("GetOutboundIP faild,err:", err)
	}

	etcdGetconfkey := fmt.Sprintf(cfg.Etcd.Key, ip)
	fmt.Println("the address of etcdgetconfkey: ", etcdGetconfkey)
	//------------
	//从etcd中拉取日志收集项的配置信息
	LogEntryConfs, err := etcd.GetConf(etcdGetconfkey)
	if err != nil {
		fmt.Println("etcd GetConf faild")
		return
	}

	//3循环LogEntryConfs并创造一个goruntine执行发送到kafka的task
	Tail_log.Init(LogEntryConfs)

	//派一个哨兵去watch etcd 中key的变化,有变化实时通知logagent，实现热加载
	var wg sync.WaitGroup
	wg.Add(1)
	go etcd.WatchConf(etcdGetconfkey, Tail_log.ExposenewConfChan()) //哨兵发现更新了信息发送到通道里面
	wg.Wait()
}
