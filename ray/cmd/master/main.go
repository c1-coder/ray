package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"ray/src/common"
	"ray/src/master"
	"ray/src/ray_integration"
	pb "ray/src/common"
)

func main() {
	// Create a new service registry with a 30-second heartbeat timeout
	registry := common.NewServiceRegistry(30 * time.Second)

	// Create a new task queue
	taskQueue := common.NewTaskQueue()

	// Create and start Ray cluster
	rayConfig := &ray_integration.RayConfig{
		MasterAddress: "localhost:6379",
		WorkerNodes:   []string{"localhost:6380", "localhost:6381", "localhost:6382"},
		Resources:     map[string]float64{"CPU": 4.0, "memory": 8192.0},
	}

	rayCluster := ray_integration.NewRayCluster(rayConfig, registry, taskQueue)
	if err := rayCluster.Start(); err != nil {
		log.Printf("Failed to start Ray cluster: %v", err)
	} else {
		log.Println("Ray cluster started successfully")
	}

	// Create a new registry server
	registryServer := master.NewRegistryServer(registry)

	// Create a new task server
	taskServer := master.NewTaskServer(taskQueue)

	// Create a new fault tolerance manager
	ftm := master.NewFaultToleranceManager(registry, taskQueue, 10*time.Second)

	// Create a vulnerability manager and scan tools
	vulnerabilityManager := common.NewVulnerabilityManager("/workspace")
	go func() {
		err := vulnerabilityManager.ScanTools()
		if err != nil {
			log.Printf("Error scanning tools: %v", err)
		}
		
		report, err := vulnerabilityManager.GenerateSecurityReport()
		if err != nil {
			log.Printf("Error generating security report: %v", err)
		} else {
			log.Println("Security report generated:")
			log.Println(report)
		}
	}()

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
