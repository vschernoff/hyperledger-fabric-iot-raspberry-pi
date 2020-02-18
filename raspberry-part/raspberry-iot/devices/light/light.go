package light

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

type Light struct {
	pin       gpio.Pin
	Light     uint  `json:"light"`
	Timestamp int64 `json:"timestamp"`
}

func Init(gprsSensorInput *gprs.Gprs) *Light {
	gprsSensor = gprsSensorInput
	light := &Light{}
	var err error

	light.pin, err = gpio.OpenPin(gpio.GPIO22, gpio.ModeInput)
	if err != nil {
		fmt.Println(err)
	}
	light.pin.Clear()

	return light
}

func callBackBoth() {
	light := Init(gprsSensor)
	newState := light.GetPinData()
	newTime := time.Now()
	checkTime := config.ItsTime(currentTime, newTime)
	if currentState != newState && checkTime {
		currentState = newState
		currentTime = newTime
		gprsSensor.AddToQueue(light.GetQueueElement())
	}
}

func (light *Light) SetCallBack() {
	light.pin.BeginWatch(gpio.EdgeBoth, callBackBoth)
}

func (light *Light) GetPinData() bool {
	return light.pin.Get()
}

func (light *Light) GetDataInJsonString() *config.SendData {
	now := time.Now()
	light.Timestamp = int64(now.Unix())
	light.Light = config.B2i(currentState)
	sendData := &config.SendData{}
	sendData.Fcn = config.FCN_NAME_LIGHT
	sendData.Args = []string{fmt.Sprintf("%d", light.Light), fmt.Sprintf("%d", light.Timestamp)}

	return sendData
}

func (light *Light) GetQueueElement() *config.QueueStructure {
	queueStructure := &config.QueueStructure{}
	queueStructure.GetDataFcn = nil
	queueStructure.CheckResponseFcn = config.CheckInsertToBC
	queueStructure.JsonString = light.GetDataInJsonString()
	queueStructure.HttpAction = config.GPRS_HTTPACTION_POST

	return queueStructure
}
