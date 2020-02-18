package vibration

import (
	"fmt"
	"github.com/davecheney/gpio"
	"hlf-iot/config"
	"hlf-iot/devices/gprs"
	"time"
)

var currentState = false
var currentTime = time.Now()
var gprsSensor *gprs.Gprs

type Vibration struct {
	Pin       gpio.Pin
	Vibration uint  `json:"vibration"`
	Timestamp int64 `json:"timestamp"`
}

func Init(gprsSensorInput *gprs.Gprs) *Vibration {
	gprsSensor = gprsSensorInput
	vibration := &Vibration{}
	var err error

	vibration.Pin, err = gpio.OpenPin(gpio.GPIO4, gpio.ModeInput)
	if err != nil {
		fmt.Println(err)
	}
	vibration.Pin.Clear()

	return vibration
}

func callBackBoth() {
	vibration := Init(gprsSensor)
	newState := vibration.GetPinData()
	newTime := time.Now()
	checkTime := config.ItsTime(currentTime, newTime)
	if currentState != newState && checkTime {
		currentState = newState
		currentTime = newTime
		gprsSensor.AddToQueue(vibration.GetQueueElement())
	}
}

func (vibration *Vibration) SetCallBack() {
	vibration.Pin.BeginWatch(gpio.EdgeBoth, callBackBoth)
}

func (vibration *Vibration) GetPinData() bool {
	return vibration.Pin.Get()
}

func (vibration *Vibration) GetDataInJsonString() *config.SendData {
	now := time.Now()
	vibration.Timestamp = int64(now.Unix())
	vibration.Vibration = config.B2i(currentState)
	sendData := &config.SendData{}
	sendData.Fcn = config.FCN_NAME_VIBRATION
	sendData.Args = []string{fmt.Sprintf("%d", vibration.Vibration), fmt.Sprintf("%d", vibration.Timestamp)}

	return sendData
}

func (vibration *Vibration) GetQueueElement() *config.QueueStructure {
	queueStructure := &config.QueueStructure{}
	queueStructure.GetDataFcn = nil
	queueStructure.CheckResponseFcn = config.CheckInsertToBC
	queueStructure.JsonString = vibration.GetDataInJsonString()
	queueStructure.HttpAction = config.GPRS_HTTPACTION_POST

	return queueStructure
}
