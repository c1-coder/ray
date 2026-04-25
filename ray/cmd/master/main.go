package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"ray/src/common"
	"ray/src/master"
	pb "ray/src/common"
)

func main() {
	// Create a new service registry with a 30-second heartbeat timeout
	registry := common.NewServiceRegistry(30 * time.Second)

	// Create a new task queue
	taskQueue := common.NewTaskQueue()

	// Create a new registry server
	registryServer := master.NewRegistryServer(registry)

	// Create a new task server
	taskServer := master.NewTaskServer(taskQueue)

	// Create a new fault tolerance manager
	ftm := master.NewFaultToleranceManager(registry, taskQueue, 10*time.Second)

	// Start the fault tolerance manager
	ftm.Start()

	// Start the gRPC server
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	s := grpc.NewServer()

	// Register the registry server with the gRPC server
	pb.RegisterServiceRegistryServiceServer(s, registryServer)

	// Register the task server with the gRPC server
	pb.RegisterTaskServiceServer(s, taskServer)

	// Start the server
	log.Printf("Master node started on port %d", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
