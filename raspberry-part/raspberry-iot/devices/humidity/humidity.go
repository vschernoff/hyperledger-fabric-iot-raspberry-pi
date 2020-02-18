package humidity

import (
	"fmt"
	"github.com/d2r2/go-dht"
	"hlf-iot/config"
	"time"
)

const (
	GPIO  = 17
	retry = 10
)

type Humidity struct {
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	Timestamp   int64   `json:"timestamp"`
}

func Init() *Humidity {
	humidity := &Humidity{}

	return humidity
}

func (humidity *Humidity) GetDataInJsonString() *config.SendData {
	var err error
	humidity.Temperature, humidity.Humidity, _, err =
		dht.ReadDHTxxWithRetry(dht.DHT11, GPIO, false, retry)
	if err != nil {
		fmt.Println(err)
	}
	now := time.Now()
	humidity.Timestamp = int64(now.Unix())
	sendData := &config.SendData{}
	sendData.Fcn = config.FCN_NAME_HUMIDITY
	sendData.Args = []string{fmt.Sprintf("%f", humidity.Humidity), fmt.Sprintf("%f", humidity.Temperature), fmt.Sprintf("%d", humidity.Timestamp)}

	return sendData
}

func (humidity *Humidity) GetQueueElement() *config.QueueStructure {
	queueStructure := &config.QueueStructure{}
	queueStructure.GetDataFcn = nil
	queueStructure.CheckResponseFcn = config.CheckInsertToBC
	queueStructure.JsonString = humidity.GetDataInJsonString()
	queueStructure.HttpAction = config.GPRS_HTTPACTION_POST

	return queueStructure
}
