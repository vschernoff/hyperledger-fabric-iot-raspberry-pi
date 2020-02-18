package barometer

import (
	"fmt"
	"github.com/d2r2/go-i2c"
	"hlf-iot/config"
	"hlf-iot/devices/led"
	"hlf-iot/helpers/queuewrapper"
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
	Led         *led.Led
}

func Init(ledBadData *led.Led) (*Barometer, error) {
	i2cData, err := i2c.NewI2C(address, bus)
	if err != nil {
		return nil, err
	}
	barometer := &Barometer{}
	barometer.i2c = i2cData
	barometer.Led = ledBadData

	return barometer, nil
}

func (barometer *Barometer) Activate() error {
	var err error

	barometer.mode = BMP180_STANDARD

	calAC1, err := barometer.readS16(BMP180_CAL_AC1)
	if err != nil {
		return err
	}
	barometer.calAC1 = int64(calAC1)

	calAC2, err := barometer.readS16(BMP180_CAL_AC2)
	if err != nil {
		return err
	}
	barometer.calAC2 = int64(calAC2)

	calAC3, err := barometer.readS16(BMP180_CAL_AC3)
	if err != nil {
		return err
	}
	barometer.calAC3 = int64(calAC3)

	calAC4, err := barometer.readU16(BMP180_CAL_AC4)
	if err != nil {
		return err
	}
	barometer.calAC4 = int64(calAC4)

	calAC5, err := barometer.readU16(BMP180_CAL_AC5)
	if err != nil {
		return err
	}
	barometer.calAC5 = int64(calAC5)

	calAC6, err := barometer.readU16(BMP180_CAL_AC6)
	if err != nil {
		return err
	}
	barometer.calAC6 = int64(calAC6)

	calB1, err := barometer.readS16(BMP180_CAL_B1)
	if err != nil {
		return err
	}
	barometer.calB1 = int64(calB1)

	calB2, err := barometer.readS16(BMP180_CAL_B2)
	if err != nil {
		return err
	}
	barometer.calB2 = int64(calB2)

	calMB, err := barometer.readS16(BMP180_CAL_MB)
	if err != nil {
		return err
	}
	barometer.calMB = int64(calMB)

	calMC, err := barometer.readS16(BMP180_CAL_MC)
	if err != nil {
		return err
	}
	barometer.calMC = int64(calMC)

	calMD, err := barometer.readS16(BMP180_CAL_MD)
	if err != nil {
		return err
	}
	barometer.calMD = int64(calMD)

	return nil
}

func (barometer *Barometer) ReadPressure() (float64, error) {
	var p float64

	UT, err := barometer.ReadRawTemperature()
	if err != nil {
		return p, err
	}

	UP, err := barometer.ReadRawPressure()
	if err != nil {
		return p, err
	}

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

	if B7 < 0x80000000 {
		p = float64(B7*2) / float64(B4)
	} else {
		p = float64(B7/B4) * 2
	}

	X1 = (int64(p) >> 8) * (int64(p) >> 8)
	X1 = (X1 * 3038) >> 16
	X2 = int64(-7357*p) >> 16

	p = p + float64((X1+X2+3791)>>4)

	return p, nil
}

func (barometer *Barometer) ReadTemperature() (float64, error) {
	var temp float64

	UT, err := barometer.ReadRawTemperature()
	if err != nil {
		return temp, err
	}

	X1 := ((UT - barometer.calAC6) * barometer.calAC5) >> 15
	X2 := (barometer.calMC << 11) / (X1 + barometer.calMD)
	B5 := X1 + X2
	temp = float64((B5+8)>>4) / 10.0

	return temp, nil
}

func (barometer *Barometer) ReadAltitude() (float64, error) {

	pressure, err := barometer.ReadPressure()
	if err != nil {
		return pressure, err
	}
	pressure = float64(pressure)

	altitude := 44330.0 * (1.0 - math.Pow(pressure/SEALEVEL_PA, (1.0/5.255)))

	return altitude, nil
}

func (barometer *Barometer) GetDataInJsonString() (*queuewrapper.SendData, error) {
	var err error

	now := time.Now()
	barometer.Timestamp = int64(now.Unix())

	barometer.Temperature, err = barometer.ReadTemperature()
	if err != nil {
		return nil, err
	}

	barometer.Pressure, err = barometer.ReadPressure()
	if err != nil {
		return nil, err
	}
	barometer.Pressure = barometer.Pressure / 100

	barometer.Altitude, err = barometer.ReadAltitude()
	if err != nil {
		return nil, err
	}

	sendData := &queuewrapper.SendData{}
	sendData.Fcn = config.FCN_NAME_BAROMETER
	sendData.Args = []string{fmt.Sprintf("%f", barometer.Pressure), fmt.Sprintf("%f", barometer.Altitude), fmt.Sprintf("%f", barometer.Temperature), fmt.Sprintf("%d", barometer.Timestamp)}
	if barometer.Pressure*barometer.Altitude*barometer.Temperature == 0 {
		barometer.Led.SetOn()
	}

	return sendData, nil
}

func (barometer *Barometer) GetQueueElement() *queuewrapper.QueueStructure {
	return &queuewrapper.QueueStructure{GetDataFcn: barometer.GetDataInJsonString}
}

func (barometer *Barometer) readS16(reg byte) (int16, error) {
	data, err := barometer.i2c.ReadRegS16BE(reg)

	return data, err
}

func (barometer *Barometer) readU16(reg byte) (uint16, error) {
	data, err := barometer.i2c.ReadRegU16BE(reg)

	return data, err
}

func (barometer *Barometer) ReadByte(reg byte) ([]byte, error) {
	data, _, err := barometer.i2c.ReadRegBytes(reg, 1)

	return data, err
}

func (barometer *Barometer) WriteByte(reg byte, cmd byte) error {
	err := barometer.i2c.WriteRegU8(reg, cmd)

	return err
}

func (barometer *Barometer) ReadRawPressure() (int64, error) {
	var raw int64

	err := barometer.WriteByte(BMP180_CONTROL, byte(BMP180_READPRESSURECMD+(barometer.mode<<6)))
	if err != nil {
		return raw, err
	}

	if barometer.mode == BMP180_ULTRALOWPOWER {
		time.Sleep(5 * time.Millisecond)
	} else if barometer.mode == BMP180_HIGHRES {
		time.Sleep(14 * time.Millisecond)
	} else if barometer.mode == BMP180_ULTRAHIGHRES {
		time.Sleep(26 * time.Millisecond)
	} else {
		time.Sleep(8 * time.Millisecond)
	}

	MSB, err := barometer.ReadByte(BMP180_PRESSUREDATA)
	if err != nil {
		return raw, err
	}

	LSB, err := barometer.ReadByte(BMP180_PRESSUREDATA + 1)
	if err != nil {
		return raw, err
	}

	XLSB, err := barometer.ReadByte(BMP180_PRESSUREDATA + 2)
	if err != nil {
		return raw, err
	}

	MSBInt := int64(MSB[0])
	LSBInt := int64(LSB[0])
	XLSBInt := int64(XLSB[0])

	raw = ((MSBInt << 16) + (LSBInt << 8) + XLSBInt) >> (8 - barometer.mode)

	return raw, nil
}

func (barometer *Barometer) ReadRawTemperature() (int64, error) {
	var raw int64

	err := barometer.WriteByte(BMP180_CONTROL, byte(BMP180_READTEMPCMD))
	if err != nil {
		return raw, err
	}
	time.Sleep(5 * time.Millisecond)

	MSB, err := barometer.ReadByte(BMP180_TEMPDATA)
	if err != nil {
		return raw, err
	}

	LSB, err := barometer.ReadByte(BMP180_TEMPDATA + 1)
	if err != nil {
		return raw, err
	}

	raw = (int64(MSB[0]) << 8) + int64(LSB[0])

	return raw, nil
}
