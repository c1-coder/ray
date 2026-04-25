package common

import (
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewTask creates a new Task instance
func NewTask(id string, taskType TaskType, priority TaskPriority, payload *structpb.Struct) *Task {
	now := time.Now()
	timestampNow := timestamppb.New(now)
	return &Task{
		Id:          id,
		Type:        taskType,
		Priority:    priority,
		Status:      TaskStatus_TASK_STATUS_PENDING,
		Payload:     payload,
		CreatedAt:   timestampNow,
		UpdatedAt:   timestampNow,
	}
}

// UpdateStatus updates the status of the task
func (t *Task) UpdateStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = timestamppb.New(time.Now())
	if status == TaskStatus_TASK_STATUS_COMPLETED {
		t.CompletedAt = t.UpdatedAt
	}
}

// Assign assigns the task to a worker node
func (t *Task) Assign(workerID string) {
	t.AssignedTo = workerID
	t.Status = TaskStatus_TASK_STATUS_ASSIGNED
	t.UpdatedAt = timestamppb.New(time.Now())
}

// Complete marks the task as completed with a result
func (t *Task) Complete(result *structpb.Struct) {
	t.Result = result
	t.Status = TaskStatus_TASK_STATUS_COMPLETED
	t.UpdatedAt = timestamppb.New(time.Now())
	t.CompletedAt = t.UpdatedAt
}

// Fail marks the task as failed with an error
func (t *Task) Fail(error string) {
	t.Error = error
	t.Status = TaskStatus_TASK_STATUS_FAILED
	t.UpdatedAt = timestamppb.New(time.Now())
}
