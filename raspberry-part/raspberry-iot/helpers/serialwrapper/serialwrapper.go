package serialwrapper

import (
	"bytes"
	"log"
	"time"

	"github.com/tarm/serial"
)

const maxRW = 1048576

const (
	serialDeviceFile = "/dev/ttyS0"
	Lockfile         = "/var/lock"
)

// Send comand through serial port
func Send(s *serial.Port, command string) int {
	var buffer bytes.Buffer
	buffer.WriteString(command)
	buffer.WriteString("\n")
	b := []byte(buffer.String())
	number, err := s.Write(b)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	return number
}

// Read serial output
func Read(s *serial.Port) (string, bool) {
	eof := false
	buf := make([]byte, maxRW)
	n, err := s.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	if n < maxRW {
		eof = true
	}
	return string(buf[:n]), eof
}

// Init serial port
func Init() *serial.Port {
	c := &serial.Config{Name: serialDeviceFile, Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
