package main

import (
	"fmt"
	"hlf-iot/config"
	"hlf-iot/devices/barometer"
	"hlf-iot/devices/gprs"
	"hlf-iot/devices/gyroscope"
	"hlf-iot/devices/humidity"
	"hlf-iot/devices/light"
	"hlf-iot/devices/vibration"
	"hlf-iot/helpers/ca"
	"time"
)

func main() {
	// Init Sensors
	fabricCa := ca.GetInstance()
	gprsSensor := gprs.Init(1)

	i := 0
	for {
		fmt.Printf("**************** Getting certificate from fabric CA, retry№ %d ****************\n", i)
		i++

		fabricCa.GeneratePrivateKey()
		success := gprsSensor.SendRequest(fabricCa.TbsCsr())
		if !success {
			continue
		}

		fabricCa.SignTbsCsr()
		success = gprsSensor.SendRequest(fabricCa.EnrollCsr())
		if !success {
			continue
		} else {
			break
		}

		time.Sleep(2000 * time.Millisecond)
	}

	humiditySensor := humidity.Init()
	barometerSensor := barometer.Init()
	barometerSensor.Activate()
	gyroscopeSensor := gyroscope.Init()
	gyroscopeSensor.Activate()

	vibrationSensor := vibration.Init(gprsSensor)
	lightSensor := light.Init(gprsSensor)

	// Set Callbacks
	vibrationSensor.SetCallBack()
	lightSensor.SetCallBack()

	// Start GPRS daemon
	go gprsSensor.StartDaemon()

	// Gathering data in cycle
	i = 0
	for {
		fmt.Printf("**************** UPDATES %d ****************\n", i)

		gprsSensor.AddToQueue(gprsSensor.GetQueueElement())
		gprsSensor.AddToQueue(humiditySensor.GetQueueElement())
		gprsSensor.AddToQueue(barometerSensor.GetQueueElement())
		gprsSensor.AddToQueue(gyroscopeSensor.GetQueueElement())

		i++
		time.Sleep(config.DELAY_FOR_GATHERING_DATA_IN_CYCLE_SECONDS * time.Second)
	}

	fmt.Println("**************** END ****************")
}
