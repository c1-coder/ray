package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "ray/src/common"
)

// RegistryClient is a client for the ServiceRegistryService
type RegistryClient struct {
	client pb.ServiceRegistryServiceClient
	conn   *grpc.ClientConn
}

// NewRegistryClient creates a new RegistryClient instance
func NewRegistryClient(masterAddress string) (*RegistryClient, error) {
	conn, err := grpc.Dial(masterAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewServiceRegistryServiceClient(conn)
	return &RegistryClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *RegistryClient) Close() error {
	return c.conn.Close()
}

// Register registers the worker node with the master
func (c *RegistryClient) Register(node *pb.NodeInfo) error {
	resp, err := c.client.Register(context.Background(), &pb.RegisterRequest{Node: node})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	log.Printf("Worker node %s registered successfully", node.Id)
	return nil
}

// Unregister unregisters the worker node from the master
func (c *RegistryClient) Unregister(nodeID string) error {
	resp, err := c.client.Unregister(context.Background(), &pb.UnregisterRequest{NodeId: nodeID})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("unregistration failed: %s", resp.Message)
	}

	log.Printf("Worker node %s unregistered successfully", nodeID)
	return nil
}

// SendHeartbeat sends a heartbeat to the master
func (c *RegistryClient) SendHeartbeat(nodeID string) error {
	resp, err := c.client.Heartbeat(context.Background(), &pb.HeartbeatRequest{NodeId: nodeID})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("heartbeat failed: %s", resp.Message)
	}

	return nil
}

// StartHeartbeatLoop starts a loop to send heartbeats periodically
func (c *RegistryClient) StartHeartbeatLoop(nodeID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.SendHeartbeat(nodeID)
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}
