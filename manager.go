package nervo

import (
	"bytes"
	"errors"
	"log"
	"time"
)

type controllerInfo struct {
	name     string
	portName string
}

type readOutputMessage struct {
	portName   string
	answerChan chan string
}

type flashAnswer struct {
	Error  error
	Output string
}

type flashMessage struct {
	portName       string
	hexFileContent []byte
	answerChan     chan flashAnswer
}

type readContinuousMessage struct {
	portName   string
	answerChan chan chan []byte
}

type stopReadingMessage struct {
	portName string
}

type nameControllerMessage struct {
	portName string
	name     string
}

type pingMessage struct {
	pongChan chan struct{}
}

type writeToControllerMessage struct {
	portName string
	message  []byte
	doneChan chan error
}

// Manager controls all interactions with the controllers from outside
type Manager struct {
	controllers           []*controller
	currentPortsChan      chan []string
	readOutputChan        chan readOutputMessage
	flashChan             chan flashMessage
	readContinuousChan    chan readContinuousMessage
	stopReadingChan       chan stopReadingMessage
	nameControllerChan    chan nameControllerMessage
	pingChan              chan pingMessage
	writeToControllerChan chan writeToControllerMessage
}

// NewManager retuns a Manager that is ready for use
func NewManager() *Manager {
	m := &Manager{
		currentPortsChan:      make(chan []string),
		readOutputChan:        make(chan readOutputMessage),
		flashChan:             make(chan flashMessage),
		readContinuousChan:    make(chan readContinuousMessage),
		stopReadingChan:       make(chan stopReadingMessage),
		nameControllerChan:    make(chan nameControllerMessage),
		pingChan:              make(chan pingMessage),
		writeToControllerChan: make(chan writeToControllerMessage),
	}

	go m.lookForNewPorts()
	go m.manageControllers()
	go watchManagerHealth(m, func() {
		panic("I don't know, just kill him I guess")
	})
	return m
}

func (m *Manager) manageControllers() {
	for {
		select {
		case currentPorts := <-m.currentPortsChan:
			m.handleCurrentPorts(currentPorts)
			break
		case message := <-m.readOutputChan:
			controller := m.controllerForPort(message.portName)
			if controller != nil {
				controller.useOutput(func(buffer *bytes.Buffer) {
					outputBuffer := make([]byte, buffer.Len())
					_, err := buffer.Read(outputBuffer)
					if err != nil {
						panic(err)
					}
					message.answerChan <- string(outputBuffer)
				})
			} else {
				message.answerChan <- "no controller found at " + message.portName
			}
			break
		case message := <-m.flashChan:
			controller := m.controllerForPort(message.portName)
			if controller != nil {
				output, err := controller.flash(message.hexFileContent)
				message.answerChan <- flashAnswer{Error: err, Output: output}
			} else {
				message.answerChan <- flashAnswer{Error: errors.New("no controller found at " + message.portName)}
			}
			break
		case message := <-m.readContinuousChan:
			controller := m.controllerForPort(message.portName)
			if controller != nil {
				notifierChan := controller.notifyOnRead()
				message.answerChan <- notifierChan
			} else {
				message.answerChan <- nil
			}
			break
		case message := <-m.stopReadingChan:
			controller := m.controllerForPort(message.portName)
			if controller != nil {
				controller.clearNotifier()
			}
			break
		case message := <-m.nameControllerChan:
			controller := m.controllerForPort(message.portName)
			if controller != nil {
				controller.Name = message.name
			}
			break
		case message := <-m.writeToControllerChan:
			controller := m.controllerForPort(message.portName)
			if controller != nil {
				message.doneChan <- controller.write(message.message)
			} else {
				message.doneChan <- errors.New("no controller found at " + message.portName)
			}
			break
		case m := <-m.pingChan:
			m.pongChan <- struct{}{}
			break
		}
	}
}

func (m *Manager) listControllers() []controllerInfo {
	infos := []controllerInfo{}
	for _, controller := range m.controllers {
		infos = append(infos, controllerInfo{portName: controller.SerialPortPath, name: controller.Name})
	}
	return infos
}

func (m *Manager) readFromController(portName string) string {
	answerChan := make(chan string)
	message := readOutputMessage{answerChan: answerChan, portName: portName}
	m.readOutputChan <- message
	return <-answerChan
}

func (m *Manager) flashController(portName string, hexFileContent []byte) flashAnswer {
	answerChan := make(chan flashAnswer)
	message := flashMessage{answerChan: answerChan, portName: portName, hexFileContent: hexFileContent}
	m.flashChan <- message
	return <-answerChan
}

func (m *Manager) readContinuouslyFromController(portName string) chan []byte {
	answerChan := make(chan chan []byte)
	message := readContinuousMessage{answerChan: answerChan, portName: portName}
	m.readContinuousChan <- message
	return <-answerChan
}

func (m *Manager) stopReadingFromController(portName string) {
	message := stopReadingMessage{portName: portName}
	m.stopReadingChan <- message
}

func (m *Manager) setControllerName(portName string, name string) {
	message := nameControllerMessage{portName: portName, name: name}
	m.nameControllerChan <- message
}

func (m *Manager) controllerForPort(portName string) *controller {
	for _, controller := range m.controllers {
		if controller.SerialPortPath == portName {
			return controller
		}
	}

	return nil
}

func (m *Manager) lookForNewPorts() {
	t := time.NewTicker(time.Second)
	for {
		<-t.C
		ports, err := discoverAttachedControllers()
		if err != nil {
			panic(err)
		}

		m.currentPortsChan <- ports
	}
}

func (m *Manager) pingWithTimeout(timeout time.Duration) error {
	return withTimeOut(timeout, func() {
		pongChan := make(chan struct{})
		m.pingChan <- pingMessage{
			pongChan: pongChan,
		}
		<-pongChan
	})
}

func (m *Manager) writeToController(controllerPortName string, message []byte) error {
	doneChan := make(chan error)
	m.writeToControllerChan <- writeToControllerMessage{
		portName: controllerPortName,
		message:  message,
		doneChan: doneChan,
	}
	return <-doneChan
}

func (m *Manager) handleCurrentPorts(currentPorts []string) {
	newPorts := []string{}
	for _, port := range currentPorts {
		isIncluded := false
		for _, controller := range m.controllers {
			if port == controller.SerialPortPath {
				isIncluded = true
			}
		}

		if !isIncluded {
			newPorts = append(newPorts, port)
		}
	}

	for _, newPort := range newPorts {
		log.Println("discovered new port: ", newPort)
		controller := newController(newPort)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}
			}()
			controller.readFromSerial()
		}()
		m.controllers = append(m.controllers, controller)
	}

	removedControllers := []*controller{}
	for _, controller := range m.controllers {
		isIncluded := false
		for _, port := range currentPorts {
			if controller.SerialPortPath == port {
				isIncluded = true
			}
		}

		if !isIncluded {
			removedControllers = append(removedControllers, controller)
		}
	}

	for _, removed := range removedControllers {
		removed.closeSerial()
	}

	currentControllers := []*controller{}
	for _, controller := range m.controllers {
		isRemoved := false
		for _, removed := range removedControllers {
			if controller == removed {
				isRemoved = true
			}
		}

		if !isRemoved {
			currentControllers = append(currentControllers, controller)
		}
	}
	m.controllers = currentControllers
}
