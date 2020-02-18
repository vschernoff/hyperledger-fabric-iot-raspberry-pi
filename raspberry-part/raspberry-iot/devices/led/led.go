package led

import (
	"github.com/davecheney/gpio"
)

type Led struct {
	Pin gpio.Pin
}

func Init(pin int) (*Led, error) {
	led := &Led{}
	var err error

	led.Pin, err = gpio.OpenPin(pin, gpio.ModeOutput)
	if err != nil {
		return nil, err
	}
	led.Pin.Clear()

	return led, nil
}

func (led *Led) SetOn() {
	led.Pin.Set()
}

func (led *Led) SetOff() {
	led.Pin.Clear()
}
