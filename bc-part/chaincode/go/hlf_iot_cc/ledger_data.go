package main

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/statebased"
	"github.com/satori/go.uuid"
	"math/big"
	"strings"
	"sync/atomic"
	"time"
)

var ledgerDataLogger = shim.NewLogger("LedgerData")

const (
	NoticeUnknown = iota
	NoticeRuningType
	NoticeSuccessType
)

type LedgerData interface {
	FillFromArguments(stub shim.ChaincodeStubInterface, args []string) error

	FillFromCompositeKeyParts(compositeKeyParts []string) error

	FillFromLedgerValue(ledgerValue []byte) error

	ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error)

	ToLedgerValue() ([]byte, error)
}

func ExistsIn(stub shim.ChaincodeStubInterface, data LedgerData, index string) bool {

	existResult := false
	compositeKey, err := data.ToCompositeKey(stub)
	if err != nil {
		return false
	}
	collections, err := GetCollectionName(stub, index, []string{""})
	if err != nil {
		return false
	}

	if len(collections) != 0 && collections[0] != "" {
		for _, collectionName := range collections {
			var data []byte
			Logger.Debug(fmt.Sprintf("GetPrivateData. collectionName: %s", collectionName))
			if data, err = stub.GetPrivateData(collectionName, compositeKey); err != nil {
				return existResult
			}
			if data != nil {
				existResult = true
			}
		}
	} else {
		Logger.Debug("GetState")
		var data []byte
		if data, err = stub.GetState(compositeKey); err != nil {
			return existResult
		}
		if data != nil {
			existResult = true
		}
	}

	return existResult
}

func LoadFrom(stub shim.ChaincodeStubInterface, data LedgerData, index string) error {
	var bytes []byte
	compositeKey, err := data.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	collections, err := GetCollectionName(stub, index, []string{""})
	if err != nil {
		message := fmt.Sprintf("cannot get collection name from config: %s", err.Error())
		return errors.New(message)
	}

	if len(collections) != 0 && collections[0] != "" {
		for _, collectionName := range collections {
			Logger.Debug(fmt.Sprintf("GetPrivateData. collectionName: %s", collectionName))
			if bytes, err = stub.GetPrivateData(collectionName, compositeKey); err != nil {
				return err
			}
			if bytes != nil {
				break
			}
		}
	} else {
		Logger.Debug("GetState")
		bytes, err = stub.GetState(compositeKey)
	}

	if err != nil {
		return err
	}

	return data.FillFromLedgerValue(bytes)
}

func UpdateOrInsertIn(stub shim.ChaincodeStubInterface, data LedgerData, index string, participiants []string, endorserRoleType statebased.RoleType) error {
	compositeKey, err := data.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	value, err := data.ToLedgerValue()
	if err != nil {
		return err
	}

	Logger.Debug("PutState")
	if err = stub.PutState(compositeKey, value); err != nil {
		return err
	}
	if len(participiants) != 1 && participiants[0] == "" {
		// set new endorsement policy. Start
		ep, err := statebased.NewStateEP(nil)
		if err != nil {
			return err
		}

		err = ep.AddOrgs(endorserRoleType, participiants[1:]...)
		if err != nil {
			return err
		}
		// set the endorsement policy
		epBytes, err := ep.Policy()
		if err != nil {
			return err
		}

		err = stub.SetStateValidationParameter(compositeKey, epBytes)
		if err != nil {
			return err
		}
		//set new endorsement policy. End
	}

	return nil
}

type FactoryMethod func() LedgerData

type FilterFunction func(data LedgerData) bool

func EmptyFilter(data LedgerData) bool {
	return true
}

func Query(stub shim.ChaincodeStubInterface, index string, partialKey []string,
	createEntry FactoryMethod, filterEntry FilterFunction) ([]byte, error) {

	ledgerDataLogger.Info(fmt.Sprintf("Query(%s) is running", index))
	ledgerDataLogger.Debug("Query " + index)

	entries := []LedgerData{}
	it, err := stub.GetStateByPartialCompositeKey(index, partialKey)
	if err != nil {
		message := fmt.Sprintf("unable to get state by partial composite key %s: %s", index, err.Error())
		ledgerDataLogger.Error(message)
		return nil, errors.New(message)
	}
	defer it.Close()

	entries, err = queryImpl(it, createEntry, stub, filterEntry)
	if err != nil {
		ledgerDataLogger.Error(err.Error())
		return nil, err
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}
	ledgerDataLogger.Debug("Result: " + string(result))

	ledgerDataLogger.Info(fmt.Sprintf("Query(%s) exited without errors", index))
	ledgerDataLogger.Debug("Success: Query " + index)
	return result, nil
}

func queryImpl(it shim.StateQueryIteratorInterface, createEntry FactoryMethod, stub shim.ChaincodeStubInterface,
	filterEntry FilterFunction) ([]LedgerData, error) {

	entries := []LedgerData{}

	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			message := fmt.Sprintf("unable to get an element next to a query iterator: %s", err.Error())
			return nil, errors.New(message)
		}

		ledgerDataLogger.Debug(fmt.Sprintf("Response: {%s, %s}", response.Key, string(response.Value)))

		entry := createEntry()

		if err := entry.FillFromLedgerValue(response.Value); err != nil {
			message := fmt.Sprintf("cannot fill entry value from response value: %s", err.Error())
			return nil, errors.New(message)
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			message := fmt.Sprintf("cannot split response key into composite key parts slice: %s", err.Error())
			return nil, errors.New(message)
		}

		if err := entry.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
			message := fmt.Sprintf("cannot fill entry key from composite key parts: %s", err.Error())
			return nil, errors.New(message)
		}

		if bytes, err := json.Marshal(entry); err == nil {
			ledgerDataLogger.Debug("Entry: " + string(bytes))
		}

		if filterEntry(entry) {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func getOrganization(certificate []byte) (string, error) {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	organization := cert.Issuer.Organization[0]
	return strings.Split(organization, ".")[0], nil
}

func getOrganizationlUnit(certificate []byte) (string, error) {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	organizationalUnit := cert.Issuer.OrganizationalUnit[0]
	return strings.Split(organizationalUnit, ".")[0], nil
}

func getCustomFieldFromCertificate(certificate []byte) (string, error) {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	certEmailAddresses := cert.EmailAddresses

	return strings.Join(certEmailAddresses, ", "), nil
}

func GetCreatorOrganization(stub shim.ChaincodeStubInterface) (string, error) {
	certificate, err := stub.GetCreator()
	if err != nil {
		return "", err
	}
	return getOrganization(certificate)
}

func GetCreatorOrganizationalUnit(stub shim.ChaincodeStubInterface) (string, error) {
	certificate, err := stub.GetCreator()
	if err != nil {
		return "", err
	}
	return getOrganizationlUnit(certificate)
}

func GetCustomFieldFromCertificate(stub shim.ChaincodeStubInterface) (string, error) {
	certificate, err := stub.GetCreator()
	if err != nil {
		return "", err
	}
	return getCustomFieldFromCertificate(certificate)
}

func CheckCertificate(stub shim.ChaincodeStubInterface, certificateString string) (byte, error) {
	if certificateString == "" {
		certificateStringFromStub, err := stub.GetCreator()
		if err != nil {
			return 0, err
		}

		certificateStringFromStub = certificateStringFromStub[strings.Index(string(certificateStringFromStub), "-----") : strings.LastIndex(string(certificateStringFromStub), "-----")+5]
		certificateString = string(certificateStringFromStub)
	}

	certificates := []Certificate{}
	certificateBytes, err := Query(stub, iotCertificateIndex, []string{}, CreateCertificate, EmptyFilter)
	if err != nil {
		return 0, err
	}

	if err := json.Unmarshal(certificateBytes, &certificates); err != nil {
		return 0, err
	}

	for i := 0; i < len(certificates); i++ {
		if certificates[i].Value.Certificate == certificateString {
			return 1, nil
		}
	}

	return 0, nil
}

func GetMSPID(stub shim.ChaincodeStubInterface) (string, error) {
	// Get the client ID object
	mspid := ""
	id, err := cid.New(stub)
	if err != nil {
		message := fmt.Sprintf("Failure getting client ID object: %s", err.Error())
		return mspid, errors.New(message)
	}
	mspid, err = id.GetMSPID()
	if err != nil {
		message := fmt.Sprintf("Failure getting MSPID from client ID object: %s", err.Error())
		return mspid, errors.New(message)
	}
	return mspid, nil
}

func Contains(m map[int][]int, key int) bool {
	_, ok := m[key]
	if !ok {
		return false
	}

	return true
}

func CheckStateValidity(statesAutomaton map[int][]int, oldState, newState int) bool {
	possibleStates, ok := statesAutomaton[oldState]
	if ok {
		for _, state := range possibleStates {
			if state == newState {
				return true
			}
		}
	}

	return false
}

func Notifier(stub shim.ChaincodeStubInterface, typeNotice int) {
	fnc, _ := stub.GetFunctionAndParameters()

	switch typeNotice {
	case NoticeRuningType:
		Logger.Info(fmt.Sprintf("%s.%s is running", chaincodeName, fnc))
		Logger.Debug(fmt.Sprintf("%s.%s", chaincodeName, fnc))
	case NoticeSuccessType:
		Logger.Info(fmt.Sprintf("%s.%s exited without errors", chaincodeName, fnc))
		Logger.Debug(fmt.Sprintf("Success: %s.%s", chaincodeName, fnc))
	default:
		Logger.Debug("Unknown typeNotice: %d", typeNotice)
	}
}

func GetCollectionName(stub shim.ChaincodeStubInterface, index string, participiants []string) ([]string, error) {
	var collectionName []string

	return collectionName, nil
}

func IncUUID(currentStringID string) (string, error) {
	var id uuid.UUID
	var err error
	var currentHexString string

	//parse UUID from string
	if id, err = uuid.FromString(currentStringID); err != nil {
		return "", errors.New(fmt.Sprintf("unable to parse an ID from \"%s\"", currentStringID))
	}

	//build string of hex values
	for _, i := range id {
		if i <= 15 {
			currentHexString = currentHexString + "0" + fmt.Sprintf("%x", i)
		} else {
			currentHexString = currentHexString + fmt.Sprintf("%x", i)
		}
	}

	//convert hex to big int
	currentBigIntValue, _ := new(big.Int).SetString(currentHexString, 16)

	//increment current big int value
	incrementedBigIntValue := currentBigIntValue.Add(currentBigIntValue, big.NewInt(1))

	newHexString := fmt.Sprintf("%x", incrementedBigIntValue)

	//make byte set from hex string
	bs, err := hex.DecodeString(newHexString)
	if err != nil {
		panic(err)
	}

	//replace byte's values of UUID to new values
	for k, _ := range bs {
		id[k] = bs[k]
	}

	//check on UUID
	if id.Version() != uuid.V4 {
		return "", errors.New("wrong ID format; expected UUID version 4")
	}

	return id.String(), nil
}

func UUIDv4FromTXTimestamp(stub shim.ChaincodeStubInterface, delta int) (string, error) {

	//getting transaction Timestamp
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		message := fmt.Sprintf("unable to get transaction timestamp: %s", err.Error())
		Logger.Debug(message)
		return "", errors.New(message)
	}
	//getiing txID
	txid := stub.GetTxID()

	sec := timestamp.Seconds
	nanosec := int64(timestamp.Nanos)

	aTime := time.Unix(sec, nanosec)

	var u uuid.UUID

	//getting hash from time
	h := sha1.New()
	h.Write([]byte(txid))
	sha := h.Sum(nil) // "sha" is uint8 type, encoded in base16
	var clockSeq uint32

	utcTime := aTime.In(time.UTC)
	t := uint64(utcTime.Unix())*10000000 + uint64(utcTime.Nanosecond()/100)
	u[0], u[1], u[2], u[3] = byte(t>>24), byte(t>>16), byte(t>>8), byte(t)
	u[4], u[5] = byte(t>>40), byte(t>>32)
	u[6], u[7] = byte(t>>56)&0x0F, byte(t>>48)

	clock := atomic.AddUint32(&clockSeq, uint32(delta))
	u[8] = byte(clock >> 8)
	u[9] = byte(clock)

	copy(u[10:], sha)

	u[6] = (u[6] & 0x0f) | (byte(4) << 4) // set V4
	u[8] &= 0x3F                          // clear variant
	u[8] |= 0x80                          // set to IETF variant

	return u.String(), nil
}

func (events *Events) EmitEvent(stub shim.ChaincodeStubInterface) error {

	Logger.Debug("### emitEvent started ###")

	for i, value := range events.Values {
		eventAction := value.Action
		var err error

		newID, err := UUIDv4FromTXTimestamp(stub, i+1)
		if err != nil {
			message := fmt.Sprintf(err.Error())
			return errors.New(message)
		}

		event := Event{}
		if err := event.FillFromCompositeKeyParts([]string{newID}); err != nil {
			message := fmt.Sprintf(err.Error())
			return errors.New(message)
		}
		event.Value = value

		creator, err := GetCreatorOrganizationalUnit(stub)
		if err != nil {
			message := fmt.Sprintf("cannot obtain creator's OrganizationalUnit from the certificate: %s", err.Error())
			Logger.Error(message)
			return errors.New(message)
		}
		Logger.Debug("OrganizationalUnit: " + creator)

		//getting transaction Timestamp
		timestamp, err := stub.GetTxTimestamp()
		if err != nil {
			message := fmt.Sprintf("unable to get transaction timestamp: %s", err.Error())
			Logger.Error(message)
			return errors.New(message)
		}

		event.Value.Creator = creator
		event.Value.Timestamp = timestamp.Seconds

		bytes, err := json.Marshal(event)
		if err != nil {
			message := fmt.Sprintf("Error marshaling: %s", err.Error())
			return errors.New(message)
		}
		eventName := eventIndex + "." + eventAction + "." + newID
		events.Keys = append(events.Keys, EventKey{ID: eventName})

		if err := UpdateOrInsertIn(stub, &event, eventIndex, []string{""}, ""); err != nil {
			message := fmt.Sprintf("persistence error: %s", err.Error())
			Logger.Error(message)
			return errors.New(message)
		}

		Logger.Info(fmt.Sprintf("Event set: %s without errors", string(bytes)))
		Logger.Debug(fmt.Sprintf("Success: Event set: %s", string(bytes)))
	}

	generalKey, err := json.Marshal(events.Keys)
	if err != nil {
		message := fmt.Sprintf("Error marshaling: %s", err.Error())
		return errors.New(message)
	}

	if err := stub.SetEvent(string(generalKey), nil); err != nil {
		message := fmt.Sprintf("Error setting event: %s", err.Error())
		return errors.New(message)
	}
	Logger.Debug(fmt.Sprintf("generalEventName: %s", string(generalKey)))

	Logger.Debug("### emitEvent success ###")
	return nil
}
