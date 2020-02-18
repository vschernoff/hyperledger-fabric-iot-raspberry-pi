package vibration

import (
	"fmt"
	"github.com/davecheney/gpio"
	"hlf-iot/config"
	"hlf-iot/helpers/queuewrapper"
	"sync"
	"time"
)

type Vibration struct {
	Pin          gpio.Pin
	Timestamp    int64                      `json:"timestamp"`
	CurrentTime  time.Time                  `json:"currenttime"`
	CurrentState bool                       `json:"currentstate"`
	QueueEntity  *queuewrapper.QueueWrapper `json:"queueentity"`
}

var instance *Vibration
var once sync.Once

func GetInstance(queueInput *queuewrapper.QueueWrapper) (*Vibration, error) {
	var err error
	once.Do(func() {
		instance = &Vibration{}
		instance.CurrentTime = time.Now()
		instance.CurrentState = false
		instance.QueueEntity = queueInput
		instance.Pin, err = gpio.OpenPin(gpio.GPIO4, gpio.ModeInput)
		instance.Pin.Clear()
	})

	return instance, err
}

func callBackBoth() {
	vibration, err := GetInstance(nil)
	if err != nil {
		panic(err.Error())
	}

	newTime := time.Now()
	if config.ItsTime(vibration.CurrentTime, newTime) {
		vibration.CurrentTime = newTime
		vibration.CurrentState = vibration.GetPinData()
		vibration.QueueEntity.AddToQueue(vibration.GetQueueElement())
	}
}

func (vibration *Vibration) SetCallBack() error {
	err := vibration.Pin.BeginWatch(gpio.EdgeBoth, callBackBoth)

	return err
}

func (vibration *Vibration) GetPinData() bool {
	return vibration.Pin.Get()
}

func (vibration *Vibration) GetDataInJsonString() (*queuewrapper.SendData, error) {
	now := time.Now()
	vibration.Timestamp = int64(now.Unix())
	sendData := &queuewrapper.SendData{}
	sendData.Fcn = config.FCN_NAME_VIBRATION
	sendData.Args = []string{fmt.Sprintf("%d", config.B2i(vibration.CurrentState)), fmt.Sprintf("%d", vibration.Timestamp)}

	return sendData, nil
}

func (vibration *Vibration) GetQueueElement() *queuewrapper.QueueStructure {
	preparedData, err := vibration.GetDataInJsonString()
	if err != nil {
		panic(err.Error())
	}

	return &queuewrapper.QueueStructure{GetDataFcn: nil, PreparedData: preparedData}
}
