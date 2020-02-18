package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/satori/go.uuid"
	"strings"
)

const (
	iotCertificateIndex = "IotCertificate"
)

const (
	iotCertificateKeyFieldsNumber      = 1
	iotCertificateBasicArgumentsNumber = 1
)

type iotCertificateKey struct {
	ID string `json:"id"`
}

type certificateValue struct {
	Certificate string `json:"certificate"`
	Timestamp   int64  `json:"timestamp"`
}

type Certificate struct {
	Key   iotCertificateKey `json:"key"`
	Value certificateValue  `json:"value"`
}

func CreateCertificate() LedgerData {
	return new(Certificate)
}

//argument order
//0
//Certificate
func (entity *Certificate) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < iotCertificateBasicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", iotCertificateBasicArgumentsNumber))
	}

	u, err := uuid.NewV4()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to generate uuid: %s", err.Error()))
	}
	entity.Key.ID = u.String()

	certificateString := args[0]
	if certificateString == "" {
		message := fmt.Sprintf("certificate must be not empty")
		return errors.New(message)
	}
	certificateString = certificateString[strings.Index(certificateString, "-----") : strings.LastIndex(certificateString, "-----")+5]
	entity.Value.Certificate = certificateString

	return nil
}

func (entity *Certificate) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < iotCertificateKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", iotCertificateKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Certificate) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Certificate) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(iotCertificateIndex, compositeKeyParts)
}

func (entity *Certificate) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
