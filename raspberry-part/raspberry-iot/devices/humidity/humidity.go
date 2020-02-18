package humidity

import (
	"fmt"
	"github.com/d2r2/go-dht"
	"hlf-iot/config"
	"hlf-iot/devices/led"
	"hlf-iot/helpers/queuewrapper"
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
	Led         *led.Led
}

func Init(ledBadData *led.Led) *Humidity {
	humidity := &Humidity{}
	humidity.Led = ledBadData

	return humidity
}

func (humidity *Humidity) GetDataInJsonString() (*queuewrapper.SendData, error) {
	var err error

	humidity.Temperature, humidity.Humidity, _, err =
		dht.ReadDHTxxWithRetry(dht.DHT11, GPIO, false, retry)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	humidity.Timestamp = int64(now.Unix())
	sendData := &queuewrapper.SendData{}
	sendData.Fcn = config.FCN_NAME_HUMIDITY
	sendData.Args = []string{fmt.Sprintf("%f", humidity.Humidity), fmt.Sprintf("%f", humidity.Temperature), fmt.Sprintf("%d", humidity.Timestamp)}
	if humidity.Humidity*humidity.Temperature == 0 {
		humidity.Led.SetOn()
	}

	return sendData, nil
}

func (humidity *Humidity) GetQueueElement() *queuewrapper.QueueStructure {
	return &queuewrapper.QueueStructure{GetDataFcn: humidity.GetDataInJsonString}
}
