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
	iotBarometerIndex = "IotBarometer"
)

const (
	iotBarometerKeyFieldsNumber      = 1
	iotBarometerBasicArgumentsNumber = 4
)

type iotBarometerKey struct {
	ID string `json:"id"`
}

type barometerValue struct {
	Pressure    float32 `json:"pressure"`
	Altitude    float32 `json:"altitude"`
	Temperature float32 `json:"temperature"`
	CustomField string  `json:"customfield"`
	Valid       byte    `json:"valid"`
	Timestamp   int64   `json:"timestamp"`
}

type Barometer struct {
	Key   iotBarometerKey `json:"key"`
	Value barometerValue  `json:"value"`
}

func CreateBarometer() LedgerData {
	return new(Barometer)
}

//argument order
//0			1			2			3
//Pressure	Altitude	Temperature	Timestamp
func (entity *Barometer) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotBarometerBasicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotBarometerBasicArgumentsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	pressureString := args[0]
	if pressureString == "" {
		message := fmt.Sprintf("pressure must be not empty")
		return errors.New(message)
	}
	// checking pressure
	pressure, err := strconv.ParseFloat(pressureString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the pressure: %s", err.Error()))
	}
	entity.Value.Pressure = float32(pressure)

	altitudeString := args[1]
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

	temperatureString := args[2]
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

func (entity *Barometer) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotBarometerKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotBarometerKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Barometer) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Barometer) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotBarometerIndex, compositeKeyParts)
}

func (entity *Barometer) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
