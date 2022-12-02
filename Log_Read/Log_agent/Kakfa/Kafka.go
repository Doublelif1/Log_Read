package Kafkalog

import (
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

//往kafka里写入日志文件
type logData struct {
	topic string
	data  string
}

var (
	client      sarama.SyncProducer //声明一个全局的连接kafka的生产者 client
	logDataChan chan *logData       //声明一个数据通道记录需要往kafka写的topic和data
)

//kafka初始化
func Init(addrs []string, Maxchan int) (err error) {
	//配置kafka参数
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	client, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return err
	}
	fmt.Println("Kakfa 连接成功！")

	//初始化logDataChan 数据通道 用于接收需要发送到kafka的topic和data
	logDataChan = make(chan *logData, Maxchan)

	//然后派一个小弟在数据通道出口处等待 在出口处去不断的取数据
	go sentToKafka()
	return nil
}

//向外暴露此函数 ，可以利用此函数将需要发送到kafka的topic和data放进数据通道logDataChan中
//logDataChan <-
func SentToChan(topic, data string) {
	msage := &logData{
		topic: topic,
		data:  data,
	}

	logDataChan <- msage
}

//从数据通道里取topic和data发往kafka
//kafka <- logDataChan
func sentToKafka() {

	for {
		select {
		case newlogdata := <-logDataChan:

			//构造一个消息msg 取出数据通道中的数据topic data赋值给此消息
			msg := &sarama.ProducerMessage{}
			msg.Topic = newlogdata.topic
			msg.Value = sarama.StringEncoder(newlogdata.data)

			//根据消息的topic将消息的data发送到kafka
			partition, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Println("日志发送到kafka成功咯 送达的Topic为:", msg.Topic)
			fmt.Printf("partition:%v offset:%v\n", partition, offset)
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}
}
