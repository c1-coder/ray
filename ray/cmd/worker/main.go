package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"ray/src/common"
	"ray/src/worker"
	pb "ray/src/common"
)

func main() {
	// Get the master address from the environment variable
	masterAddress := os.Getenv("MASTER_ADDRESS")
	if masterAddress == "" {
		masterAddress = "localhost:50051"
	}

	// Generate a unique ID for the worker node
	nodeID := uuid.New().String()

	// Get the worker address and port
	workerAddress := os.Getenv("WORKER_ADDRESS")
	if workerAddress == "" {
		workerAddress = "localhost"
	}

	workerPort := 50052

	// Create a new node info
	node := common.NewNodeInfo(
		nodeID,
		pb.NodeType_WORKER_NODE,
		workerAddress,
		workerPort,
		[]string{"task_execution", "monitoring", "ai_assistance"},
	)

	// Create a gRPC connection to the master
	conn, err := grpc.Dial(masterAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()

	// Create a new registry client
	registryClient, err := worker.NewRegistryClient(masterAddress)
	if err != nil {
		log.Fatalf("Failed to create registry client: %v", err)
	}
	defer registryClient.Close()

	// Create a new task client
	taskClient := worker.NewTaskClient(conn)

	// Register the worker node with the master
	err = registryClient.Register(node)
	if err != nil {
		log.Fatalf("Failed to register worker node: %v", err)
	}

	// Start the heartbeat loop
	go registryClient.StartHeartbeatLoop(nodeID, 5*time.Second)

	// Start the task processing loop
	go startTaskProcessingLoop(nodeID, taskClient)

	// Keep the worker node running
	log.Printf("Worker node %s started and registered with master at %s", nodeID, masterAddress)
	select {}
}

// startTaskProcessingLoop starts a loop to process tasks
func startTaskProcessingLoop(workerID string, taskClient *worker.TaskClient) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Request a task from the master
			task, err := taskClient.RequestTask(workerID)
			if err != nil {
				// No tasks available, continue
				continue
			}

			// Process the task
			result, err := processTask(task)
			if err != nil {
				// Fail the task
				err := taskClient.FailTask(task.Id, err.Error())
				if err != nil {
					log.Printf("Failed to fail task %s: %v", task.Id, err)
				}
				continue
			}

			// Complete the task
			err = taskClient.CompleteTask(task.Id, result)
			if err != nil {
				log.Printf("Failed to complete task %s: %v", task.Id, err)
			}
		}
	}
}

// processTask processes a task based on its type
func processTask(task *pb.Task) (*structpb.Struct, error) {
	log.Printf("Processing task %s of type %s", task.Id, task.Type)

	// Create a result struct
	result := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	// Simulate task processing based on type
	switch task.Type {
	case pb.TaskType_TASK_TYPE_MONITORING:
		// Simulate monitoring task
		data := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"cpu_usage": {
					Kind: &structpb.Value_NumberValue{NumberValue: 0.5},
				},
				"memory_usage": {
					Kind: &structpb.Value_NumberValue{NumberValue: 0.3},
				},
				"disk_usage": {
					Kind: &structpb.Value_NumberValue{NumberValue: 0.7},
				},
			},
		}

		result.Fields["status"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "success"},
		}
		result.Fields["message"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "Monitoring task completed"},
		}
		result.Fields["data"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: data},
		}

	case pb.TaskType_TASK_TYPE_AI:
		// Simulate AI task
		data := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"result": {
					Kind: &structpb.Value_StringValue{StringValue: "AI task result"},
				},
				"confidence": {
					Kind: &structpb.Value_NumberValue{NumberValue: 0.95},
				},
			},
		}

		result.Fields["status"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "success"},
		}
		result.Fields["message"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "AI task completed"},
		}
		result.Fields["data"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: data},
		}

	case pb.TaskType_TASK_TYPE_COMMUNICATION:
		// Simulate communication task
		data := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"recipient": {
					Kind: &structpb.Value_StringValue{StringValue: "user@example.com"},
				},
				"message": {
					Kind: &structpb.Value_StringValue{StringValue: "Hello from the distributed task platform"},
				},
			},
		}

		result.Fields["status"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "success"},
		}
		result.Fields["message"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "Communication task completed"},
		}
		result.Fields["data"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: data},
		}

	case pb.TaskType_TASK_TYPE_CUSTOM:
		// Simulate custom task
		result.Fields["status"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "success"},
		}
		result.Fields["message"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "Custom task completed"},
		}
		result.Fields["data"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: task.Payload},
		}

	default:
		return nil, fmt.Errorf("unknown task type: %s", task.Type)
	}

	return result, nil
}
