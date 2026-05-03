package master

import (
	"log"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"ray/src/common"
	pb "ray/src/common"
)

// FaultToleranceManager manages fault tolerance and failover mechanisms
type FaultToleranceManager struct {
	registry   *common.ServiceRegistry
	taskQueue  *common.TaskQueue
	checkInterval time.Duration
}

// NewFaultToleranceManager creates a new FaultToleranceManager instance
func NewFaultToleranceManager(registry *common.ServiceRegistry, taskQueue *common.TaskQueue, checkInterval time.Duration) *FaultToleranceManager {
	return &FaultToleranceManager{
		registry:   registry,
		taskQueue:  taskQueue,
		checkInterval: checkInterval,
	}
}

// Start starts the fault tolerance manager
func (ftm *FaultToleranceManager) Start() {
	go ftm.monitorWorkers()
}

// monitorWorkers monitors worker nodes for failures and handles task reassignments
func (ftm *FaultToleranceManager) monitorWorkers() {
	ticker := time.NewTicker(ftm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ftm.checkWorkerHealth()
		}
	}
}

// checkWorkerHealth checks the health of all worker nodes and handles failures
func (ftm *FaultToleranceManager) checkWorkerHealth() {
	// Clean up offline nodes
	offlineCount := ftm.registry.CleanupOfflineNodes()
	if offlineCount > 0 {
		log.Printf("Cleaned up %d offline nodes", offlineCount)
	}

	// Get all pending tasks
	pendingTasks := ftm.taskQueue.GetPendingTasks()

	// Get all available worker nodes
	availableWorkers := ftm.registry.GetWorkerNodes()

	// If there are available workers and pending tasks, ensure tasks are assigned
	if len(availableWorkers) > 0 && len(pendingTasks) > 0 {
		log.Printf("Found %d pending tasks and %d available workers", len(pendingTasks), len(availableWorkers))
		// Tasks are already in the queue and will be assigned by the task service
	}

	// Check for tasks assigned to offline workers
	ftm.checkTasksForOfflineWorkers()
}

// checkTasksForOfflineWorkers checks for tasks assigned to offline workers and reassigns them
func (ftm *FaultToleranceManager) checkTasksForOfflineWorkers() {
	// Get all tasks that are assigned or running
	assignedTasks := ftm.taskQueue.GetTasksByStatus(pb.TaskStatus_TASK_STATUS_ASSIGNED)
	runningTasks := ftm.taskQueue.GetTasksByStatus(pb.TaskStatus_TASK_STATUS_RUNNING)

	// Combine the two lists
	tasks := append(assignedTasks, runningTasks...)

	// Check each task's assigned worker
	for _, task := range tasks {
		// Check if the worker is still online
		_, err := ftm.registry.GetNode(task.AssignedTo)
		if err != nil {
			// Worker is offline, reassign the task
			log.Printf("Worker %s is offline, reassigning task %s", task.AssignedTo, task.Id)
			
			// Reset the task status to pending
			task.Status = pb.TaskStatus_TASK_STATUS_PENDING
			task.AssignedTo = ""
			task.UpdatedAt = timestamppb.New(time.Now())
			
			// Add the task back to the pending queue
			ftm.taskQueue.AddTask(task)
			log.Printf("Task %s has been reset to pending and added back to the queue", task.Id)
		}
	}
}

// HandleWorkerFailure handles the failure of a worker node
func (ftm *FaultToleranceManager) HandleWorkerFailure(workerID string) {
	log.Printf("Handling failure of worker %s", workerID)
	
	// Unregister the worker
	err := ftm.registry.Unregister(workerID)
	if err != nil {
		log.Printf("Failed to unregister worker %s: %v", workerID, err)
	}
	
	// Reassign tasks assigned to this worker
	tasks := ftm.taskQueue.GetTasksByWorker(workerID)
	for _, task := range tasks {
		// Reset the task status to pending
		task.Status = pb.TaskStatus_TASK_STATUS_PENDING
		task.AssignedTo = ""
		task.UpdatedAt = timestamppb.New(time.Now())
		
		// Add the task back to the pending queue
		ftm.taskQueue.AddTask(task)
		log.Printf("Task %s has been reset to pending and added back to the queue", task.Id)
	}
}
