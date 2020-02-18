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
	iotVibrationIndex = "IotVibration"
)

const (
	iotVibrationKeyFieldsNumber      = 1
	iotVibrationBasicArgumentsNumber = 2
)

type iotVibrationKey struct {
	ID string `json:"id"`
}

type vibrationValue struct {
	Vibration   uint   `json:"vibration"`
	CustomField string `json:"customfield"`
	Valid       byte   `json:"valid"`
	Timestamp   int64  `json:"timestamp"`
}

type Vibration struct {
	Key   iotVibrationKey `json:"key"`
	Value vibrationValue  `json:"value"`
}

func CreateVibration() LedgerData {
	return new(Vibration)
}

//argument order
//0			1
//Vibration	Timestamp
func (entity *Vibration) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotVibrationKeyFieldsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotVibrationKeyFieldsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	vibrationString := args[0]
	if vibrationString == "" {
		message := fmt.Sprintf("vibration must be not empty")
		return errors.New(message)
	}
	// checking vibration
	vibration, err := strconv.ParseUint(vibrationString, 10, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the vibration: %s", err.Error()))
	}
	entity.Value.Vibration = uint(vibration)

	timestampString := args[1]
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

func (entity *Vibration) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotVibrationKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotVibrationKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Vibration) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Vibration) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotVibrationIndex, compositeKeyParts)
}

func (entity *Vibration) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
