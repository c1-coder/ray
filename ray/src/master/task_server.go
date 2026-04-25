package master

import (
	"context"
	"log"

	"github.com/google/uuid"

	"ray/src/common"
	pb "ray/src/common"
)

// TaskServer implements the TaskService gRPC service
type TaskServer struct {
	pb.UnimplementedTaskServiceServer
	taskQueue *common.TaskQueue
}

// NewTaskServer creates a new TaskServer instance
func NewTaskServer(taskQueue *common.TaskQueue) *TaskServer {
	return &TaskServer{
		taskQueue: taskQueue,
	}
}

// CreateTask implements the CreateTask method of the TaskService
func (s *TaskServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	// Generate a unique ID for the task
	taskID := uuid.New().String()

	// Create a new task
	task := common.NewTask(
		taskID,
		req.Type,
		req.Priority,
		req.Payload,
	)

	// Add the task to the queue
	err := s.taskQueue.AddTask(task)
	if err != nil {
		log.Printf("Failed to create task: %v", err)
		return &pb.CreateTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Task %s created successfully", taskID)
	return &pb.CreateTaskResponse{
		Success: true,
		Message: "Task created successfully",
		Task:    task,
	}, nil
}

// GetTask implements the GetTask method of the TaskService
func (s *TaskServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	// Get the task from the queue
	task, err := s.taskQueue.GetTask(req.TaskId)
	if err != nil {
		log.Printf("Failed to get task %s: %v", req.TaskId, err)
		return &pb.GetTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetTaskResponse{
		Success: true,
		Message: "Task retrieved successfully",
		Task:    task,
	}, nil
}

// AssignTask implements the AssignTask method of the TaskService
func (s *TaskServer) AssignTask(ctx context.Context, req *pb.AssignTaskRequest) (*pb.AssignTaskResponse, error) {
	// Assign the next pending task to the worker
	task, err := s.taskQueue.AssignTask(req.WorkerId)
	if err != nil {
		log.Printf("Failed to assign task to worker %s: %v", req.WorkerId, err)
		return &pb.AssignTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Task %s assigned to worker %s", task.Id, req.WorkerId)
	return &pb.AssignTaskResponse{
		Success: true,
		Message: "Task assigned successfully",
		Task:    task,
	}, nil
}

// CompleteTask implements the CompleteTask method of the TaskService
func (s *TaskServer) CompleteTask(ctx context.Context, req *pb.CompleteTaskRequest) (*pb.CompleteTaskResponse, error) {
	// Complete the task
	err := s.taskQueue.CompleteTask(req.TaskId, req.Result)
	if err != nil {
		log.Printf("Failed to complete task %s: %v", req.TaskId, err)
		return &pb.CompleteTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Get the updated task
	task, err := s.taskQueue.GetTask(req.TaskId)
	if err != nil {
		log.Printf("Failed to get completed task %s: %v", req.TaskId, err)
		return &pb.CompleteTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Task %s completed successfully", req.TaskId)
	return &pb.CompleteTaskResponse{
		Success: true,
		Message: "Task completed successfully",
		Task:    task,
	}, nil
}

// FailTask implements the FailTask method of the TaskService
func (s *TaskServer) FailTask(ctx context.Context, req *pb.FailTaskRequest) (*pb.FailTaskResponse, error) {
	// Fail the task
	err := s.taskQueue.FailTask(req.TaskId, req.Error)
	if err != nil {
		log.Printf("Failed to fail task %s: %v", req.TaskId, err)
		return &pb.FailTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Get the updated task
	task, err := s.taskQueue.GetTask(req.TaskId)
	if err != nil {
		log.Printf("Failed to get failed task %s: %v", req.TaskId, err)
		return &pb.FailTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Task %s failed with error: %s", req.TaskId, req.Error)
	return &pb.FailTaskResponse{
		Success: true,
		Message: "Task failed successfully",
		Task:    task,
	}, nil
}

// GetPendingTasks implements the GetPendingTasks method of the TaskService
func (s *TaskServer) GetPendingTasks(ctx context.Context, req *pb.GetPendingTasksRequest) (*pb.GetPendingTasksResponse, error) {
	// Get all pending tasks
	tasks := s.taskQueue.GetPendingTasks()

	return &pb.GetPendingTasksResponse{
		Success: true,
		Message: "Pending tasks retrieved successfully",
		Tasks:   tasks,
	}, nil
}

// GetTasksByWorker implements the GetTasksByWorker method of the TaskService
func (s *TaskServer) GetTasksByWorker(ctx context.Context, req *pb.GetTasksByWorkerRequest) (*pb.GetTasksByWorkerResponse, error) {
	// Get tasks by worker
	tasks := s.taskQueue.GetTasksByWorker(req.WorkerId)

	return &pb.GetTasksByWorkerResponse{
		Success: true,
		Message: "Tasks by worker retrieved successfully",
		Tasks:   tasks,
	}, nil
}
