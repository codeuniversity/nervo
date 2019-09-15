package nervo

import (
	"context"
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
	portNames := s.Manager.listControllers()
	infos := []*proto.ControllerInfo{}
	for _, name := range portNames {
		infos = append(infos, &proto.ControllerInfo{
			PortName: name,
		})
	}

	return &proto.ControllerListResponse{ControllerInfos: infos}, nil
}

// ReadControllerOutput for the grpc NervoService
func (s *GrpcServer) ReadControllerOutput(_ context.Context, request *proto.ReadControllerOutputRequest) (*proto.ReadControllerOutputResponse, error) {
	output := s.Manager.readFromController(request.ControllerPortName)

	return &proto.ReadControllerOutputResponse{Output: output}, nil
}
