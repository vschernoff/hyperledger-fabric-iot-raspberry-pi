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
	iotGyroscopeIndex = "IotGyroscope"
)

const (
	iotGyroscopeKeyFieldsNumber      = 1
	iotGyroscopeBasicArgumentsNumber = 13
)

type iotGyroscopeKey struct {
	ID string `json:"id"`
}

type gyroscopeValue struct {
	Xout                   float32 `json:"xout"`
	XoutScaled             float32 `json:"xoutscaled"`
	Yout                   float32 `json:"yout"`
	YoutScaled             float32 `json:"youtscaled"`
	Zout                   float32 `json:"zout"`
	ZoutScaled             float32 `json:"zoutscaled"`
	AccelerationXout       float32 `json:"accelerationxout"`
	AccelerationXoutScaled float32 `json:"accelerationxoutscaled"`
	AccelerationYout       float32 `json:"accelerationyout"`
	AccelerationYoutScaled float32 `json:"accelerationyoutscaled"`
	AccelerationZout       float32 `json:"accelerationZout"`
	AccelerationZoutScaled float32 `json:"accelerationZoutscaled"`
	CustomField            string  `json:"customfield"`
	Valid                  byte    `json:"valid"`
	Timestamp              int64   `json:"timestamp"`
}

type Gyroscope struct {
	Key   iotGyroscopeKey `json:"key"`
	Value gyroscopeValue  `json:"value"`
}

func CreateGyroscope() LedgerData {
	return new(Gyroscope)
}

//argument order
//0		1			2		3			4		5			6					7						8					9						10					11						12
//Xout	XoutScaled	Yout	YoutScaled	Zout	ZoutScaled	AccelerationXout	AccelerationXoutScaled	AccelerationYout	AccelerationYoutScaled	AccelerationZout	AccelerationZoutScaled	Timestamp
func (entity *Gyroscope) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotGyroscopeBasicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotGyroscopeBasicArgumentsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	xOutString := args[0]
	if xOutString == "" {
		message := fmt.Sprintf("xOut must be not empty")
		return errors.New(message)
	}
	// checking xout
	xOut, err := strconv.ParseFloat(xOutString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the xOut: %s", err.Error()))
	}
	entity.Value.Xout = float32(xOut)

	xOutScaledString := args[1]
	if xOutScaledString == "" {
		message := fmt.Sprintf("xOutScaled must be not empty")
		return errors.New(message)
	}
	// checking XoutScaled
	xOutScaled, err := strconv.ParseFloat(xOutScaledString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the xOutScaled: %s", err.Error()))
	}
	entity.Value.XoutScaled = float32(xOutScaled)

	yOutString := args[2]
	if yOutString == "" {
		message := fmt.Sprintf("yOut must be not empty")
		return errors.New(message)
	}
	// checking yOut
	yOut, err := strconv.ParseFloat(yOutString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the yOut: %s", err.Error()))
	}
	entity.Value.Yout = float32(yOut)

	yOutScaledString := args[3]
	if yOutScaledString == "" {
		message := fmt.Sprintf("yOutScaled must be not empty")
		return errors.New(message)
	}
	// checking yOutScaled
	yOutScaled, err := strconv.ParseFloat(yOutScaledString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the yOutScaled: %s", err.Error()))
	}
	entity.Value.YoutScaled = float32(yOutScaled)

	zOutString := args[4]
	if zOutString == "" {
		message := fmt.Sprintf("zOut must be not empty")
		return errors.New(message)
	}
	// checking zOut
	zOut, err := strconv.ParseFloat(zOutString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the zOut: %s", err.Error()))
	}
	entity.Value.Zout = float32(zOut)

	zOutScaledString := args[5]
	if zOutScaledString == "" {
		message := fmt.Sprintf("zOutScaled must be not empty")
		return errors.New(message)
	}
	// checking zOutScaled
	zOutScaled, err := strconv.ParseFloat(zOutScaledString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the zOutScaled: %s", err.Error()))
	}
	entity.Value.ZoutScaled = float32(zOutScaled)

	accelerationXoutString := args[6]
	if accelerationXoutString == "" {
		message := fmt.Sprintf("accelerationXout must be not empty")
		return errors.New(message)
	}
	// checking accelerationXout
	accelerationXout, err := strconv.ParseFloat(accelerationXoutString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the accelerationXout: %s", err.Error()))
	}
	entity.Value.AccelerationXout = float32(accelerationXout)

	accelerationXoutScaledString := args[7]
	if accelerationXoutScaledString == "" {
		message := fmt.Sprintf("accelerationXoutScaled must be not empty")
		return errors.New(message)
	}
	// checking accelerationXoutScaled
	accelerationXoutScaled, err := strconv.ParseFloat(accelerationXoutScaledString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the accelerationXoutScaled: %s", err.Error()))
	}
	entity.Value.AccelerationXoutScaled = float32(accelerationXoutScaled)

	accelerationYoutString := args[8]
	if accelerationYoutString == "" {
		message := fmt.Sprintf("accelerationYout must be not empty")
		return errors.New(message)
	}
	// checking accelerationXout
	accelerationYout, err := strconv.ParseFloat(accelerationYoutString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the accelerationYout: %s", err.Error()))
	}
	entity.Value.AccelerationYout = float32(accelerationYout)

	accelerationYoutScaledString := args[9]
	if accelerationYoutScaledString == "" {
		message := fmt.Sprintf("accelerationYoutScaled must be not empty")
		return errors.New(message)
	}
	// checking accelerationYoutScaled
	accelerationYoutScaled, err := strconv.ParseFloat(accelerationYoutScaledString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the accelerationYoutScaled: %s", err.Error()))
	}
	entity.Value.AccelerationYoutScaled = float32(accelerationYoutScaled)

	accelerationZoutString := args[10]
	if accelerationZoutString == "" {
		message := fmt.Sprintf("accelerationZout must be not empty")
		return errors.New(message)
	}
	// checking accelerationZout
	accelerationZout, err := strconv.ParseFloat(accelerationZoutString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the accelerationZout: %s", err.Error()))
	}
	entity.Value.AccelerationZout = float32(accelerationZout)

	accelerationZoutScaledString := args[11]
	if accelerationZoutScaledString == "" {
		message := fmt.Sprintf("accelerationZoutScaled must be not empty")
		return errors.New(message)
	}
	// checking accelerationZoutScaled
	accelerationZoutScaled, err := strconv.ParseFloat(accelerationZoutScaledString, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to parse the accelerationZoutScaled: %s", err.Error()))
	}
	entity.Value.AccelerationZoutScaled = float32(accelerationZoutScaled)

	timestampString := args[12]
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

func (entity *Gyroscope) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotGyroscopeKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotGyroscopeKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Gyroscope) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Gyroscope) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotGyroscopeIndex, compositeKeyParts)
}

func (entity *Gyroscope) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
