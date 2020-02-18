package serialwrapper

import (
	"bytes"
	"time"

	"github.com/tarm/serial"
)

const maxRW = 1048576

const (
	serialDeviceFile = "/dev/ttyS0"
	Lockfile         = "/var/lock"
)

// Send comand through serial port
func Send(s *serial.Port, command string) (int, error) {
	var buffer bytes.Buffer
	var number int

	_, err := buffer.WriteString(command)
	if err != nil {
		return number, err
	}

	_, err = buffer.WriteString("\n")
	if err != nil {
		return number, err
	}

	b := []byte(buffer.String())
	number, err = s.Write(b)
	if err != nil {
		return number, err
	}

	time.Sleep(100 * time.Millisecond)

	return number, nil
}

// Read serial output
func Read(s *serial.Port) (string, bool, error) {
	var output string
	var eof bool
	buf := make([]byte, maxRW)

	n, err := s.Read(buf)
	if err != nil {
		return output, eof, err
	}

	if n < maxRW {
		eof = true
	}

	output = string(buf[:n])

	return output, eof, nil
}

// Init serial port
func Init() (*serial.Port, error) {
	c := &serial.Config{Name: serialDeviceFile, Baud: 9600}

	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}

	return s, nil
}
