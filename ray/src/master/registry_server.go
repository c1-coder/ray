package master

import (
	"context"
	"log"

	"ray/src/common"
	pb "ray/src/common"
)

// RegistryServer implements the ServiceRegistryService gRPC service
type RegistryServer struct {
	pb.UnimplementedServiceRegistryServiceServer
	registry *common.ServiceRegistry
}

// NewRegistryServer creates a new RegistryServer instance
func NewRegistryServer(registry *common.ServiceRegistry) *RegistryServer {
	return &RegistryServer{
		registry: registry,
	}
}

// Register implements the Register method of the ServiceRegistryService
func (s *RegistryServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := s.registry.Register(req.Node)
	if err != nil {
		log.Printf("Failed to register node %s: %v", req.Node.Id, err)
		return &pb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Node %s registered successfully", req.Node.Id)
	return &pb.RegisterResponse{
		Success: true,
		Message: "Node registered successfully",
	}, nil
}

// Unregister implements the Unregister method of the ServiceRegistryService
func (s *RegistryServer) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	err := s.registry.Unregister(req.NodeId)
	if err != nil {
		log.Printf("Failed to unregister node %s: %v", req.NodeId, err)
		return &pb.UnregisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Node %s unregistered successfully", req.NodeId)
	return &pb.UnregisterResponse{
		Success: true,
		Message: "Node unregistered successfully",
	}, nil
}

// Heartbeat implements the Heartbeat method of the ServiceRegistryService
func (s *RegistryServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	err := s.registry.UpdateHeartbeat(req.NodeId)
	if err != nil {
		log.Printf("Failed to update heartbeat for node %s: %v", req.NodeId, err)
		return &pb.HeartbeatResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.HeartbeatResponse{
		Success: true,
		Message: "Heartbeat updated successfully",
	}, nil
}

// GetNode implements the GetNode method of the ServiceRegistryService
func (s *RegistryServer) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	node, err := s.registry.GetNode(req.NodeId)
	if err != nil {
		log.Printf("Failed to get node %s: %v", req.NodeId, err)
		return &pb.GetNodeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetNodeResponse{
		Success: true,
		Message: "Node retrieved successfully",
		Node:    node,
	}, nil
}

// GetAllNodes implements the GetAllNodes method of the ServiceRegistryService
func (s *RegistryServer) GetAllNodes(ctx context.Context, req *pb.GetAllNodesRequest) (*pb.GetAllNodesResponse, error) {
	nodes := s.registry.GetAllNodes()

	return &pb.GetAllNodesResponse{
		Success: true,
		Message: "Nodes retrieved successfully",
		Nodes:   nodes,
	}, nil
}

// GetWorkerNodes implements the GetWorkerNodes method of the ServiceRegistryService
func (s *RegistryServer) GetWorkerNodes(ctx context.Context, req *pb.GetWorkerNodesRequest) (*pb.GetWorkerNodesResponse, error) {
	nodes := s.registry.GetWorkerNodes()

	return &pb.GetWorkerNodesResponse{
		Success: true,
		Message: "Worker nodes retrieved successfully",
		Nodes:   nodes,
	}, nil
}
