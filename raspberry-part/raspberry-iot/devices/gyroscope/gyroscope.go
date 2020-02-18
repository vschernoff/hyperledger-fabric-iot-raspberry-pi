package gyroscope

import (
	"fmt"
	"github.com/d2r2/go-i2c"
	"hlf-iot/config"
	"time"
)

const (
	power_mgmt_1               = 0x6b
	address                    = 0x68
	bus                        = 1
	gyroscope_xout_address     = 0x43
	gyroscope_yout_address     = 0x45
	gyroscope_zout_address     = 0x47
	accelerometer_xout_address = 0x3b
	accelerometer_yout_address = 0x3d
	accelerometer_zout_address = 0x3f
	out_address_bytes_count    = 2
)

type Gyroscope struct {
	i2c                    *i2c.I2C
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
	Timestamp              int64   `json:"timestamp"`
}

func Init() *Gyroscope {
	i2cData, err := i2c.NewI2C(address, bus)
	if err != nil {
		fmt.Println(err)
	}
	gyroscope := &Gyroscope{}
	gyroscope.i2c = i2cData
	return gyroscope
}

func (gyroscope *Gyroscope) Activate() {
	err := gyroscope.i2c.WriteRegU8(power_mgmt_1, 0)
	time.Sleep(5 * time.Millisecond)
	if err != nil {
		fmt.Println(err)
	}
}

func (gyroscope *Gyroscope) GetDataInJsonString() *config.SendData {
	gyroscope.GetData()
	now := time.Now()
	gyroscope.Timestamp = int64(now.Unix())
	sendData := &config.SendData{}
	sendData.Fcn = config.FCN_NAME_GYROSCOPE
	sendData.Args = []string{fmt.Sprintf("%f", gyroscope.Xout), fmt.Sprintf("%f", gyroscope.XoutScaled),
		fmt.Sprintf("%f", gyroscope.Yout), fmt.Sprintf("%f", gyroscope.YoutScaled),
		fmt.Sprintf("%f", gyroscope.Zout), fmt.Sprintf("%f", gyroscope.ZoutScaled),
		fmt.Sprintf("%f", gyroscope.AccelerationXout), fmt.Sprintf("%f", gyroscope.AccelerationXoutScaled),
		fmt.Sprintf("%f", gyroscope.AccelerationYout), fmt.Sprintf("%f", gyroscope.AccelerationYoutScaled),
		fmt.Sprintf("%f", gyroscope.AccelerationZout), fmt.Sprintf("%f", gyroscope.AccelerationZoutScaled),
		fmt.Sprintf("%d", gyroscope.Timestamp)}

	return sendData
}

func (gyroscope *Gyroscope) GetQueueElement() *config.QueueStructure {
	queueStructure := &config.QueueStructure{}
	queueStructure.GetDataFcn = nil
	queueStructure.CheckResponseFcn = config.CheckInsertToBC
	queueStructure.JsonString = gyroscope.GetDataInJsonString()
	queueStructure.HttpAction = config.GPRS_HTTPACTION_POST

	return queueStructure
}

func (gyroscope *Gyroscope) GetData() {
	dataX, _, err := gyroscope.i2c.ReadRegBytes(gyroscope_xout_address, out_address_bytes_count)
	if err != nil {
		fmt.Println(err)
	}
	value := (int(dataX[0]) << 8) + int(dataX[1])
	if value >= 0x8000 {
		value = -((65535 - value) + 1)
	}
	gyroscope.Xout = float32(value)
	gyroscope.XoutScaled = float32(value) / 131

	dataY, _, err := gyroscope.i2c.ReadRegBytes(gyroscope_yout_address, out_address_bytes_count)
	if err != nil {
		fmt.Println(err)
	}
	value = (int(dataY[0]) << 8) + int(dataY[1])
	if value >= 0x8000 {
		value = -((65535 - value) + 1)
	}
	gyroscope.Yout = float32(value)
	gyroscope.YoutScaled = float32(value) / 131

	dataZ, _, err := gyroscope.i2c.ReadRegBytes(gyroscope_zout_address, out_address_bytes_count)
	if err != nil {
		fmt.Println(err)
	}
	value = (int(dataZ[0]) << 8) + int(dataZ[1])
	if value >= 0x8000 {
		value = -((65535 - value) + 1)
	}
	gyroscope.Zout = float32(value)
	gyroscope.ZoutScaled = float32(value) / 131

	accelerometerDataX, _, err := gyroscope.i2c.ReadRegBytes(accelerometer_xout_address, out_address_bytes_count)
	if err != nil {
		fmt.Println(err)
	}
	value = (int(accelerometerDataX[0]) << 8) + int(accelerometerDataX[1])
	if value >= 0x8000 {
		value = -((65535 - value) + 1)
	}
	gyroscope.AccelerationXout = float32(value)
	gyroscope.AccelerationXoutScaled = float32(value) / 16384.0

	accelerometerDataY, _, err := gyroscope.i2c.ReadRegBytes(accelerometer_yout_address, out_address_bytes_count)
	if err != nil {
		fmt.Println(err)
	}
	value = (int(accelerometerDataY[0]) << 8) + int(accelerometerDataY[1])
	if value >= 0x8000 {
		value = -((65535 - value) + 1)
	}
	gyroscope.AccelerationYout = float32(value)
	gyroscope.AccelerationYoutScaled = float32(value) / 16384.0

	accelerometerDataZ, _, err := gyroscope.i2c.ReadRegBytes(accelerometer_zout_address, out_address_bytes_count)
	if err != nil {
		fmt.Println(err)
	}
	value = (int(accelerometerDataZ[0]) << 8) + int(accelerometerDataZ[1])
	if value >= 0x8000 {
		value = -((65535 - value) + 1)
	}
	gyroscope.AccelerationZout = float32(value)
	gyroscope.AccelerationZoutScaled = float32(value) / 16384.0
}
