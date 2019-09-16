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
	if answer.Error != nil {
		return nil, answer.Error
	}
	return &proto.FlashControllerResponse{Output: answer.Output}, nil
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
