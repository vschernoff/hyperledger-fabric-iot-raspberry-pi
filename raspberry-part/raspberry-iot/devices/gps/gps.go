package gps

import (
	"bytes"
	"fmt"
	"github.com/tarm/serial"
	"hlf-iot/config"
	"hlf-iot/devices/led"
	"hlf-iot/helpers/queuewrapper"
	"hlf-iot/helpers/serialwrapper"
	"regexp"
	"strconv"
	"time"
)

// Response
// +CGNSINF:  <GNSS run status>,<Fix status>,<UTC date & Time>,<Latitude>,<Longitude>,<MSL Altitude>
var geoRegex = regexp.MustCompile(`(?s)\+CGNSINF: (\d+),(\d+),(\d+.\d+),(\d+.\d+),(\d+.\d+),(\d+.\d+)`)

const (
	LATITUDE = iota + 4
	LONGITUDE
	ALTITUDE
)

type Gps struct {
	Serial    *serial.Port
	BufferStr bytes.Buffer
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
	Altitude  float32 `json:"altitude"`
	Timestamp int64   `json:"timestamp"`
	Led       *led.Led
}

func Init(ledBadData *led.Led) (*Gps, error) {
	gps := &Gps{}
	var err error

	gps.Serial, err = serialwrapper.Init()
	if err != nil {
		return nil, err
	}

	gps.Led = ledBadData

	return gps, nil
}

func (gps *Gps) Send(command string) error {
	_, err := serialwrapper.Send(gps.Serial, command)

	return err
}

func (gps *Gps) Read() (string, error) {
	var err error
	var output string
	var eof bool

	gps.BufferStr.Reset()
	for {
		output, eof, err = serialwrapper.Read(gps.Serial)
		_, err = gps.BufferStr.WriteString(output)
		if eof { // end of data
			break
		}
	}

	return gps.GetBufferStr(), err
}

func (gps *Gps) GetBufferStr() string {
	return gps.BufferStr.String()
}

func (gps *Gps) GetQueueElement() *queuewrapper.QueueStructure {
	return &queuewrapper.QueueStructure{GetDataFcn: gps.GetGpsDataInJsonString}
}

func (gps *Gps) GetGpsDataInJsonString() (*queuewrapper.SendData, error) {
	now := time.Now()

	err := gps.GetGpsData()
	if err != nil {
		return nil, err
	}

	gps.Timestamp = int64(now.Unix())
	sendData := &queuewrapper.SendData{}
	sendData.Fcn = config.FCN_NAME_GPS
	sendData.Args = []string{fmt.Sprintf("%f", gps.Longitude), fmt.Sprintf("%f", gps.Latitude),
		fmt.Sprintf("%f", gps.Altitude), fmt.Sprintf("%d", gps.Timestamp)}
	if gps.Longitude*gps.Latitude*gps.Altitude == 0 {
		gps.Led.SetOn()
	}

	return sendData, nil
}

func (gps *Gps) GetGpsData() error {
	gps.Send("AT+CGNSPWR=1") // turn on GNSS power supply
	time.Sleep(500 * time.Millisecond)
	gps.Send("AT+CGPSSTATUS?") // returns the Status of GPS whether it has got FIX or not
	time.Sleep(500 * time.Millisecond)
	gps.Send("AT+CGNSINF") // return GNSS navigation information parsed from NMEA sentences
	time.Sleep(500 * time.Millisecond)

	buffer, err := gps.Read()
	if err != nil {
		return err
	}

	fmt.Println("GNSS output")
	fmt.Println(buffer)

	geoMeta := geoRegex.FindAllStringSubmatch(buffer, -1)
	for i := 0; i < len(geoMeta); i++ {
		if len(geoMeta[i]) >= ALTITUDE {
			longitude, err := strconv.ParseFloat(geoMeta[i][LONGITUDE], 32)
			if err != nil {
				return err
			}

			latitude, err := strconv.ParseFloat(geoMeta[i][LATITUDE], 32)
			if err != nil {
				return err
			}

			altitude, err := strconv.ParseFloat(geoMeta[i][ALTITUDE], 32)
			if err != nil {
				return err
			}

			gps.Longitude = float32(longitude)
			gps.Latitude = float32(latitude)
			gps.Altitude = float32(altitude)
		}
	}

	return nil
}
