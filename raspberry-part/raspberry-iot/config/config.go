package config

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"time"
)

//response
var responseRegex = regexp.MustCompile(`{"result": "([\s\S]*?)"`)

type SendData struct {
	Fcn  string   `json:"fcn"`
	Args []string `json:"args"`
}

type QueueStructure struct {
	GetDataFcn       func() *SendData  `json:"getdatafcn"`
	CheckResponseFcn func(string) bool `json:"checkresponsefcn"`
	JsonString       *SendData         `json:"jsonstring"`
	HttpAction       int               `json:"httpaction"`
}

type GprsSendElementStructure struct {
	Data       string            `json:"data"`
	Url        string            `json:"url"`
	HttpAction int               `json:"httpaction"`
	CheckFcn   func(string) bool `json:"checkfcn"`
}

const (
	CHANNEL_ID   = "common"
	CHAINCODE_ID = "hlf_iot_cc"
	MSP_ID       = "hlfiotMSP"
)

var EndorsementPeers = []string{"hlfiot/peer0", "device/peer0"}

const (
	GPRS_WAITING_TIME_MILLISECONDS = 5000
	GPRS_BYTE_SIZE                 = 319488
	GPRS_RETRY_COUNT               = 5
	API_BASE_URL                   = "https://hlfiot.dev.altoros.com/api/"
	GPRS_API_URL                   = API_BASE_URL + "channels/common/chaincodes/" + CHAINCODE_ID
	GPRS_APN                       = "vmi.velcom.by"
)

const (
	CALLBACK_SENSORS_DELAY_TIME_SECONDS       = 180.0
	DELAY_FOR_GATHERING_DATA_IN_CYCLE_SECONDS = 900
	DELAY_FOR_DAEMON_MILLISECONDS             = 500
)

const (
	GPRS_HTTPACTION_GET = iota
	GPRS_HTTPACTION_POST
)

const (
	CA_LOGIN                 = "admin"
	CA_PASSWORD              = "adminpw"
	CA_TBS_CSR_URL           = API_BASE_URL + "ca/tbs-csr"
	CA_ENROLL_CSR_URL        = API_BASE_URL + "ca/enroll-csr"
	CA_PROPOSAL_URL          = API_BASE_URL + "tx/proposal"
	CA_BROADCAST_PAYLOAD_URL = API_BASE_URL + "tx/prepare-broadcast"
	CA_BROADCAST_URL         = API_BASE_URL + "tx/broadcast"
)

const (
	FCN_NAME_HUMIDITY  = "addIotHumidity"
	FCN_NAME_BAROMETER = "addIotBarometer"
	FCN_NAME_GYROSCOPE = "addIotGyroscope"
	FCN_NAME_GPS       = "addIotGps"
	FCN_NAME_VIBRATION = "addIotVibration"
	FCN_NAME_LIGHT     = "addIotLight"
)

func B2i(b bool) uint {
	if b {
		return 1
	}
	return 0
}

func ItsTime(time1, time2 time.Time) bool {
	diff := time2.Sub(time1).Seconds()
	return diff > CALLBACK_SENSORS_DELAY_TIME_SECONDS
}

func B64Decode(str string) (buf []byte, err error) {
	return base64.StdEncoding.DecodeString(str)
}

func CheckInsertToBC(buffer string) bool {
	fmt.Println("output")
	fmt.Println(buffer)
	response := responseRegex.FindAllStringSubmatch(buffer, -1)

	return len(response) > 0
}
