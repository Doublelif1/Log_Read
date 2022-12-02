package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var cli *clientv3.Client //初始化一个连通etcd的全局变量

type LogEntry struct { //定义一个结构来获取etcd上对应key中需要操作的path和topic数据
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

// 利用clientv3连接etcd
func Init(address string, timeout time.Duration) (err error) {
	//利用clientv3连接etcd

	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: timeout,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return err
	}
	fmt.Println("etcd 连接成功！")
	return nil
}

// 从etcd中根据key获取值(日志配置项)
func GetConf(key string) (logentryconfs []*LogEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, key)
	cancel() //这三条语句为绑定操作 根据etcd上的某个key去get对应的值
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return nil, err
	}

	for _, ev := range resp.Kvs { //将获取的值利用Unmarshal赋给logentryconfs结构体切片中
		err = json.Unmarshal(ev.Value, &logentryconfs)
		if err != nil {
			fmt.Println("Unmarshal Conf faild ,error : ", err)
			return
		}
	}

	return logentryconfs, nil
}

// 哨兵:随时监视并上传etcd对应key值的所变更的信息
func WatchConf(key string, newConfChan chan<- []*LogEntry) {
	fmt.Println("哨兵出击，随时等待etcd新配置的更新...")
	rch := cli.Watch(context.Background(), key) // <-chan WatchResponse  watch(监视)对应的key
	// 从通道尝试取值（监视的信息）↓
	for wresp := range rch {
		for _, ev := range wresp.Events {
			//1.先判断对key所作的操作的类型
			fmt.Println("------------此次etcd更新的操作为:----------")
			fmt.Printf("Type: %s\n", ev.Type)     //输出对key操作类型
			fmt.Printf("Key:%s\n", ev.Kv.Key)     //输出操作的key
			fmt.Printf("Value:%s\n", ev.Kv.Value) //输出key的对应值

			var newConf []*LogEntry //定义一个LogEntry的切片类型结构体去随时获取etcd中变更的path和topic

			if int32(ev.Type) == 1 { //int32(ev.Type)==1时 说明此时的ev.Type为"DELETE"，也就是etcd上对应key的值被删除
				fmt.Println("etcd对应数据被清空")
				newConfChan <- newConf //发送空的数据到数据通道中，提示值被删除
				continue               //如果是删除的话，则continue到下个循环 ，不拉取此次key的数据
			}

			err := json.Unmarshal(ev.Kv.Value, &newConf) //如果不是删除的话,则拉取此次key的path和topic数据到newConf
			if err != nil {
				fmt.Printf("Unmarshal faild err :", err)
				continue
			}
			fmt.Println("发送新配置:")
			newConfChan <- newConf //发送获取的新配置到数据通道中，等待被获取
		}
	}
}
