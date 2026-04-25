package worker

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	pb "ray/src/common"
)

// TaskClient is a client for the TaskService
type TaskClient struct {
	client pb.TaskServiceClient
}

// NewTaskClient creates a new TaskClient instance
func NewTaskClient(conn *grpc.ClientConn) *TaskClient {
	client := pb.NewTaskServiceClient(conn)
	return &TaskClient{
		client: client,
	}
}

// RequestTask requests a task from the master node
func (c *TaskClient) RequestTask(workerID string) (*pb.Task, error) {
	resp, err := c.client.AssignTask(context.Background(), &pb.AssignTaskRequest{WorkerId: workerID})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to assign task: %s", resp.Message)
	}

	log.Printf("Worker %s received task %s", workerID, resp.Task.Id)
	return resp.Task, nil
}

// CompleteTask marks a task as completed
func (c *TaskClient) CompleteTask(taskID string, result *structpb.Struct) error {
	resp, err := c.client.CompleteTask(context.Background(), &pb.CompleteTaskRequest{
		TaskId: taskID,
		Result: result,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("failed to complete task: %s", resp.Message)
	}

	log.Printf("Task %s completed successfully", taskID)
	return nil
}

// FailTask marks a task as failed
func (c *TaskClient) FailTask(taskID string, error string) error {
	resp, err := c.client.FailTask(context.Background(), &pb.FailTaskRequest{
		TaskId: taskID,
		Error:  error,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("failed to fail task: %s", resp.Message)
	}

	log.Printf("Task %s failed with error: %s", taskID, error)
	return nil
}

// GetTask retrieves a task by its ID
func (c *TaskClient) GetTask(taskID string) (*pb.Task, error) {
	resp, err := c.client.GetTask(context.Background(), &pb.GetTaskRequest{TaskId: taskID})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get task: %s", resp.Message)
	}

	return resp.Task, nil
}
