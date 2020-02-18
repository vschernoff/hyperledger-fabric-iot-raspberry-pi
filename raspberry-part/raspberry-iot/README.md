# raspberry-iot

1) Execute command
```
make build
```
2) Add to ~/.bashrc $GOPATH and $GOROOT

For example:
```
export GOPATH=$HOME/Examples
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

### Useful links:

- configuring i2c (https://learn.adafruit.com/adafruits-raspberry-pi-lesson-4-gpio-setup/configuring-i2c)
- mounting usb flash (https://www.raspberrypi-spy.co.uk/2014/05/how-to-mount-a-usb-flash-disk-on-the-raspberry-pi/)

### Sensors and modules:

- Humidity dht11
- Barometric pressure/temperature/altitude BMP180
- Light sensor module (4-wire, with both digital and analog output)
- Gyroscope + Accelerometer + Temperature MPU6050
- Shock sensor 801S
- SIM808 GPS GSM GPRS Module