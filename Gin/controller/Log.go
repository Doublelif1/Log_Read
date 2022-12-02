package controller

import (
	"awesomeProject1/Log_agent/config"
	"awesomeProject1/Log_agent/etcd"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.etcd.io/etcd/clientv3"
	"gopkg.in/ini.v1"
	"net/http"
	"time"
)

var cli *clientv3.Client
var err error

func Log_etcdctlPage(c *gin.Context) {
	c.HTML(http.StatusOK, "etcd.html", gin.H{})
}
func Log_etcdctl(c *gin.Context) {
	etcdkey := c.PostForm("etcdkey")
	etcdvalue := c.PostForm("etcdvalue")
	chose := c.PostForm("chose")

	cli, err = initetcd()
	if err != nil {
		c.String(200, "init etcd faild,err:", err)
		return
	}
	defer cli.Close()

	if string(chose) == "get" {
		get(c, cli, etcdkey)
	}
	if string(chose) == "del" {
		del(c, cli, etcdkey)
	}
	if string(chose) == "change" {
		change(c, cli, etcdkey, etcdvalue)
	}
}

func get(c *gin.Context, cli *clientv3.Client, etcdkey string) {
	var logentryconfs *[]etcd.LogEntry
	fmt.Println("开始查询")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//ctx, cancel = context.WithCancel(context.Background())
	resp, err := cli.Get(ctx, etcdkey)
	cancel()
	if err != nil {
		fmt.Printf("查询数据失败, error:%v\n", err)
		c.String(200, "查询数据失败, error:", err)
		return
	}
	if resp.Count == 0 {
		c.String(200, "数据为空")
		return
	}
	//for _, ev := range resp.Kvs {
	//	fmt.Printf("%s:%s\n", ev.Key, ev.Value)
	//}
	for _, ev := range resp.Kvs {
		err = json.Unmarshal(ev.Value, &logentryconfs)
		if err != nil {
			fmt.Println("Unmarshal Conf faild ,error : ", err)
			c.String(200, "Unmarshal Conf faild ,error : ", err)
			return
		}
		for index, value := range *logentryconfs {
			fmt.Println("index: ", index)
			fmt.Println("value.Topic:", value.Topic, "  value.Path:", value.Path)
			c.String(200, "index: ", index)
			c.String(200, "value.Topic:", value.Topic, "  value.Path:", value.Path)
		}

	}
}
func del(c *gin.Context, cli *clientv3.Client, etcdkey string) {
	fmt.Println("开始删除")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//ctx, cancel = context.WithCancel(context.Background())
	_, err := cli.Delete(ctx, etcdkey)
	cancel()
	if err != nil {
		fmt.Printf("删除数据失败, err:%v\n", err)
		return
	}
	fmt.Println("删除成功")
	c.String(200, "删除成功")

	//for _, ev := range resp.Kvs {
	//	fmt.Printf("%s:%s\n", ev.Key, ev.Value)
	//}
}
func change(c *gin.Context, cli *clientv3.Client, etcdkey string, etcdvalue string) {
	fmt.Println("开始修改数据")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//value := `[{"path":"D:/testlog/redis.log","topic":"testlog_agent_log"},{"path":"D:/testlog/mysql.log","topic":"test_topic"}]`
	_, err := cli.Put(ctx, etcdkey, etcdvalue)
	cancel()
	if err != nil {
		fmt.Printf("修改数据失败, error:%v\n", err)
		return
	}
	fmt.Println("修改数据成功")
	//修改后的数据
	c.String(200, "修改后的数据为:")

	var logentryconfs *[]etcd.LogEntry
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	//ctx, cancel = context.WithCancel(context.Background())
	resp, err := cli.Get(ctx, etcdkey)
	cancel()
	if err != nil {
		fmt.Printf("查询数据失败, error:%v\n", err)
		c.String(200, "查询数据失败, error:", err)
		return
	}
	for _, ev := range resp.Kvs {

		err = json.Unmarshal(ev.Value, &logentryconfs)
		if err != nil {
			fmt.Println("Unmarshal Conf faild ,error : ", err)
			c.String(200, "Unmarshal Conf faild ,error : ", err)
			return
		}
		for index, value := range *logentryconfs {
			fmt.Println("index: ", index)
			fmt.Println("value.Topic:", value.Topic, "  value.Path:", value.Path)
			c.String(200, "index: ", index)
			c.String(200, "value.Topic:", value.Topic, "  value.Path:", value.Path)
		}

	}

}

func initetcd() (cli *clientv3.Client, err error) {
	var (
		cfg = new(config.Config)
	)

	err = ini.MapTo(cfg, "Log_agent/config/config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{cfg.Etcd.Address},
		DialTimeout: time.Duration(cfg.Etcd.Timeout) * time.Second,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
	}
	fmt.Println("etcd 连接成功！")
	return cli, err
}
