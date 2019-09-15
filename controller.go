package nervo

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/tarm/serial"
)

const (
	maxBufferLength      = 2 << 20
	retainedBufferLength = 2 << 10
)

type controller struct {
	SerialPortPath string
	serialPort     *serial.Port
	outputbuffer   *bytes.Buffer
	outputMutex    *sync.Mutex
	Error          error
}

func newController(serialPort string) *controller {
	return &controller{
		SerialPortPath: serialPort,
		outputbuffer:   &bytes.Buffer{},
		outputMutex:    &sync.Mutex{},
	}
}

func (c *controller) flash(hexFilePath string) error {
	c.closeSerial()

	cmd := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf("avrdude -p m328p -c arduino -P %s -b 115200 -U flash:w:%s", c.SerialPortPath, hexFilePath),
	)

	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	return err
}

func (c *controller) readFromSerial() error {
	conf := &serial.Config{Name: c.SerialPortPath, Baud: 9600}
	s, err := serial.OpenPort(conf)
	if err != nil {
		c.Error = err
		return err
	}
	c.serialPort = s
	r := bufio.NewReader(c.serialPort)

	for {
		var l string
		l, err = r.ReadString('\n')
		if err != nil {
			c.Error = err
			log.Println(c.SerialPortPath, err)
			break
		}
		c.appendToCappedOutputBuffer([]byte(l))
	}
	return err
}

func (c *controller) appendToCappedOutputBuffer(b []byte) {
	c.outputMutex.Lock()
	defer c.outputMutex.Unlock()

	if c.outputbuffer.Len()+len(b) > maxBufferLength {
		toRetain := c.outputbuffer.Bytes()[:retainedBufferLength]
		retained := make([]byte, retainedBufferLength)
		copy(retained, toRetain)
		c.outputbuffer = bytes.NewBuffer(retained)
	}

	c.outputbuffer.Write(b)
}

func (c *controller) useOutput(f func(outputBuffer *bytes.Buffer)) {
	c.outputMutex.Lock()
	defer c.outputMutex.Unlock()

	f(c.outputbuffer)
}

func (c *controller) closeSerial() {
	if c.serialPort != nil {
		c.serialPort.Close()
		c.serialPort = nil
	}
}
