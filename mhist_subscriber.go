package nervo

import (
	"context"
	"log"

	"github.com/alexmorten/mhist/proto"
	"google.golang.org/grpc"
)

// MhistSubscriber reads new messages from mhist and distributes them to the correct controller
type MhistSubscriber struct {
	manager *Manager
	client  proto.MhistClient
	filter  *proto.Filter
}

// NewMhistSubscriber returns a connected subscriber
func NewMhistSubscriber(address string, filter *proto.Filter, manager *Manager) (*MhistSubscriber, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := proto.NewMhistClient(conn)

	return &MhistSubscriber{
		manager: manager,
		client:  c,
		filter:  filter,
	}, nil
}

// ReadMessages reads the from the subscription and relays messages to the given controllers if possible
func (s *MhistSubscriber) ReadMessages() {
	stream, err := s.client.Subscribe(context.Background(), s.filter)
	if err != nil {
		panic(err)
	}

	for {
		m, err := stream.Recv()
		if err != nil {
			log.Println(err)
			break
		}

		s.handleNewMessage(m)
	}
}

func (s *MhistSubscriber) handleNewMessage(m *proto.MeasurementMessage) {
	c := m.Measurement.GetRaw()
	if c == nil {
		log.Println("ignoring subscribed message, is not raw")
		return
	}

	v := c.Value
	legName, message, ok := ParseGaitAction(string(v))
	if !ok {
		log.Println("is not valid gait action:", v)
		return
	}

	infos := s.manager.listControllers()
	portName := controllerPortForName(legName, infos)
	if portName == "" {
		log.Println("could not find controller with name:", legName)
		return
	}

	err := s.manager.writeToController(portName, []byte(message))
	if err != nil {
		log.Println("writting resulted in error:", err)
	}
}

func controllerPortForName(name string, controllers []controllerInfo) string {

	for _, info := range controllers {
		if info.name == name {
			return info.portName
		}
	}

	return ""
}
