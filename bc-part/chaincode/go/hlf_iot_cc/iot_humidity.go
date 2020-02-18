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
	iotHumidityIndex = "IotHumidity"
)

const (
	iotHumidityKeyFieldsNumber      = 1
	iotHumidityBasicArgumentsNumber = 3
)

type iotHumidityKey struct {
	ID string `json:"id"`
}

type humidityValue struct {
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	CustomField string  `json:"customfield"`
	Valid       byte    `json:"valid"`
	Timestamp   int64   `json:"timestamp"`
}

type Humidity struct {
	Key   iotHumidityKey `json:"key"`
	Value humidityValue  `json:"value"`
}

func CreateHumidity() LedgerData {
	return new(Humidity)
}

//argument order
//0			1			3
//Humidity	Temperature	Timestamp
func (entity *Humidity) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotHumidityBasicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotHumidityBasicArgumentsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	humidityString := args[0]
	if humidityString == "" {
		message := fmt.Sprintf("humidity must be not empty")
		return errors.New(message)
	}
	// checking humidity
	humidity, err := strconv.ParseFloat(humidityString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the humidity: %s", err.Error()))
	}
	entity.Value.Humidity = float32(humidity)

	temperatureString := args[1]
	if temperatureString == "" {
		message := fmt.Sprintf("temperature must be not empty")
		return errors.New(message)
	}
	// checking temperature
	temperature, err := strconv.ParseFloat(temperatureString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the temperature: %s", err.Error()))
	}
	entity.Value.Temperature = float32(temperature)

	timestampString := args[2]
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

func (entity *Humidity) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotHumidityKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotHumidityKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Humidity) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Humidity) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotHumidityIndex, compositeKeyParts)
}

func (entity *Humidity) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
