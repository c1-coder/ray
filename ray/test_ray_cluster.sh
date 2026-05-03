#!/bin/bash

# Test script for the Ray cluster integration

# Start the master node with Ray cluster
cd /workspace/ray/cmd/master
go run main.go &
MASTER_PID=$!
sleep 10

# Start multiple worker nodes
for i in {1..3}
do
    cd /workspace/ray/cmd/worker
    WORKER_ADDRESS="localhost" MASTER_ADDRESS="localhost:50051" go run main.go &
    WORKER_PIDS[$i]=$!
    sleep 2
done

# Wait for the system to stabilize
sleep 5

# Create test tasks using Ray
mkdir -p /workspace/ray/test
cd /workspace/ray/test
cat > test_ray_tasks.go << 'EOF'
package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	pb "ray/src/common"
)

func main() {
	// Connect to the master node
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to connect to master: %v\n", err)
		return
	}
	defer conn.Close()

	// Create a task client
	client := pb.NewTaskServiceClient(conn)

	// Create a test payload
	payload := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"test_field": {
				Kind: &structpb.Value_StringValue{
					StringValue: "test_value",
				},
			},
		},
	}

	// Create monitoring tasks
	for i := 0; i < 2; i++ {
		resp, err := client.CreateTask(context.Background(), &pb.CreateTaskRequest{
			Type:     pb.TaskType_TASK_TYPE_MONITORING,
			Priority: pb.TaskPriority_TASK_PRIORITY_MEDIUM,
			Payload:  payload,
		})

		if err != nil {
			fmt.Printf("Failed to create monitoring task: %v\n", err)
			continue
		}

		if !resp.Success {
			fmt.Printf("Failed to create monitoring task: %s\n", resp.Message)
			continue
		}

		fmt.Printf("Created monitoring task: %s\n", resp.Task.Id)
	}

	// Create AI tasks
	for i := 0; i < 2; i++ {
		resp, err := client.CreateTask(context.Background(), &pb.CreateTaskRequest{
			Type:     pb.TaskType_TASK_TYPE_AI,
			Priority: pb.TaskPriority_TASK_PRIORITY_HIGH,
			Payload:  payload,
		})

		if err != nil {
			fmt.Printf("Failed to create AI task: %v\n", err)
			continue
		}

		if !resp.Success {
			fmt.Printf("Failed to create AI task: %s\n", resp.Message)
			continue
		}

		fmt.Printf("Created AI task: %s\n", resp.Task.Id)
	}

	// Create communication tasks
	for i := 0; i < 2; i++ {
		resp, err := client.CreateTask(context.Background(), &pb.CreateTaskRequest{
			Type:     pb.TaskType_TASK_TYPE_COMMUNICATION,
			Priority: pb.TaskPriority_TASK_PRIORITY_LOW,
			Payload:  payload,
		})

		if err != nil {
			fmt.Printf("Failed to create communication task: %v\n", err)
			continue
		}

		if !resp.Success {
			fmt.Printf("Failed to create communication task: %s\n", resp.Message)
			continue
		}

		fmt.Printf("Created communication task: %s\n", resp.Task.Id)
	}
}
EOF

# Run the test task creation
cd /workspace/ray/test
go run test_ray_tasks.go

# Wait for tasks to be processed
sleep 15

# Kill one worker node to test fault tolerance
echo "Killing worker node 2..."
kill ${WORKER_PIDS[2]}

# Create more tasks to test fault tolerance
cd /workspace/ray/test
go run test_ray_tasks.go

# Wait for tasks to be processed
sleep 15

# Kill the master node
kill $MASTER_PID

# Kill remaining worker nodes
for i in {1..3}
do
    if [ -n "${WORKER_PIDS[$i]}" ]; then
        kill ${WORKER_PIDS[$i]}
    fi
done

# Clean up
rm -rf /workspace/ray/test
