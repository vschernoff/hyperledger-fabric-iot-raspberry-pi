package barometer

import (
	"fmt"
	"github.com/d2r2/go-i2c"
	"hlf-iot/config"
	"math"
	"time"
)

// Operating Modes
const (
	BMP180_ULTRALOWPOWER = iota
	BMP180_STANDARD
	BMP180_HIGHRES
	BMP180_ULTRAHIGHRES
)

const (
	address = 0x77 // BMP default address
	bus     = 1
	// BMP085 Registers
	BMP180_CAL_AC1      = 0xAA // Calibration data (16 bits)
	BMP180_CAL_AC2      = 0xAC // Calibration data (16 bits)
	BMP180_CAL_AC3      = 0xAE // Calibration data (16 bits)
	BMP180_CAL_AC4      = 0xB0 // Calibration data (16 bits)
	BMP180_CAL_AC5      = 0xB2 // Calibration data (16 bits)
	BMP180_CAL_AC6      = 0xB4 // Calibration data (16 bits)
	BMP180_CAL_B1       = 0xB6 // Calibration data (16 bits)
	BMP180_CAL_B2       = 0xB8 // Calibration data (16 bits)
	BMP180_CAL_MB       = 0xBA // Calibration data (16 bits)
	BMP180_CAL_MC       = 0xBC // Calibration data (16 bits)
	BMP180_CAL_MD       = 0xBE // Calibration data (16 bits)
	BMP180_CONTROL      = 0xF4
	BMP180_TEMPDATA     = 0xF6
	BMP180_PRESSUREDATA = 0xF6
	// Commands
	BMP180_READTEMPCMD     = 0x2E
	BMP180_READPRESSURECMD = 0x34
	// Sea level in pascal
	SEALEVEL_PA = 101325.0
)

type Barometer struct {
	i2c         *i2c.I2C
	mode        uint
	calAC1      int64
	calAC2      int64
	calAC3      int64
	calAC4      int64
	calAC5      int64
	calAC6      int64
	calB1       int64
	calB2       int64
	calMB       int64
	calMC       int64
	calMD       int64
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Altitude    float64 `json:"altitude"`
	Timestamp   int64   `json:"timestamp"`
}

func Init() *Barometer {
	i2cData, err := i2c.NewI2C(address, bus)
	if err != nil {
		fmt.Println(err)
	}
	barometer := &Barometer{}
	barometer.i2c = i2cData
	return barometer
}

func (barometer *Barometer) Activate() {
	barometer.mode = BMP180_STANDARD
	barometer.calAC1 = int64(barometer.readS16(BMP180_CAL_AC1))
	barometer.calAC2 = int64(barometer.readS16(BMP180_CAL_AC2))
	barometer.calAC3 = int64(barometer.readS16(BMP180_CAL_AC3))
	barometer.calAC4 = int64(barometer.readU16(BMP180_CAL_AC4))
	barometer.calAC5 = int64(barometer.readU16(BMP180_CAL_AC5))
	barometer.calAC6 = int64(barometer.readU16(BMP180_CAL_AC6))
	barometer.calB1 = int64(barometer.readS16(BMP180_CAL_B1))
	barometer.calB2 = int64(barometer.readS16(BMP180_CAL_B2))
	barometer.calMB = int64(barometer.readS16(BMP180_CAL_MB))
	barometer.calMC = int64(barometer.readS16(BMP180_CAL_MC))
	barometer.calMD = int64(barometer.readS16(BMP180_CAL_MD))
}

func (barometer *Barometer) ReadPressure() float64 {
	UT := barometer.ReadRawTemperature()
	UP := barometer.ReadRawPressure()

	X1 := ((UT - barometer.calAC6) * barometer.calAC5) >> 15
	X2 := (barometer.calMC << 11) / (X1 + barometer.calMD)
	B5 := X1 + X2

	// Pressure Calculations
	B6 := B5 - 4000
	X1 = (barometer.calB2 * (B6 * B6) >> 12) >> 11
	X2 = (barometer.calAC2 * B6) >> 11
	X3 := X1 + X2
	B3 := (((barometer.calAC1*4 + X3) << barometer.mode) + 2) / 4

	X1 = (barometer.calAC3 * B6) >> 13
	X2 = (barometer.calB1 * ((B6 * B6) >> 12)) >> 16
	X3 = ((X1 + X2) + 2) >> 2
	B4 := (barometer.calAC4 * (X3 + 32768)) >> 15
	B7 := (UP - B3) * (50000 >> barometer.mode)

	var p float64
	if B7 < 0x80000000 {
		p = float64(B7*2) / float64(B4)
	} else {
		p = float64(B7/B4) * 2
	}

	X1 = (int64(p) >> 8) * (int64(p) >> 8)
	X1 = (X1 * 3038) >> 16
	X2 = int64(-7357*p) >> 16

	p = p + float64((X1+X2+3791)>>4)

	return p
}

func (barometer *Barometer) ReadTemperature() float64 {
	UT := barometer.ReadRawTemperature()

	X1 := ((UT - barometer.calAC6) * barometer.calAC5) >> 15
	X2 := (barometer.calMC << 11) / (X1 + barometer.calMD)
	B5 := X1 + X2
	temp := float64((B5+8)>>4) / 10.0

	return temp
}

func (barometer *Barometer) ReadAltitude() float64 {
	pressure := float64(barometer.ReadPressure())
	altitude := 44330.0 * (1.0 - math.Pow(pressure/SEALEVEL_PA, (1.0/5.255)))

	return altitude
}

func (barometer *Barometer) GetDataInJsonString() *config.SendData {
	now := time.Now()
	barometer.Timestamp = int64(now.Unix())
	barometer.Temperature = barometer.ReadTemperature()
	barometer.Pressure = barometer.ReadPressure() / 100
	barometer.Altitude = barometer.ReadAltitude()
	sendData := &config.SendData{}
	sendData.Fcn = config.FCN_NAME_BAROMETER
	sendData.Args = []string{fmt.Sprintf("%f", barometer.Pressure), fmt.Sprintf("%f", barometer.Altitude), fmt.Sprintf("%f", barometer.Temperature), fmt.Sprintf("%d", barometer.Timestamp)}

	return sendData
}

func (barometer *Barometer) GetQueueElement() *config.QueueStructure {
	queueStructure := &config.QueueStructure{}
	queueStructure.GetDataFcn = nil
	queueStructure.CheckResponseFcn = config.CheckInsertToBC
	queueStructure.JsonString = barometer.GetDataInJsonString()
	queueStructure.HttpAction = config.GPRS_HTTPACTION_POST

	return queueStructure
}

func (barometer *Barometer) readS16(reg byte) int16 {
	data, err := barometer.i2c.ReadRegS16BE(reg)
	if err != nil {
		fmt.Println(err)
	}

	return data
}

func (barometer *Barometer) readU16(reg byte) uint16 {
	data, err := barometer.i2c.ReadRegU16BE(reg)
	if err != nil {
		fmt.Println(err)
	}

	return data
}

func (barometer *Barometer) ReadByte(reg byte) []byte {
	data, _, err := barometer.i2c.ReadRegBytes(reg, 1)
	if err != nil {
		fmt.Println(err)
	}

	return data
}

func (barometer *Barometer) WriteByte(reg byte, cmd byte) {
	err := barometer.i2c.WriteRegU8(reg, cmd)
	if err != nil {
		fmt.Println(err)
	}
}

func (barometer *Barometer) ReadRawPressure() int64 {
	barometer.WriteByte(BMP180_CONTROL, byte(BMP180_READPRESSURECMD+(barometer.mode<<6)))

	if barometer.mode == BMP180_ULTRALOWPOWER {
		time.Sleep(5 * time.Millisecond)
	} else if barometer.mode == BMP180_HIGHRES {
		time.Sleep(14 * time.Millisecond)
	} else if barometer.mode == BMP180_ULTRAHIGHRES {
		time.Sleep(26 * time.Millisecond)
	} else {
		time.Sleep(8 * time.Millisecond)
	}

	MSB := barometer.ReadByte(BMP180_PRESSUREDATA)
	LSB := barometer.ReadByte(BMP180_PRESSUREDATA + 1)
	XLSB := barometer.ReadByte(BMP180_PRESSUREDATA + 2)

	MSBInt := int64(MSB[0])
	LSBInt := int64(LSB[0])
	XLSBInt := int64(XLSB[0])

	raw := ((MSBInt << 16) + (LSBInt << 8) + XLSBInt) >> (8 - barometer.mode)

	return raw
}

func (barometer *Barometer) ReadRawTemperature() int64 {
	barometer.WriteByte(BMP180_CONTROL, byte(BMP180_READTEMPCMD))
	time.Sleep(5 * time.Millisecond)

	MSB := barometer.ReadByte(BMP180_TEMPDATA)
	LSB := barometer.ReadByte(BMP180_TEMPDATA + 1)

	raw := (int64(MSB[0]) << 8) + int64(LSB[0])

	return raw
}
