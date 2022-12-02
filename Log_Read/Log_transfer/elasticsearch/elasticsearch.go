package elasticsearch

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"time"
)

type LogData struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}

var (
	client     *elastic.Client
	esdatachan chan *LogData
)

func EsInit(address string) (err error) {
	//连接elasticsearch端口
	client, err = elastic.NewClient(elastic.SetURL("http://" + address))
	if err != nil {
		// Handle error
		fmt.Println("elasticsearch init faild")
		panic(err)
	}
	esdatachan = make(chan *LogData, 100000)
	fmt.Println("Connect to ElasticSearch Success")
	go listen_data_Sendtoes()
	return err
}

func Expose_sentto_esdatachan(topic string, value []byte) {
	data := &LogData{
		Topic: topic,
		Data:  string(value),
	}
	esdatachan <- data
}

func listen_data_Sendtoes() (err error) {
	for {
		select {
		case bodydata := <-esdatachan:
			fmt.Println("准备发送,bodydata：", bodydata.Data, bodydata.Topic)
			_, err = client.Index().
				Index(bodydata.Topic).
				BodyJson(bodydata).
				Do(context.Background())
			if err != nil {
				fmt.Println("sent to elasticsearch faild")
				panic(err)
				return err
			}
			fmt.Println("Sent to ElasticSearch Success!")
			fmt.Println("index :", bodydata.Topic, " data: ", bodydata.Data)
		default:
			time.Sleep(time.Second)
		}
	}
}

//
//type Person struct {
//	Name    string `json:"name"`
//	Age     int    `json:"age"`
//	Married bool   `json:"married"`
//	Sex     string `json:"sex"`
//}
//
//func main() {
//	//连接elasticsearch端口
//	client, err := elastic.NewClient(elastic.SetURL("http://127.0.0.1:9200"))
//	if err != nil {
//		// Handle error
//		panic(err)
//	}
//
//	fmt.Println("connect to elasticsearch success")
//	p1 := Person{
//		Name:    "hahah",
//		Age:     623,
//		Married: true,
//	}
//
//	put1, err := client.Index().
//		Index("testindex").
//		BodyJson(p1).
//		Do(context.Background())
//	if err != nil {
//		// Handle error
//		panic(err)
//	}
//
//	fmt.Printf("Indexed  %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
//}
