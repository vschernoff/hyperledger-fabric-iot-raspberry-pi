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
	iotLightIndex = "IotLight"
)

const (
	iotLightKeyFieldsNumber      = 1
	iotLightBasicArgumentsNumber = 2
)

type iotLightKey struct {
	ID string `json:"id"`
}

type lightValue struct {
	Light       uint   `json:"light"`
	CustomField string `json:"customfield"`
	Valid       byte   `json:"valid"`
	Timestamp   int64  `json:"timestamp"`
}

type Light struct {
	Key   iotLightKey `json:"key"`
	Value lightValue  `json:"value"`
}

func CreateLight() LedgerData {
	return new(Light)
}

//argument order
//0			1
//Vibration	Timestamp
func (entity *Light) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotLightKeyFieldsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotLightKeyFieldsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	lightString := args[0]
	if lightString == "" {
		message := fmt.Sprintf("light must be not empty")
		return errors.New(message)
	}
	// checking light
	light, err := strconv.ParseUint(lightString, 10, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the light: %s", err.Error()))
	}
	entity.Value.Light = uint(light)

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

func (entity *Light) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotLightKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotLightKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Light) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Light) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotLightIndex, compositeKeyParts)
}

func (entity *Light) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
