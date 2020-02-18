package light

import (
	"fmt"
	"github.com/davecheney/gpio"
	"hlf-iot/config"
	"hlf-iot/helpers/queuewrapper"
	"sync"
	"time"
)

type Light struct {
	pin          gpio.Pin
	Timestamp    int64                      `json:"timestamp"`
	CurrentTime  time.Time                  `json:"currenttime"`
	CurrentState bool                       `json:"currentstate"`
	QueueEntity  *queuewrapper.QueueWrapper `json:"queueentity"`
}

var instance *Light
var once sync.Once

func GetInstance(queueInput *queuewrapper.QueueWrapper) (*Light, error) {
	var err error
	once.Do(func() {
		instance = &Light{}
		instance.CurrentTime = time.Now()
		instance.CurrentState = false
		instance.QueueEntity = queueInput
		instance.pin, err = gpio.OpenPin(gpio.GPIO22, gpio.ModeInput)
		instance.pin.Clear()
	})

	return instance, err
}

func callBackBoth() {
	light, err := GetInstance(nil)
	if err != nil {
		panic(err.Error())
	}

	newTime := time.Now()
	if config.ItsTime(light.CurrentTime, newTime) {
		light.CurrentTime = newTime
		light.CurrentState = light.GetPinData()
		light.QueueEntity.AddToQueue(light.GetQueueElement())
	}
}

func (light *Light) SetCallBack() error {
	err := light.pin.BeginWatch(gpio.EdgeBoth, callBackBoth)

	return err
}

func (light *Light) GetPinData() bool {
	return light.pin.Get()
}

func (light *Light) GetDataInJsonString() (*queuewrapper.SendData, error) {
	now := time.Now()
	light.Timestamp = int64(now.Unix())
	sendData := &queuewrapper.SendData{}
	sendData.Fcn = config.FCN_NAME_LIGHT
	sendData.Args = []string{fmt.Sprintf("%d", config.B2i(light.CurrentState)), fmt.Sprintf("%d", light.Timestamp)}

	return sendData, nil
}

func (light *Light) GetQueueElement() *queuewrapper.QueueStructure {
	preparedData, err := light.GetDataInJsonString()
	if err != nil {
		panic(err.Error())
	}

	return &queuewrapper.QueueStructure{GetDataFcn: nil, PreparedData: preparedData}
}
