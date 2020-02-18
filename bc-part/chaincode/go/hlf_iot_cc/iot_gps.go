package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/satori/go.uuid"
	"strconv"
)

const (
	iotGpsIndex = "IotGps"
)

const (
	iotGpsKeyFieldsNumber      = 1
	iotGpsBasicArgumentsNumber = 4
)

type iotGpsKey struct {
	ID string `json:"id"`
}

type gpsValue struct {
	Longitude   float32 `json:"longitude"`
	Latitude    float32 `json:"latitude"`
	Altitude    float32 `json:"altitude"`
	CustomField string  `json:"customfield"`
	Valid       byte    `json:"valid"`
	Timestamp   int64   `json:"timestamp"`
}

type Gps struct {
	Key   iotGpsKey `json:"key"`
	Value gpsValue  `json:"value"`
}

func CreateGps() LedgerData {
	return new(Gps)
}

//argument order
//0			1			2			3
//Longitude	Latitude	Altitude	Timestamp
func (entity *Gps) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotGpsBasicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotGpsBasicArgumentsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	longitudeString := args[0]
	if longitudeString == "" {
		message := fmt.Sprintf("longitude must be not empty")
		return errors.New(message)
	}
	// checking longitude
	longitude, err := strconv.ParseFloat(longitudeString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the longitude: %s", err.Error()))
	}
	entity.Value.Longitude = float32(longitude)

	latitudeString := args[1]
	if latitudeString == "" {
		message := fmt.Sprintf("latitude must be not empty")
		return errors.New(message)
	}
	// checking latitude
	latitude, err := strconv.ParseFloat(latitudeString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the latitude: %s", err.Error()))
	}
	entity.Value.Latitude = float32(latitude)

	altitudeString := args[2]
	if altitudeString == "" {
		message := fmt.Sprintf("altitude must be not empty")
		return errors.New(message)
	}
	// checking altitude
	altitude, err := strconv.ParseFloat(altitudeString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the altitude: %s", err.Error()))
	}
	entity.Value.Altitude = float32(altitude)

	timestampString := args[3]
	if timestampString == "" {
		message := fmt.Sprintf("timestamp must be not empty")
		return errors.New(message)
	}
	timestamp, err := strconv.ParseInt(timestampString, 10, 64)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the timestamp: %s", err.Error()))
	}
	if timestamp < 0 {
		return errors.New("timestamp must be larger than zero")
	}
	entity.Value.Timestamp = int64(timestamp)

	//get custom field from certificate
	customField, err := GetCustomFieldFromCertificate(stub)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot obtain creator's custom field from the certificate: %s", err.Error()))
	}
	entity.Value.CustomField = customField

	//check is certificate valid
	valid, err := CheckCertificate(stub, "")
	if err != nil {
		return errors.New(fmt.Sprintf("cannot check the certificate: %s", err.Error()))
	}
	entity.Value.Valid = valid

	return nil
}

func (entity *Gps) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotGpsKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotGpsKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Gps) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Gps) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotGpsIndex, compositeKeyParts)
}

func (entity *Gps) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
