package common

import (
	"errors"
	"sort"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"
)

// TaskQueue manages the queue of tasks
type TaskQueue struct {
	tasks      map[string]*Task
	pending    []*Task
	mu         sync.RWMutex
	priorityOrder map[TaskPriority]int
}

// NewTaskQueue creates a new TaskQueue instance
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks:      make(map[string]*Task),
		pending:    make([]*Task, 0),
		priorityOrder: map[TaskPriority]int{
			TaskPriority_TASK_PRIORITY_CRITICAL: 0,
			TaskPriority_TASK_PRIORITY_HIGH:    1,
			TaskPriority_TASK_PRIORITY_MEDIUM:  2,
			TaskPriority_TASK_PRIORITY_LOW:     3,
		},
	}
}

// AddTask adds a task to the queue
func (q *TaskQueue) AddTask(task *Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.tasks[task.Id]; exists {
		return errors.New("task already exists")
	}

	q.tasks[task.Id] = task
	q.pending = append(q.pending, task)
	q.sortPendingTasks()

	return nil
}

// GetTask retrieves a task by its ID
func (q *TaskQueue) GetTask(taskID string) (*Task, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

// GetPendingTasks returns all pending tasks
func (q *TaskQueue) GetPendingTasks() []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	pending := make([]*Task, len(q.pending))
	copy(pending, q.pending)
	return pending
}

// AssignTask assigns the next pending task to a worker node
func (q *TaskQueue) AssignTask(workerID string) (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pending) == 0 {
		return nil, errors.New("no pending tasks")
	}

	// Get the highest priority task
	task := q.pending[0]

	// Remove the task from the pending list
	q.pending = q.pending[1:]

	// Assign the task to the worker
	task.Assign(workerID)

	return task, nil
}

// CompleteTask marks a task as completed
func (q *TaskQueue) CompleteTask(taskID string, result *structpb.Struct) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	task.Complete(result)
	return nil
}

// FailTask marks a task as failed
func (q *TaskQueue) FailTask(taskID string, error string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	task.Fail(error)
	return nil
}

// sortPendingTasks sorts the pending tasks by priority
func (q *TaskQueue) sortPendingTasks() {
	sort.Slice(q.pending, func(i, j int) bool {
		// First sort by priority
		if q.priorityOrder[q.pending[i].Priority] != q.priorityOrder[q.pending[j].Priority] {
			return q.priorityOrder[q.pending[i].Priority] < q.priorityOrder[q.pending[j].Priority]
		}
		// Then sort by creation time
		return q.pending[i].CreatedAt.AsTime().Before(q.pending[j].CreatedAt.AsTime())
	})
}

// GetTasksByStatus returns all tasks with the specified status
func (q *TaskQueue) GetTasksByStatus(status TaskStatus) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range q.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// GetTasksByWorker returns all tasks assigned to the specified worker
func (q *TaskQueue) GetTasksByWorker(workerID string) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range q.tasks {
		if task.AssignedTo == workerID && (task.Status == TaskStatus_TASK_STATUS_ASSIGNED || task.Status == TaskStatus_TASK_STATUS_RUNNING) {
			tasks = append(tasks, task)
		}
	}

	return tasks
}
