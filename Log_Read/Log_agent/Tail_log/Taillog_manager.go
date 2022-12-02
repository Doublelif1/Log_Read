package Tail_log

import (
	"awesomeProject1/Log_agent/etcd"
	"fmt"
	"time"
)

var (
	TaskMgr *TailMgr
)

// 定义一个记录所有任务的结构体,相当于记账本 每条任务及信息都会在此记录
type TailMgr struct {
	logEntry    []*etcd.LogEntry      //记录每个任务的path和topic信息
	taskmap     map[string]*TailTask  //记录每个正在执行的任务
	newConfChan chan []*etcd.LogEntry //更新后的etcd数据(所有path和topic)置会被哨兵部分发送到这个通道中，方便更新任务
}

// 第一次启动时自动拉取现有etcd的数据(所有path和topic)，并启动他们
func Init(logEntryconf []*etcd.LogEntry) {
	//构造一个记录所有任务的结构体，并将里面的变量初始化
	TaskMgr = &TailMgr{
		logEntry:    logEntryconf,
		taskmap:     make(map[string]*TailTask, 20),
		newConfChan: make(chan []*etcd.LogEntry), //无缓冲区,一次发送和接受一次etcd更改的数据

	}
	//拉取现有的所有etcd数据信息(path,topic)，并逐一启动
	for _, logEntry := range logEntryconf {

		aTailtask := NewTailTask(logEntry.Path, logEntry.Topic) //逐一启动gorountine执行对应task

		key := fmt.Sprintf("%s_%s", logEntry.Path, logEntry.Topic) //输出etcd现有的数据任务
		fmt.Println("正在运行的任务:", key)

		TaskMgr.taskmap[key] = aTailtask //将首次启动的任务记录在taskmap里面
		//---------------------------------
	}

	go TaskMgr.listennewconf() //监听etcd数据的更新

}

// 监听配置通道newConfChan 有了新配置过来之后进行处理
func (t *TailMgr) listennewconf() {
	for {
		select {
		case newConf := <-t.newConfChan: //接收新配置
			fmt.Println("配置开始变更")
			//配置变更   两次新旧任务对比完成两个操作 ：1.对新任务的增加  2.对旧任务的删除

			//1. 配置增加  配置增加就创建多一个task
			for _, anewconf := range newConf { //循环新配置的任务与旧任务map中的map比对
				newconfkey := fmt.Sprintf("%s_%s", anewconf.Path, anewconf.Topic)

				_, ok := t.taskmap[newconfkey] //在旧任务map中查看新任务是否存在 利用布尔值ok判定

				if ok {
					//已经有的task 不需要操作 执行下一个循环
					continue
				} else { //若此任务为新任务 则做出以下操作

					ANewTailTask := NewTailTask(anewconf.Path, anewconf.Topic) //1创建这个新任务的goroutine

					newkey := fmt.Sprintf("%s_%s", anewconf.Path, anewconf.Topic)

					TaskMgr.taskmap[newkey] = ANewTailTask //2在taskmap中增加这个新任务

					anewlogentry := &etcd.LogEntry{
						Path:  anewconf.Path,
						Topic: anewconf.Topic,
					}

					t.logEntry = append(t.logEntry, anewlogentry) // 3在logEntry中增加这个新任务
					fmt.Println("增加任务:", newkey)
				}

			}

			//2. 配置删除:
			needtodelkey := []string{} //建立一个切片方便记录此次删除操作
			//两个for循环对比新旧配置  若旧任务中没有新配置的任务 则将需要删除的key记录到needtodelkey中
			for _, aoldconf := range t.logEntry {
				isDelete := true //建立一个是否需要删除的布尔值

				for _, anewconf := range newConf {
					//新旧任务比对↓
					if anewconf.Path == aoldconf.Path && anewconf.Topic == aoldconf.Topic {
						isDelete = false //若新旧任务相同 ， 则标志为不加入删除记录
						break
					}
				}

				//新旧任务比对完毕之后 将需要删除的旧任务key记录到needtodelkey中
				if isDelete {
					Delkey := fmt.Sprintf("%s_%s", aoldconf.Path, aoldconf.Topic)

					needtodelkey = append(needtodelkey, Delkey)

				}
			}
			//记录完毕
			//开始以下删除操作
			for _, Delkey := range needtodelkey {

				t.taskmap[Delkey].cancelFunc() //1取消这个旧任务的goroutine

				delete(t.taskmap, Delkey) //2在taskmap中删除这个旧任务

				//3在logEntry中删除这个旧任务↓

				for i, j := range t.logEntry {
					if j.Path+"_"+j.Topic == Delkey { //找到要删除元素的索引并切切掉
						fmt.Println("删除任务:", Delkey)
						t.logEntry = append(t.logEntry[:i], t.logEntry[(i+1):]...) //找到要删除元素的索引并切切掉
					}
				}
				//3在logEntry中删除这个旧任务↑
			}
			//删除完毕

			//查看更新后的配置↓
			fmt.Println("更新后的配置:")
			for _, logEntryvalue := range t.logEntry {
				fmt.Println("Path:", logEntryvalue.Path, "  | ", "Topic:", logEntryvalue.Topic)
			}
			//查看更新后的配置↑
		default:
			time.Sleep(time.Second)
		}

	}
}

// 向外暴露结构体里的newConfChan 方便监听配置传输的通道及取值
func ExposenewConfChan() chan<- []*etcd.LogEntry {
	return TaskMgr.newConfChan
}
