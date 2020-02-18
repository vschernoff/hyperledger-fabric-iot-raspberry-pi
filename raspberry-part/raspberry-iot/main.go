package main

import (
	"fmt"
	"hlf-iot/config"
	"hlf-iot/devices/barometer"
	"hlf-iot/devices/gps"
	"hlf-iot/devices/gyroscope"
	"hlf-iot/devices/humidity"
	"hlf-iot/devices/led"
	"hlf-iot/devices/light"
	"hlf-iot/devices/vibration"
	"hlf-iot/helpers/ca"
	"hlf-iot/helpers/queuewrapper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func execute() {
	// Init Sensors Activity Grid
	sensorsActivityGrid := config.NewSensorsActivityGrid(true)

	// Init Sensors
	ledBadSensorData, err := led.Init(config.LED_PIN_BAD_SENSOR_DATA)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.LedBad = false
		ledBadSensorData = nil
	}

	ledSuccessEnroll, err := led.Init(config.LED_PIN_SUCCESS_ENROLL)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.LedSuccessEnroll = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	ledBadGpsData, err := led.Init(config.LED_PIN_BAD_GPS_DATA)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.LedBadGps = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	// Turn off LEDs before shutting down app
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTSTP)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Aborting... \n", sig)
			if sensorsActivityGrid.LedBad {
				ledBadSensorData.SetOff()
			}
			if sensorsActivityGrid.LedSuccessEnroll {
				ledSuccessEnroll.SetOff()
			}
			if sensorsActivityGrid.LedBadGps {
				ledBadGpsData.SetOff()
			}
			os.Exit(1)
		}
	}()

	fabricCa := ca.GetInstance()

	gpsSensor, err := gps.Init(ledBadGpsData)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.GpsSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	queue := queuewrapper.Init()

	fmt.Printf("**************** Getting certificate from usb key storage ****************\n")
	success, err := fabricCa.GetCertificateFromKeyStorage()
	if err != nil {
		fmt.Println("Error: ", err.Error())
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}
	if success && sensorsActivityGrid.LedSuccessEnroll {
		ledSuccessEnroll.SetOn()
	}

	humiditySensor, err := humidity.Init(ledBadSensorData)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.HumiditySensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	barometerSensor, err := barometer.Init(ledBadSensorData)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.BarometerSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	err = barometerSensor.Activate()
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.BarometerSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	gyroscopeSensor, err := gyroscope.Init(ledBadSensorData)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.GyroscopeSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	err = gyroscopeSensor.Activate()
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.GyroscopeSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	vibrationSensor, err := vibration.GetInstance(queue)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.VibrationSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	lightSensor, err := light.GetInstance(queue)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		sensorsActivityGrid.LightSensor = false
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOn()
		}
	}

	// Set Callbacks
	if sensorsActivityGrid.VibrationSensor {
		err = vibrationSensor.SetCallBack()
		if err != nil {
			fmt.Println("Error: ", err.Error())
			sensorsActivityGrid.VibrationSensor = false
			if sensorsActivityGrid.LedBad {
				ledBadSensorData.SetOn()
			}
		}
	}

	if sensorsActivityGrid.LightSensor {
		err = lightSensor.SetCallBack()
		if err != nil {
			fmt.Println("Error: ", err.Error())
			sensorsActivityGrid.LightSensor = false
			if sensorsActivityGrid.LedBad {
				ledBadSensorData.SetOn()
			}
		}
	}

	// Start queue daemon
	go queue.StartDaemon()

	// Gathering data in cycle
	i := 0
	for {
		fmt.Printf("**************** UPDATES %d ****************\n", i)

		if sensorsActivityGrid.GpsSensor {
			queue.AddToQueue(gpsSensor.GetQueueElement())
		}
		if sensorsActivityGrid.HumiditySensor {
			queue.AddToQueue(humiditySensor.GetQueueElement())
		}
		if sensorsActivityGrid.BarometerSensor {
			queue.AddToQueue(barometerSensor.GetQueueElement())
		}
		if sensorsActivityGrid.GyroscopeSensor {
			queue.AddToQueue(gyroscopeSensor.GetQueueElement())
		}

		i++
		time.Sleep(config.DELAY_FOR_GATHERING_DATA_IN_CYCLE_SECONDS * time.Second)
		if sensorsActivityGrid.LedBad {
			ledBadSensorData.SetOff()
		}
		if sensorsActivityGrid.LedBadGps {
			ledBadGpsData.SetOff()
		}
	}

	fmt.Println("**************** END ****************")
}

func main() {
	defer func() {
		fmt.Println("Main defer")
		if r := recover(); r != nil {
			fmt.Println("Recovered in execute function: ", r)
			time.Sleep(10000 * time.Millisecond)
			execute()
		}
	}()

	execute()
}
