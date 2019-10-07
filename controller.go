package nervo

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/tarm/serial"
)

const (
	maxBufferLength      = 2 << 20
	retainedBufferLength = 2 << 10
)

type controller struct {
	SerialPortPath    string
	Name              string
	serialPort        *serial.Port
	outputbuffer      *bytes.Buffer
	outputMutex       *sync.Mutex
	readNotifierChan  chan []byte
	readNotifierMutex *sync.Mutex
	Error             error
}

func newController(serialPort string) *controller {
	return &controller{
		SerialPortPath:    serialPort,
		outputbuffer:      &bytes.Buffer{},
		outputMutex:       &sync.Mutex{},
		readNotifierMutex: &sync.Mutex{},
	}
}

func (c *controller) flash(hexFileContent []byte) (output string, err error) {
	c.closeSerial()
	c.clearNotifier()
	time.Sleep(time.Millisecond * 200)
	c.Error = nil
	err = withTimeOut(time.Second*10, func() {
		hexFilePath, hexfileCleanup := writeHexFileToTemporaryPath(hexFileContent)
		defer hexfileCleanup()

		cmd := exec.Command(
			"sh",
			"-c",
			fmt.Sprintf("avrdude -p m328p -c arduino -P %s -b 115200 -U flash:w:%s", c.SerialPortPath, hexFilePath),
		)

		out, execErr := cmd.CombinedOutput()
		output = string(out)
		err = execErr
	})
	if err != nil {
		return "", err
	}
	go c.readFromSerial()
	return
}

func (c *controller) readFromSerial() error {
	handleReadErr := func(err error) {
		c.Error = err
		c.clearNotifier()
		c.closeSerial()
		log.Println(c.SerialPortPath, err)
	}

	conf := &serial.Config{Name: c.SerialPortPath, Baud: 9600}
	s, err := serial.OpenPort(conf)
	if err != nil {
		c.Error = err
		return err
	}
	c.serialPort = s
	r := bufio.NewReader(c.serialPort)

	firstLine, err := r.ReadString('\n')
	if err != nil {
		handleReadErr(err)
		return err
	}
	if name, ok := ParseAnnounceMessage(firstLine); ok {
		c.Name = name
	} else {
		c.notifyOrAppendToCappedOutputBuffer([]byte(firstLine))
	}

	for {
		var l string
		l, err = r.ReadString('\n')
		if err != nil {
			handleReadErr(err)
			break
		}
		c.notifyOrAppendToCappedOutputBuffer([]byte(l))
	}
	return err
}

func (c *controller) notifyOrAppendToCappedOutputBuffer(b []byte) {
	c.readNotifierMutex.Lock()
	defer c.readNotifierMutex.Unlock()

	if c.readNotifierChan != nil {
		c.readNotifierChan <- b
		return
	}

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

func (c *controller) notifyOnRead() chan []byte {
	c.clearNotifier()

	c.readNotifierMutex.Lock()
	defer c.readNotifierMutex.Unlock()
	notifierChan := make(chan []byte, 10)
	c.readNotifierChan = notifierChan

	return notifierChan
}

func (c *controller) clearNotifier() {
	c.readNotifierMutex.Lock()
	defer c.readNotifierMutex.Unlock()
	if c.readNotifierChan != nil {
		close(c.readNotifierChan)
		c.readNotifierChan = nil
	}
}

func (c *controller) closeSerial() {
	c.outputMutex.Lock()
	defer c.outputMutex.Unlock()
	if c.serialPort != nil {
		c.serialPort.Close()
		c.serialPort = nil
	}
}

func writeHexFileToTemporaryPath(hexFileContent []byte) (path string, cleanup func()) {
	tmpfile, err := ioutil.TempFile("", "flashing_*.hex")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write(hexFileContent); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
	cleanupFunc := func() {
		os.Remove(tmpfile.Name())
	}
	return tmpfile.Name(), cleanupFunc
}
