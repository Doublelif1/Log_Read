package Tail_log

import (
	"awesomeProject1/Log_agent/Kakfa"
	"context"
	"fmt"
	"github.com/hpcloud/tail"
)

// 一个日志收集的的任务结构体
// 包括拉取配置数据信息path和topic ，一条这个任务读对应日志的实例tailobjinstance ，一个context随时取消这个读日志的实例
type TailTask struct {
	path            string             //日志配置数据信息path
	topic           string             //日志发送到kafka的topic信息
	tailobjinstance *tail.Tail         //一条这个任务读对应日志的实例tailobjinstance
	ctx             context.Context    //实现配置任务取消
	cancelFunc      context.CancelFunc //实现配置任务取消
}

// 创建一个TailTask类型任务
func NewTailTask(path, topic string) (tailObj *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{ //传入这个任务的信息
		path:       path,
		topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	tailObj.tailinit() //把tailObjinstance实例化,排除这个tail读日志的任务
	return tailObj
}

// 把NewTailTask 里面tailObj结构体里的tailObjinstance实例化
func (t *TailTask) tailinit() {
	config := tail.Config{
		ReOpen:    true,                                 // 重新打开 写满了重新打开继续写
		Follow:    true,                                 // 是否跟随 没读完然后名字变了 还能跟随然后继续读
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读 索引，标记
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,
	}
	var err error
	//实例化,准备读取文件日志
	t.tailobjinstance, err = tail.TailFile(t.path, config)
	if err != nil {
		fmt.Println("tail file failed, err:", err)
		return
	}
	//读取日志数据并发送到kafka中
	go t.pullandtokafka() //开个小弟去拉取日志内容并发送到kafka对应的topic里面
}
func (t *TailTask) pullandtokafka() {
	for {
		select {
		case line := <-t.tailobjinstance.Lines: //把结构体里的tailobjinstance拿出来读日志
			//发往kafka
			//先把日志数据line发送到kafka的数据通道中,在初始化的时候已经启动goruntine在通道出口处等待拿数据发往kafka了
			fmt.Println("发送消息")
			fmt.Println("topic:", t.topic)
			fmt.Println("1")
			fmt.Println("line:", line.Text)
			fmt.Println("2")
			Kafkalog.SentToChan(t.topic, line.Text)
		case <-t.ctx.Done(): // cancel 如果任务取消了，则结束掉这个小弟的任务
			return
		}
	}
}
