package Kafka_and_task_mgr

import (
	"awesomeProject1/Log_transfer/elasticsearch"
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

// Kakfa consumer
var consumer sarama.Consumer

func KafkaInit(address []string) (err error) {
	//连接服务器broke
	consumer, err = sarama.NewConsumer(address, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		return err
	}
	return
}

func Consumekafkadatatoes(topic string, atask *TailTask) (err error) {
	//连接topic的分区  根据topic找到它所有的分区
	partitionList, err := consumer.Partitions(topic) // 根据topic取到所有的分区
	if err != nil {
		fmt.Printf("fail to putorget_etcd list of partition:err%v\n", err)
		return err
	}
	fmt.Println("New Topic :", topic, "分区: ", partitionList)
	// 遍历所有的分区
	for partition := range partitionList {
		// 针对每个分区创建一个对应的分区消费者
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("failed to start consumer for partition %d,err:%v\n", partition, err)
			return err
		}
		defer pc.AsyncClose()
		// 异步从每个分区消费信息
		//gorountine 获取每个分区的信息
		go run(pc, atask)
	}
	for {
		select {
		case <-atask.Ctx.Done():
			return
		default:
			time.Sleep(time.Second)
		}
	}
}
func run(pc sarama.PartitionConsumer, t *TailTask) {
	for {
		select {
		case <-t.Ctx.Done():
			return
		case msg := <-pc.Messages():

			elasticsearch.Expose_sentto_esdatachan(msg.Topic, msg.Value)
		default:
			time.Sleep(time.Second)
		}
	}
}
