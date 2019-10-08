package nervo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/codeuniversity/nervo/proto"

	"google.golang.org/grpc"
)

// GrpcServer translates grpc requests into calls to the manager
type GrpcServer struct {
	Manager  *Manager
	grpcPort int
}

// NewGrpcServer creates a GrpcServer for the given manager
func NewGrpcServer(m *Manager, grpcPort int) *GrpcServer {
	return &GrpcServer{
		Manager:  m,
		grpcPort: grpcPort,
	}
}

// Listen blocks, while listening for grpc requests on the port specified in the GrpcServer struct
func (s *GrpcServer) Listen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", s.grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterNervoServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

// ListControllers for the grpc NervoService
func (s *GrpcServer) ListControllers(_ context.Context, _ *proto.ControllerListRequest) (*proto.ControllerListResponse, error) {
	controllerInfos := s.Manager.listControllers()
	infos := []*proto.ControllerInfo{}
	for _, info := range controllerInfos {
		infos = append(infos, &proto.ControllerInfo{
			PortName: info.portName,
			Name:     info.name,
		})
	}

	return &proto.ControllerListResponse{ControllerInfos: infos}, nil
}

// ReadControllerOutput for the grpc NervoService
func (s *GrpcServer) ReadControllerOutput(_ context.Context, request *proto.ReadControllerOutputRequest) (*proto.ReadControllerOutputResponse, error) {
	output := s.Manager.readFromController(request.ControllerPortName)

	return &proto.ReadControllerOutputResponse{Output: output}, nil
}

// FlashController for the grpc NervoService
func (s *GrpcServer) FlashController(_ context.Context, request *proto.FlashControllerRequest) (*proto.FlashControllerResponse, error) {
	answer := s.Manager.flashController(request.ControllerPortName, request.HexFileContent)
	return &proto.FlashControllerResponse{Output: answer.Output}, answer.Error
}

// ReadControllerOutputContinuously for the grpc NervoService
func (s *GrpcServer) ReadControllerOutputContinuously(request *proto.ReadControllerOutputRequest, stream proto.NervoService_ReadControllerOutputContinuouslyServer) error {

	notifierChan := s.Manager.readContinuouslyFromController(request.ControllerPortName)

	if notifierChan == nil {
		return errors.New("no controller found for " + request.ControllerPortName)
	}

	output := s.Manager.readFromController(request.ControllerPortName)
	if len(output) > 0 {
		err := stream.Send(&proto.ReadControllerOutputResponse{Output: output})
		if err != nil {
			fmt.Println(err)
			s.Manager.stopReadingFromController(request.ControllerPortName)
			return err
		}
	}

	for newOutput := range notifierChan {
		if len(newOutput) > 0 {
			err := stream.Send(&proto.ReadControllerOutputResponse{Output: string(newOutput)})
			if err != nil {
				fmt.Println(err)
				s.Manager.stopReadingFromController(request.ControllerPortName)
				return err
			}
		}
	}
	return nil
}

// SetControllerName for the grpc NervoService
func (s *GrpcServer) SetControllerName(_ context.Context, request *proto.ControllerInfo) (*proto.ControllerListResponse, error) {
	s.Manager.setControllerName(request.PortName, request.Name)

	controllerInfos := s.Manager.listControllers()
	infos := []*proto.ControllerInfo{}
	for _, info := range controllerInfos {
		infos = append(infos, &proto.ControllerInfo{
			PortName: info.portName,
			Name:     info.name,
		})
	}

	return &proto.ControllerListResponse{ControllerInfos: infos}, nil
}

// ResetUsb for the grpc NervoService
func (s *GrpcServer) ResetUsb(context.Context, *proto.ResetUsbRequest) (*proto.ResetUsbResponse, error) {
	output, err := resetUsb()
	if err != nil {
		fmt.Println("resetting failed", err)
		return nil, err
	}

	return &proto.ResetUsbResponse{
		Output: output,
	}, nil
}

// WriteToController for the grpc NervoService
func (s *GrpcServer) WriteToController(_ context.Context, request *proto.WriteToControllerRequest) (*proto.WriteToControllerResponse, error) {
	err := s.Manager.writeToController(request.ControllerPortName, request.Message)
	return &proto.WriteToControllerResponse{}, err
}

// WriteToControllerContinuously for the grpc NervoService
func (s *GrpcServer) WriteToControllerContinuously(stream proto.NervoService_WriteToControllerContinuouslyServer) error {
	firstMessage, err := stream.Recv()
	if err != nil {
		return err
	}
	writeChan := make(chan []byte)
	answer := s.Manager.writeToControllerContinuously(firstMessage.ControllerPortName, writeChan)
	if answer.err != nil {
		return answer.err
	}

	receivedChan := make(chan []byte)
	doneReceivingChan := make(chan error)
	go func() {
		var err error
		for {
			message, err := stream.Recv()
			if err != nil {
				fmt.Println(err)
				break
			}
			receivedChan <- message.Message
		}
		close(receivedChan)
		doneReceivingChan <- err
	}()

	for {
		select {
		case message := <-receivedChan:
			writeChan <- message
			break
		case <-stream.Context().Done():
		case message := <-answer.stopChan:
			close(writeChan)
			err := <-answer.doneChan
			message.doneChan <- struct{}{}
			return err
		}
	}
}
