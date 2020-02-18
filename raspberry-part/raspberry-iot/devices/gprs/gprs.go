package gprs

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/tarm/serial"
	"hlf-iot/config"
	"hlf-iot/helpers/ca"
	"hlf-iot/helpers/serialwrapper"
	"regexp"
	"strconv"
	"strings"
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

type Gprs struct {
	Serial    *serial.Port
	BufferStr bytes.Buffer
	Cid       int
	Retry     int
	Queue     *list.List
	mode      string
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
	Altitude  float32 `json:"altitude"`
	Timestamp int64   `json:"timestamp"`
}

func Init(cid int) *Gprs {
	gprs := &Gprs{}
	gprs.Cid = cid
	gprs.Retry = 0
	gprs.Queue = list.New()
	gprs.Serial = serialwrapper.Init()
	gprs.BufferStr = bytes.Buffer{}
	gprs.StartModule()

	return gprs
}

func (gprs *Gprs) StartModule() {
	gprs.Send("AT+CMEE=2") // reporting of mobile equipment errors
	time.Sleep(100 * time.Millisecond)
	gprs.Send("AT+CREG?") // network Registration
	time.Sleep(100 * time.Millisecond)
	gprs.Send("AT+CGATT?")
	time.Sleep(100 * time.Millisecond)
	gprs.Send("AT+CSQ")
	time.Sleep(100 * time.Millisecond)
	gprs.Send("AT+SAPBR=3,1,\"CONTYPE\",\"GPRS\"")
	time.Sleep(500 * time.Millisecond)
	gprs.Send(fmt.Sprintf("AT+SAPBR=3,1,\"APN\",\"%s\"", config.GPRS_APN))
	time.Sleep(500 * time.Millisecond)
}

func (gprs *Gprs) RebootModule() {
	gprs.Send("AT+CFUN=1,1")
	time.Sleep(5000 * time.Millisecond)
}

func (gprs *Gprs) AddToQueue(queue *config.QueueStructure) {
	gprs.Queue.PushBack(queue)
	fmt.Println("========================================")
	fmt.Printf("Queue add: %s\n", queue.JsonString)
	fmt.Printf("Queue length +1: %d\n", gprs.Queue.Len())
	fmt.Println("========================================")
}

func (gprs *Gprs) StartDaemon() {
	for {
		e := gprs.Queue.Front()
		if e != nil {
			gprs.Retry = 0
			queue := e.Value.(*config.QueueStructure)
			if queue.GetDataFcn != nil {
				queue.JsonString = queue.GetDataFcn()
			}
			if queue.JsonString != nil {
				fmt.Println()
				fmt.Printf("Sending data: %s", queue.JsonString)
				fmt.Println()
				gprs.PrepareRequest(queue.JsonString, config.GPRS_API_URL, queue.HttpAction, queue.CheckResponseFcn)
			} else {
				fmt.Println("Remove without sending")
			}
			gprs.Queue.Remove(e)
			fmt.Println("========================================")
			fmt.Printf("Queue remove: %s\n", queue.JsonString)
			fmt.Printf("Queue length -1: %d\n", gprs.Queue.Len())
			fmt.Println("========================================")
		}
		time.Sleep(config.DELAY_FOR_DAEMON_MILLISECONDS * time.Millisecond)
	}
}

func (gprs *Gprs) PrepareRequest(data *config.SendData, url string, httpAction int, checkFcn func(string) bool) {
	fabricCa := ca.GetInstance()

	success := gprs.SendRequest(fabricCa.ProposalReq(data.Fcn, data.Args))
	if !success {
		gprs.PrepareRequest(data, url, httpAction, checkFcn)
	}
	time.Sleep(500 * time.Millisecond)

	fabricCa.SignProposal()
	time.Sleep(5000 * time.Millisecond)

	success = gprs.SendRequest(fabricCa.BroadcastPayloadReq())
	if !success {
		gprs.PrepareRequest(data, url, httpAction, checkFcn)
	}
	time.Sleep(500 * time.Millisecond)

	fabricCa.SignBroadcastPayload()
	time.Sleep(500 * time.Millisecond)

	success = gprs.SendRequest(fabricCa.BroadcastReq())
	if !success {
		gprs.PrepareRequest(data, url, httpAction, checkFcn)
	}
}

func (gprs *Gprs) SendRequest(sendElement *config.GprsSendElementStructure) bool {
	fmt.Println("sendElement.Data, sendElement.Url")
	fmt.Println(sendElement.Data, sendElement.Url)
	fmt.Println()
	response := gprs.Request(sendElement.Data, sendElement.Url, sendElement.HttpAction)

	return sendElement.CheckFcn(response)
}

func (gprs *Gprs) Request(dataJson, url string, httpAction int) string {
	gprs.Send("AT+SAPBR=1,1")
	time.Sleep(500 * time.Millisecond)
	gprs.Send("AT+HTTPINIT") // Initialize HTTP Service
	time.Sleep(500 * time.Millisecond)
	gprs.Send(fmt.Sprintf("AT+HTTPPARA=\"CID\",%d", gprs.Cid))
	time.Sleep(500 * time.Millisecond)
	if httpAction != 0 {
		gprs.Send(fmt.Sprintf("AT+HTTPPARA=\"URL\",\"%s\"", url))
		time.Sleep(500 * time.Millisecond)
		gprs.Send("AT+HTTPPARA=\"CONTENT\",\"application/json\"")
		time.Sleep(500 * time.Millisecond)
		gprs.Send(fmt.Sprintf("AT+HTTPDATA=%d,%d", len(strings.Replace(dataJson, `"`, `\"`, -1))+1, config.GPRS_WAITING_TIME_MILLISECONDS))
		time.Sleep(500 * time.Millisecond)
		gprs.Send(dataJson)
		time.Sleep(3000 * time.Millisecond)
	}
	gprs.Send(fmt.Sprintf("AT+HTTPPARA=\"URL\",\"%s\"", url))
	time.Sleep(500 * time.Millisecond)
	gprs.Send(fmt.Sprintf("AT+HTTPACTION=%d", httpAction))
	time.Sleep(3000 * time.Millisecond)
	gprs.Send(fmt.Sprintf("AT+HTTPREAD=0,%d", config.GPRS_BYTE_SIZE))
	time.Sleep(3000 * time.Millisecond)
	gprs.Send("AT+SAPBR=0,1")
	time.Sleep(500 * time.Millisecond)
	gprs.Send("AT+HTTPTERM")
	time.Sleep(2000 * time.Millisecond)

	return gprs.Read()
}

func (gprs *Gprs) Send(command string) {
	serialwrapper.Send(gprs.Serial, command)
}

func (gprs *Gprs) Read() string {
	gprs.BufferStr.Reset()
	for {
		output, eof := serialwrapper.Read(gprs.Serial)
		gprs.BufferStr.WriteString(output)
		if eof { // end of data
			break
		}
	}

	return gprs.GetBufferStr()
}

func (gprs *Gprs) GetBufferStr() string {
	return gprs.BufferStr.String()
}

func (gprs *Gprs) GetQueueElement() *config.QueueStructure {
	queueStructure := &config.QueueStructure{}
	queueStructure.GetDataFcn = GetGpsDataInJsonString
	queueStructure.CheckResponseFcn = config.CheckInsertToBC
	queueStructure.JsonString = &config.SendData{}
	queueStructure.HttpAction = config.GPRS_HTTPACTION_POST

	return queueStructure
}

func GetGpsDataInJsonString() *config.SendData {
	gprs := Init(1)
	now := time.Now()
	gprs.GetGpsData()
	gprs.Timestamp = int64(now.Unix())
	sendData := &config.SendData{}
	sendData.Fcn = config.FCN_NAME_GPS
	sendData.Args = []string{fmt.Sprintf("%f", gprs.Longitude), fmt.Sprintf("%f", gprs.Latitude),
		fmt.Sprintf("%f", gprs.Altitude), fmt.Sprintf("%d", gprs.Timestamp)}

	return sendData
}

func (gprs *Gprs) GetGpsData() {
	gprs.Send("AT+CGNSPWR=1") // turn on GNSS power supply
	time.Sleep(500 * time.Millisecond)
	gprs.Send("AT+CGPSSTATUS?") // returns the Status of GPS whether it has got FIX or not
	time.Sleep(500 * time.Millisecond)
	gprs.Send("AT+CGNSINF") // return GNSS navigation information parsed from NMEA sentences
	time.Sleep(500 * time.Millisecond)

	gprs.Read()
	buffer := gprs.GetBufferStr()

	fmt.Println("GNSS output")
	fmt.Println(buffer)

	geoMeta := geoRegex.FindAllStringSubmatch(buffer, -1)
	for i := 0; i < len(geoMeta); i++ {
		if len(geoMeta[i]) >= ALTITUDE {
			longitude, _ := strconv.ParseFloat(geoMeta[i][LONGITUDE], 32)
			latitude, _ := strconv.ParseFloat(geoMeta[i][LATITUDE], 32)
			altitude, _ := strconv.ParseFloat(geoMeta[i][ALTITUDE], 32)
			gprs.Longitude = float32(longitude)
			gprs.Latitude = float32(latitude)
			gprs.Altitude = float32(altitude)
		}
	}
}
