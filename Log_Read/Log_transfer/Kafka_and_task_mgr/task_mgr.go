package Kafka_and_task_mgr

import (
	"awesomeProject1/Log_transfer/etcd"
	"context"
	"fmt"
	"time"
)

var (
	TaskMgr *TailMgr
)

type TailTask struct {
	Topic      string             //topic信息
	Ctx        context.Context    //实现配置任务取消
	cancelFunc context.CancelFunc //实现配置任务取消
}

// 定义一个记录所有任务的结构体,相当于记账本 每条任务及信息都会在此记录
type TailMgr struct {
	topicconf   []*etcd.Topicconfs      //记录每个任务的path和topic信息
	taskmap     map[string]*TailTask    //记录每个正在执行的任务
	newConfChan chan []*etcd.Topicconfs //更新后的etcd数据(所有path和topic)置会被哨兵部分发送到这个通道中，方便更新任务
}

func TaskInit(Topics []*etcd.Topicconfs) {
	//构造一个记录所有任务的结构体，并将里面的变量初始化
	TaskMgr = &TailMgr{
		topicconf:   Topics,
		taskmap:     make(map[string]*TailTask, 20),
		newConfChan: make(chan []*etcd.Topicconfs), //无缓冲区,一次发送和接受一次etcd更改的数据
	}
	//拉取现有的所有etcd数据信息(path,topic)，并逐一启动
	for _, topicconf := range Topics {

		aTailtask := NewTailTask(topicconf.Topic) //逐一启动gorountine执行对应task

		TaskMgr.taskmap[topicconf.Topic] = aTailtask //将首次启动的任务记录在taskmap里面

		fmt.Println("正在运行的任务:", topicconf.Topic)
		//---------------------------------
	}

	go TaskMgr.listennewconf() //监听etcd数据的更新

}
func NewTailTask(topic string) (atask *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	atask = &TailTask{ //传入这个任务的信息
		Topic:      topic,
		Ctx:        ctx,
		cancelFunc: cancel,
	}
	go Consumekafkadatatoes(atask.Topic, atask)
	return atask
}

// 监听配置通道newConfChan 有了新配置过来之后进行处理
func (t *TailMgr) listennewconf() {
	for {
		select {
		case newConf := <-t.newConfChan: //接收新配置
			fmt.Println("配置开始变更")
			//配置变更   两次新旧任务对比完成两个操作 ：1.对旧任务的删除  2.对新任务的增加

			//1. 配置删除:
			needtodelkey := []string{} //建立一个切片方便记录此次删除操作
			//两个for循环对比新旧配置  若旧任务中没有新配置的任务 则将需要删除的key记录到needtodelkey中
			for _, aoldconf := range t.topicconf {
				isDelete := true //建立一个是否需要删除的布尔值

				for _, anewconf := range newConf {
					//新旧任务比对↓
					if anewconf.Topic == aoldconf.Topic {
						isDelete = false //若新旧任务相同 ， 则标志为不加入删除记录
						break
					}
				}

				//新旧任务比对完毕之后 将需要删除的旧任务key记录到needtodelkey中
				if isDelete {
					needtodelkey = append(needtodelkey, aoldconf.Topic)

				}
			}
			//记录完毕
			//开始以下删除操作
			for _, Deltopic := range needtodelkey {

				t.taskmap[Deltopic].cancelFunc() //1取消这个旧任务的goroutine

				delete(t.taskmap, Deltopic) //2在taskmap中删除这个旧任务

				//3在topicconf中删除这个旧任务↓

				for i, j := range t.topicconf {

					if j.Topic == Deltopic { //找到要删除元素的索引并切切掉
						fmt.Println("删除任务:", Deltopic)
						t.topicconf = append(t.topicconf[:i], t.topicconf[(i+1):]...) //找到要删除元素的索引并切切掉
					}
				}
				//3在topicconf中删除这个旧任务↑
			}
			//删除完毕

			//2. 配置增加  配置增加就创建多一个task
			for _, anewconf := range newConf { //循环新配置的任务与旧任务map中的map比对

				_, ok := t.taskmap[anewconf.Topic] //在旧任务map中查看新任务是否存在 利用布尔值ok判定

				if ok {
					//已经有的task 不需要操作 执行下一个循环
					continue
				} else { //若此任务为新任务 则做出以下操作

					ANewTailTask := NewTailTask(anewconf.Topic) //1创建这个新任务的goroutine

					TaskMgr.taskmap[anewconf.Topic] = ANewTailTask //2在taskmap中增加这个新任务

					anewlogentry := &etcd.Topicconfs{
						Topic: anewconf.Topic,
					}

					t.topicconf = append(t.topicconf, anewlogentry) // 3在topicconf中增加这个新任务
					fmt.Println("增加任务:", anewconf.Topic)
				}

			}

			//查看更新后的配置↓
			fmt.Println("更新后的配置:")
			for _, topicconfvalue := range t.topicconf {
				fmt.Println("topicconf:", topicconfvalue.Topic)
			}
			//查看更新后的配置↑
		default:
			time.Sleep(time.Second)
		}

	}
}

// 向外暴露结构体里的newConfChan 方便监听配置传输的通道及取值
func ExposenewConfChan() chan<- []*etcd.Topicconfs {
	return TaskMgr.newConfChan
}
