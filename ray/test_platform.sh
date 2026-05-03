#!/bin/bash

# Test script for the distributed task management platform

# Start the master node
cd /workspace/ray/cmd/master
go run main.go &
MASTER_PID=$!
sleep 5

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

# Create a test task
mkdir -p /workspace/ray/test
cd /workspace/ray/test
cat > test_task.go << 'EOF'
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

	// Create a test task
	payload := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"test_field": {
				Kind: &structpb.Value_StringValue{
					StringValue: "test_value",
				},
			},
		},
	}

	// Create the task
	resp, err := client.CreateTask(context.Background(), &pb.CreateTaskRequest{
		Type:     pb.TaskType_TASK_TYPE_MONITORING,
		Priority: pb.TaskPriority_TASK_PRIORITY_MEDIUM,
		Payload:  payload,
	})

	if err != nil {
		fmt.Printf("Failed to create task: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Failed to create task: %s\n", resp.Message)
		return
	}

	fmt.Printf("Created task: %s\n", resp.Task.Id)

	// Create more tasks
	for i := 0; i < 5; i++ {
		resp, err := client.CreateTask(context.Background(), &pb.CreateTaskRequest{
			Type:     pb.TaskType_TASK_TYPE_AI,
			Priority: pb.TaskPriority_TASK_PRIORITY_LOW,
			Payload:  payload,
		})

		if err != nil {
			fmt.Printf("Failed to create task: %v\n", err)
			continue
		}

		if !resp.Success {
			fmt.Printf("Failed to create task: %s\n", resp.Message)
			continue
		}

		fmt.Printf("Created task: %s\n", resp.Task.Id)
	}
}
EOF

# Run the test task creation
cd /workspace/ray/test
go run test_task.go

# Wait for tasks to be processed
sleep 10

# Kill one worker node to test fault tolerance
echo "Killing worker node 2..."
kill ${WORKER_PIDS[2]}

# Create more tasks to test fault tolerance
cd /workspace/ray/test
go run test_task.go

# Wait for tasks to be processed
sleep 10

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
rm /workspace/ray/test/test_task.go
rmdir /workspace/ray/test

echo "Test completed"
