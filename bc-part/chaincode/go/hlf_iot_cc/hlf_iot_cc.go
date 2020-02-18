package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SupplyChainChaincode struct {
}

func (cc *SupplyChainChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	Logger.Debug("Init")

	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	Logger.Debug("Invoke")

	function, args := stub.GetFunctionAndParameters()
	if function == "addIotGps" {
		return cc.addIotGps(stub, args)
	} else if function == "listIotGps" {
		return cc.listIotGps(stub, args)
	} else if function == "addIotBarometer" {
		return cc.addIotBarometer(stub, args)
	} else if function == "listIotBarometer" {
		return cc.listIotBarometer(stub, args)
	} else if function == "addIotGyroscope" {
		return cc.addIotGyroscope(stub, args)
	} else if function == "listIotGyroscope" {
		return cc.listIotGyroscope(stub, args)
	} else if function == "addIotHumidity" {
		return cc.addIotHumidity(stub, args)
	} else if function == "listIotHumidity" {
		return cc.listIotHumidity(stub, args)
	} else if function == "addIotVibration" {
		return cc.addIotVibration(stub, args)
	} else if function == "listIotVibration" {
		return cc.listIotVibration(stub, args)
	} else if function == "addIotLight" {
		return cc.addIotLight(stub, args)
	} else if function == "listIotLight" {
		return cc.listIotLight(stub, args)
	} else if function == "addIotCertificate" {
		return cc.addIotCertificate(stub, args)
	} else if function == "checkIotCertificate" {
		return cc.checkIotCertificate(stub, args)
	}
	// (optional) add other query functions

	fnList := "{addIotGps, listIotGps, addIotBarometer, listIotBarometer, addIotGyroscope, listIotGyroscope, addIotHumidity, listIotHumidity, addIotVibration, listIotVibration, addIotLight, listIotLight, addIotCertificate, checkIotCertificate}"
	message := fmt.Sprintf("invalid invoke function name: expected one of %s, got %s", fnList, function)
	Logger.Debug(message)

	return pb.Response{Status: 400, Message: message}
}

//0			1			2			3
//Longitude	Latitude	Altitude	Timestamp
func (cc *SupplyChainChaincode) addIotGps(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	gps := Gps{}
	if err := gps.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a gps data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(gps); err == nil {
		Logger.Debug("gps: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &gps, iotGpsIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotGpsIndex
	eventValue.EntityID = gps.Key.ID
	eventValue.Other = gps.Value
	eventValue.Action = eventAddIotGps

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) listIotGps(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)
	gps := []Gps{}
	gpsBytes, err := Query(stub, iotGpsIndex, []string{}, CreateGps, EmptyFilter)
	if err != nil {
		message := fmt.Sprintf("unable to perform method: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}
	if err := json.Unmarshal(gpsBytes, &gps); err != nil {
		message := fmt.Sprintf("unable to unmarshal query result: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	resultBytes, err := json.Marshal(gps)

	Logger.Debug("Result: " + string(resultBytes))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(resultBytes)
}

//0			1			2			3
//Pressure	Altitude	Temperature	Timestamp
func (cc *SupplyChainChaincode) addIotBarometer(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	barometer := Barometer{}
	if err := barometer.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a barometer data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(barometer); err == nil {
		Logger.Debug("barometer: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &barometer, iotBarometerIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotBarometerIndex
	eventValue.EntityID = barometer.Key.ID
	eventValue.Other = barometer.Value
	eventValue.Action = eventAddIotBarometer

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) listIotBarometer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)
	barometer := []Barometer{}
	barometerBytes, err := Query(stub, iotBarometerIndex, []string{}, CreateBarometer, EmptyFilter)
	if err != nil {
		message := fmt.Sprintf("unable to perform method: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}
	if err := json.Unmarshal(barometerBytes, &barometer); err != nil {
		message := fmt.Sprintf("unable to unmarshal query result: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	resultBytes, err := json.Marshal(barometer)

	Logger.Debug("Result: " + string(resultBytes))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(resultBytes)
}

//0		1			2		3			4		5			6					7						8					9						10					11						12
//Xout	XoutScaled	Yout	YoutScaled	Zout	ZoutScaled	AccelerationXout	AccelerationXoutScaled	AccelerationYout	AccelerationYoutScaled	AccelerationZout	AccelerationZoutScaled	Timestamp
func (cc *SupplyChainChaincode) addIotGyroscope(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	gyroscope := Gyroscope{}
	if err := gyroscope.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a gyroscope data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(gyroscope); err == nil {
		Logger.Debug("gyroscope: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &gyroscope, iotGyroscopeIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotGyroscopeIndex
	eventValue.EntityID = gyroscope.Key.ID
	eventValue.Other = gyroscope.Value
	eventValue.Action = eventAddIotGyroscope

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) listIotGyroscope(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)
	gyroscope := []Gyroscope{}
	gyroscopeBytes, err := Query(stub, iotGyroscopeIndex, []string{}, CreateGyroscope, EmptyFilter)
	if err != nil {
		message := fmt.Sprintf("unable to perform method: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}
	if err := json.Unmarshal(gyroscopeBytes, &gyroscope); err != nil {
		message := fmt.Sprintf("unable to unmarshal query result: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	resultBytes, err := json.Marshal(gyroscope)

	Logger.Debug("Result: " + string(resultBytes))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(resultBytes)
}

//0			1			3
//Humidity	Temperature	Timestamp
func (cc *SupplyChainChaincode) addIotHumidity(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	humidity := Humidity{}
	if err := humidity.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a humidity data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(humidity); err == nil {
		Logger.Debug("humidity: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &humidity, iotHumidityIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotHumidityIndex
	eventValue.EntityID = humidity.Key.ID
	eventValue.Other = humidity.Value
	eventValue.Action = eventAddIotHumidity

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) listIotHumidity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)
	humidity := []Humidity{}
	humidityBytes, err := Query(stub, iotHumidityIndex, []string{}, CreateHumidity, EmptyFilter)
	if err != nil {
		message := fmt.Sprintf("unable to perform method: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}
	if err := json.Unmarshal(humidityBytes, &humidity); err != nil {
		message := fmt.Sprintf("unable to unmarshal query result: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	resultBytes, err := json.Marshal(humidity)

	Logger.Debug("Result: " + string(resultBytes))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(resultBytes)
}

//0			1
//Vibration	Timestamp
func (cc *SupplyChainChaincode) addIotVibration(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	vibration := Vibration{}
	if err := vibration.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a vibration data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(vibration); err == nil {
		Logger.Debug("vibration: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &vibration, iotVibrationIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotVibrationIndex
	eventValue.EntityID = vibration.Key.ID
	eventValue.Other = vibration.Value
	eventValue.Action = eventAddIotVibration

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) listIotVibration(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)
	vibration := []Vibration{}
	vibrationBytes, err := Query(stub, iotVibrationIndex, []string{}, CreateVibration, EmptyFilter)
	if err != nil {
		message := fmt.Sprintf("unable to perform method: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}
	if err := json.Unmarshal(vibrationBytes, &vibration); err != nil {
		message := fmt.Sprintf("unable to unmarshal query result: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	resultBytes, err := json.Marshal(vibration)

	Logger.Debug("Result: " + string(resultBytes))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(resultBytes)
}

//0			1
//Light	Timestamp
func (cc *SupplyChainChaincode) addIotLight(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	light := Light{}
	if err := light.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a light data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(light); err == nil {
		Logger.Debug("vibration: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &light, iotLightIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotLightIndex
	eventValue.EntityID = light.Key.ID
	eventValue.Other = light.Value
	eventValue.Action = eventAddIotLight

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

func (cc *SupplyChainChaincode) listIotLight(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)
	light := []Light{}
	lightBytes, err := Query(stub, iotLightIndex, []string{}, CreateLight, EmptyFilter)
	if err != nil {
		message := fmt.Sprintf("unable to perform method: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}
	if err := json.Unmarshal(lightBytes, &light); err != nil {
		message := fmt.Sprintf("unable to unmarshal query result: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	resultBytes, err := json.Marshal(light)

	Logger.Debug("Result: " + string(resultBytes))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(resultBytes)
}

//0
//Certificate
func (cc *SupplyChainChaincode) addIotCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	Notifier(stub, NoticeRuningType)

	//filling from arguments
	certificate := Certificate{}
	if err := certificate.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a certificate data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//updating state in ledger
	if bytes, err := json.Marshal(certificate); err == nil {
		Logger.Debug("certificate: " + string(bytes))
	}

	if err := UpdateOrInsertIn(stub, &certificate, iotCertificateIndex, []string{""}, ""); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	//emitting Event
	events := Events{}

	eventValue := EventValue{}
	eventValue.EntityType = iotCertificateIndex
	eventValue.EntityID = certificate.Key.ID
	eventValue.Other = certificate.Value
	eventValue.Action = eventAddIotCertificate

	events.Values = append(events.Values, eventValue)

	if err := events.EmitEvent(stub); err != nil {
		message := fmt.Sprintf("Cannot emite event: %s", err.Error())
		Logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	Notifier(stub, NoticeSuccessType)
	return shim.Success(nil)
}

//0
//Certificate
func (cc *SupplyChainChaincode) checkIotCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	Notifier(stub, NoticeRuningType)

	//filling from arguments
	certificate := Certificate{}
	if err := certificate.FillFromArguments(stub, args); err != nil {
		message := fmt.Sprintf("cannot fill a certificate data from arguments: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	//check is certificate valid
	valid, err := CheckCertificate(stub, certificate.Value.Certificate)
	if err != nil {
		message := fmt.Sprintf("cannot check the certificate: %s", err.Error())
		Logger.Error(message)
		return shim.Error(message)
	}

	result, err := json.Marshal(valid)
	if err != nil {
		return shim.Error(err.Error())
	}

	Logger.Debug("Result: " + string(result))

	Notifier(stub, NoticeSuccessType)
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(SupplyChainChaincode))
	if err != nil {
		Logger.Error(err.Error())
	}
}
