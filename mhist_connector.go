package nervo

import (
	"context"
	"log"

	"github.com/alexmorten/mhist/models"
	"github.com/alexmorten/mhist/proto"
	"google.golang.org/grpc"
)

// MhistConnector reads new messages from mhist and distributes them to the correct controller
type MhistConnector struct {
	manager     *Manager
	client      proto.MhistClient
	filter      *proto.Filter
	writeStream proto.Mhist_StoreStreamClient
}

// NewMhistConnector returns a connected subscriber
func NewMhistConnector(address string, filter *proto.Filter, manager *Manager) (*MhistConnector, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := proto.NewMhistClient(conn)
	stream, err := c.StoreStream(context.Background())
	if err != nil {
		return nil, err
	}
	return &MhistConnector{
		manager:     manager,
		client:      c,
		filter:      filter,
		writeStream: stream,
	}, nil
}

// WriteMessage to mhist
func (c *MhistConnector) WriteMessage(verb string, message string) {
	err := c.writeStream.Send(&proto.MeasurementMessage{
		Name: verb,
		Measurement: proto.MeasurementFromModel(&models.Raw{
			Value: []byte(message),
		})})
	if err != nil {
		log.Println(err)
	}
}

// ReadMessages reads the from the subscription and relays messages to the given controllers if possible
func (c *MhistConnector) ReadMessages() {
	stream, err := c.client.Subscribe(context.Background(), c.filter)
	if err != nil {
		panic(err)
	}

	for {
		m, err := stream.Recv()
		if err != nil {
			log.Println(err)
			break
		}

		c.handleNewMessage(m)
	}
}

func (c *MhistConnector) handleNewMessage(m *proto.MeasurementMessage) {
	r := m.Measurement.GetRaw()
	if r == nil {
		log.Println("ignoring subscribed message, is not raw")
		return
	}

	v := r.Value
	legName, message, ok := ParseGaitAction(string(v))
	if !ok {
		log.Println("is not valid gait action:", v)
		return
	}

	infos := c.manager.listControllers()
	portName := controllerPortForName(legName, infos)
	if portName == "" {
		log.Println("could not find controller with name:", legName)
		return
	}

	err := c.manager.writeToController(portName, []byte(message))
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
