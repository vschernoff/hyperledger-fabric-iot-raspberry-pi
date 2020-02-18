package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/satori/go.uuid"
)

const (
	eventIndex = "Event"
)

const (
	eventKeyFieldsNumber      = 1
	eventBasicArgumentsNumber = 5
)

type EventKey struct {
	ID string `json:"id"`
}

type EventValue struct {
	Timestamp  int64       `json:"timestamp"`
	Creator    string      `json:"creator"`
	EntityType string      `json:"entityType"`
	EntityID   string      `json:"entityID"`
	Action     string      `json:"action"`
	Other      interface{} `json:"other"`
}

type Event struct {
	Key   EventKey   `json:"key"`
	Value EventValue `json:"value"`
}

type Events struct {
	Keys   []EventKey   `json:"generalKey"`
	Values []EventValue `json:"values"`
}

func CreateEvent() LedgerData {
	return new(Event)
}

//argument order
//0
//ID
func (entity *Event) FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) < eventBasicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", eventBasicArgumentsNumber))
	}
	return nil
}

func (entity *Event) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < eventKeyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", eventKeyFieldsNumber))
	}

	if id, err := uuid.FromString(compositeKeyParts[0]); err != nil {
		return errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", compositeKeyParts[0]))
	} else if id.Version() != uuid.V4 {
		return errors.New("wrong ID format; expected UUID version 4")
	}

	entity.Key.ID = compositeKeyParts[0]

	return nil
}

func (entity *Event) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &entity.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (entity *Event) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string{
		entity.Key.ID,
	}

	return stub.CreateCompositeKey(eventIndex, compositeKeyParts)
}

func (entity *Event) ToLedgerValue() ([]byte, error) {
	return json.Marshal(entity.Value)
}
