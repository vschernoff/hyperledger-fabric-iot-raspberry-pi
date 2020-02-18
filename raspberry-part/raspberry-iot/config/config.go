package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hlf-iot/helpers/httpwrapper"
	"math/big"
	"path"
	"strings"
	"time"
)

type ResultStructure struct {
	ResultHash string `json:"result"`
}

type GprsSendElementStructure struct {
	Data       string            `json:"data"`
	Url        string            `json:"url"`
	HttpAction int               `json:"httpaction"`
	CheckFcn   func(string) bool `json:"checkfcn"`
}

type CustomField struct {
	CustomFieldValue string `json:"customField"`
}

type SensorsActivityGrid struct {
	LedBad           bool `json:"ledbad"`
	LedSuccessEnroll bool `json:"ledsuccessenroll"`
	LedBadGps        bool `json:"ledbadgps"`
	GpsSensor        bool `json:"gpssensor"`
	HumiditySensor   bool `json:"humiditysensor"`
	BarometerSensor  bool `json:"barometersensor"`
	GyroscopeSensor  bool `json:"gyroscopesensor"`
	VibrationSensor  bool `json:"vibrationsensor"`
	LightSensor      bool `json:"lightsensor"`
}

type Key struct {
	PrivateKeyHash string `json:"private_key_hash"`
	Certificate    string `json:"certificate"`
}

func NewSensorsActivityGrid(value bool) SensorsActivityGrid {
	sensorsActivityGrid := SensorsActivityGrid{}
	sensorsActivityGrid.LedBad = value
	sensorsActivityGrid.LedSuccessEnroll = value
	sensorsActivityGrid.LedBadGps = value
	sensorsActivityGrid.GpsSensor = value
	sensorsActivityGrid.HumiditySensor = value
	sensorsActivityGrid.BarometerSensor = value
	sensorsActivityGrid.GyroscopeSensor = value
	sensorsActivityGrid.VibrationSensor = value
	sensorsActivityGrid.LightSensor = value

	return sensorsActivityGrid
}

const (
	CHANNEL_ID   = "common"
	CHAINCODE_ID = "hlf_iot_cc"
	MSP_ID       = "hlfiotMSP"
)

var EndorsementPeers = []string{"hlfiot/peer0", "device/peer0"}

const (
	API_BASE_URL        = "https://hlfiot.dev.altoros.com/api/"
	CA_CUSTOM_FIELD_URL = "https://cfhlfiot2.dev.altoros.com/get-custom-field"
)

const (
	CALLBACK_SENSORS_DELAY_TIME_SECONDS       = 10
	DELAY_FOR_GATHERING_DATA_IN_CYCLE_SECONDS = 30
	DELAY_FOR_DAEMON_MILLISECONDS             = 500
)

const (
	CA_LOGIN                 = "admin"
	CA_PASSWORD              = "adminpw"
	CA_CUSTOM_FIELD          = "raspberrytestemail@google.com"
	CA_TBS_CSR_URL           = API_BASE_URL + "ca/tbs-csr"
	CA_ENROLL_CSR_URL        = API_BASE_URL + "ca/enroll-csr"
	CA_PROPOSAL_URL          = API_BASE_URL + "tx/proposal"
	CA_BROADCAST_PAYLOAD_URL = API_BASE_URL + "tx/prepare-broadcast"
	CA_BROADCAST_URL         = API_BASE_URL + "tx/broadcast"
)

const (
	FCN_NAME_HUMIDITY              = "addIotHumidity"
	FCN_NAME_BAROMETER             = "addIotBarometer"
	FCN_NAME_GYROSCOPE             = "addIotGyroscope"
	FCN_NAME_GPS                   = "addIotGps"
	FCN_NAME_VIBRATION             = "addIotVibration"
	FCN_NAME_LIGHT                 = "addIotLight"
	FCN_NAME_CHECK_IOT_CERTIFICATE = "checkIotCertificate"
)

const (
	LED_PIN_SUCCESS_ENROLL  = 27
	LED_PIN_BAD_GPS_DATA    = 25
	LED_PIN_BAD_SENSOR_DATA = 18
)

const (
	MEDIA_ROOT_PATH            = "/media/usb/"
	CERTIFICATE_FILE_EXTENSION = "pem"
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

func CheckInsertToBC(buffer string) (bool, error) {
	fmt.Println("CheckInsertToBC output")
	fmt.Println(buffer)

	var check bool
	resultPayload := &ResultStructure{}
	if err := json.Unmarshal([]byte(buffer), &resultPayload); err != nil {
		return check, err
	}

	if len(resultPayload.ResultHash) > 0 {
		check = true
	} else {
		check = false
	}

	return check, nil
}

func GetCustomField(defaultField string) string {
	response, err := httpwrapper.GetReq(CA_CUSTOM_FIELD_URL)
	if err != nil {
		return defaultField
	}

	customField := &CustomField{}
	if err := json.Unmarshal([]byte(response), &customField); err != nil {
		return defaultField
	}

	if len(customField.CustomFieldValue) > 0 {
		return customField.CustomFieldValue
	} else {
		return defaultField
	}
}

func HexToPrivateKey(hexStr string) (*ecdsa.PrivateKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}

	k := new(big.Int)
	k.SetBytes(bytes)

	privateKey := new(ecdsa.PrivateKey)
	curve := elliptic.P256()
	privateKey.PublicKey.Curve = curve
	privateKey.D = k
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())

	return privateKey, nil
}

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}
