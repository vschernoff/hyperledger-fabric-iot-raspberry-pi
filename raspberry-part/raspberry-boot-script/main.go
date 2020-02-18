package main

import (
	"bytes"
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	GPIO_BUTTON_PIN = 23
	GPIO_LED_PIN    = 24
	APP_PATH_FORMAT = "%s/src/hlf-iot-bc-full/raspberry-part/raspberry-iot/build/main"
	GOPATH          = "/home/pi/Examples"
)

func rpioInit() {
	err := rpio.Open()
	if err != nil {
		log.Println(err)
		time.Sleep(5000 * time.Millisecond)
		rpioInit()
	}
}

func main() {
	var cmd *exec.Cmd
	var out bytes.Buffer
	var outStr string

	rpioInit()

	button := rpio.Pin(GPIO_BUTTON_PIN)
	button.Input()
	button.PullUp()

	led := rpio.Pin(GPIO_LED_PIN)
	led.Output()
	led.Low()

	time.Sleep(100 * time.Millisecond)

	ch := make(chan string)

	// print output from executing program
	go func() {
		freshOutput := ""
		for {
			logData := <-ch

			if len(logData) > 0 {
				if len(freshOutput) <= len(logData) {
					freshOutput = logData[len(freshOutput):]
					fmt.Println(freshOutput)
				}
			}
		}
	}()

	// button pushing handling
	go func(cmd *exec.Cmd, out bytes.Buffer, outStr string, button, led rpio.Pin) {
		buttonOn := false
		buttonState := rpio.Low
		for {
			buttonState = button.Read()
			if buttonState == rpio.Low {
				if buttonOn {
					log.Println("App was terminated")
					led.Low()

					if err := cmd.Process.Signal(syscall.SIGTSTP); err != nil {
						log.Println("failed to kill process: ", err)
					}

					cmd = nil
					buttonOn = false

					time.Sleep(100 * time.Millisecond)
				} else {
					goPath := os.ExpandEnv("$GOPATH")
					if len(goPath) == 0 {
						goPath = GOPATH
					}

					cmd = exec.Command(fmt.Sprintf(APP_PATH_FORMAT, goPath))
					cmd.Stdout = &out

					log.Println("App was started")
					led.High()

					if err := cmd.Start(); err != nil {
						log.Println(err)
					}

					buttonOn = true
					time.Sleep(100 * time.Millisecond)
				}
			}
			if cmd != nil {
				outStr = out.String()
				out.Reset()

				ch <- outStr
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(cmd, out, outStr, button, led)

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
