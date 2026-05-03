package ray_integration

import (
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"ray/src/common"
)

type SchedulerPolicy string

const (
	PolicyFIFO        SchedulerPolicy = "fifo"
	PolicyPriority    SchedulerPolicy = "priority"
	PolicyFairShare   SchedulerPolicy = "fair_share"
	PolicyDataLocality SchedulerPolicy = "data_locality"
)

type TaskRequest struct {
	ID          string
	Type        common.TaskType
	Priority    int32
	Payload     *structpb.Struct
	Requirements map[ResourceType]float64
	Dependencies []string
	DataLocality string
	CreatedAt   time.Time
}

type TaskScheduleResult struct {
	TaskID       string
	AssignedNode string
	ScheduleTime time.Time
	WaitTime     time.Duration
	ExecutionTime time.Duration
}

type RayScheduler struct {
	cluster      *RayCluster
	resourceManager *RayResourceManager
	pendingQueue []*TaskRequest
	scheduledTasks map[string]*TaskScheduleResult
	policy       SchedulerPolicy
	mu           sync.RWMutex
	taskHistory  []*TaskScheduleResult
	maxHistory   int
}

func NewRayScheduler(cluster *RayCluster, resourceManager *RayResourceManager, policy SchedulerPolicy) *RayScheduler {
	return &RayScheduler{
		cluster:      cluster,
		resourceManager: resourceManager,
		pendingQueue: make([]*TaskRequest, 0),
		scheduledTasks: make(map[string]*TaskScheduleResult),
		policy:       policy,
		taskHistory:  make([]*TaskScheduleResult, 0),
		maxHistory:   1000,
	}
}

func (s *RayScheduler) SubmitTask(req *TaskRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}

	s.pendingQueue = append(s.pendingQueue, req)
	s.sortQueue()

	log.Printf("Task %s submitted with priority %d", req.ID, req.Priority)
	return nil
}

func (s *RayScheduler) ScheduleTask() (*TaskScheduleResult, error) {
	s.mu.Lock()
	if len(s.pendingQueue) == 0 {
		s.mu.Unlock()
		return nil, fmt.Errorf("no pending tasks")
	}

	task := s.pendingQueue[0]
	s.pendingQueue = s.pendingQueue[1:]
	s.mu.Unlock()

	startTime := time.Now()
	
	nodeID, err := s.findBestNode(task)
	if err != nil {
		s.mu.Lock()
		task.Priority++
		s.pendingQueue = append(s.pendingQueue, task)
		s.mu.Unlock()
		return nil, fmt.Errorf("failed to schedule task: %w", err)
	}

	allocation, err := s.resourceManager.AllocateResources(nodeID, task.ID, task.Requirements)
	if err != nil {
		s.mu.Lock()
		s.pendingQueue = append(s.pendingQueue, task)
		s.mu.Unlock()
		return nil, fmt.Errorf("failed to allocate resources: %w", err)
	}

	result := &TaskScheduleResult{
		TaskID:       task.ID,
		AssignedNode: allocation.NodeID,
		ScheduleTime: time.Now(),
		WaitTime:     startTime.Sub(task.CreatedAt),
	}

	s.mu.Lock()
	s.scheduledTasks[task.ID] = result
	s.taskHistory = append(s.taskHistory, result)
	if len(s.taskHistory) > s.maxHistory {
		s.taskHistory = s.taskHistory[1:]
	}
	s.mu.Unlock()

	log.Printf("Task %s scheduled on node %s (wait time: %v)", task.ID, nodeID, result.WaitTime)
	return result, nil
}

func (s *RayScheduler) findBestNode(task *TaskRequest) (string, error) {
	switch s.policy {
	case PolicyPriority:
		return s.findNodeByPriority(task)
	case PolicyFairShare:
		return s.findNodeByFairShare(task)
	case PolicyDataLocality:
		return s.findNodeByDataLocality(task)
	default:
		return s.findNodeByFIFO(task)
	}
}

func (s *RayScheduler) findNodeByFIFO(task *TaskRequest) (string, error) {
	return s.resourceManager.FindBestNode(task.Requirements)
}

func (s *RayScheduler) findNodeByPriority(task *TaskRequest) (string, error) {
	nodes := s.cluster.GetNodes()
	bestNode := ""
	bestScore := float64(math.MaxInt32)

	for nodeID, node := range nodes {
		if node.Status != "online" {
			continue
		}

		available := s.resourceManager.GetAvailableResources(nodeID)
		if !s.canSatisfy(available, task.Requirements) {
			continue
		}

		score := s.calculatePriorityScore(node, task)
		if score < bestScore {
			bestScore = score
			bestNode = nodeID
		}
	}

	if bestNode == "" {
		return "", fmt.Errorf("no suitable node found")
	}
	return bestNode, nil
}

func (s *RayScheduler) findNodeByFairShare(task *TaskRequest) (string, error) {
	nodes := s.cluster.GetNodes()
	bestNode := ""
	minLoad := float64(math.MaxInt32)

	for nodeID, node := range nodes {
		if node.Status != "online" {
			continue
		}

		available := s.resourceManager.GetAvailableResources(nodeID)
		if !s.canSatisfy(available, task.Requirements) {
			continue
		}

		totalAvailable := 0.0
		for _, amount := range available {
			totalAvailable += amount
		}

		if totalAvailable < minLoad {
			minLoad = totalAvailable
			bestNode = nodeID
		}
	}

	if bestNode == "" {
		return "", fmt.Errorf("no suitable node found")
	}
	return bestNode, nil
}

func (s *RayScheduler) findNodeByDataLocality(task *TaskRequest) (string, error) {
	if task.DataLocality == "" {
		return s.resourceManager.FindBestNode(task.Requirements)
	}

	nodes := s.cluster.GetNodes()
	for nodeID, node := range nodes {
		if node.Status != "online" {
			continue
		}

		available := s.resourceManager.GetAvailableResources(nodeID)
		if !s.canSatisfy(available, task.Requirements) {
			continue
		}

		if node.Address == task.DataLocality {
			return nodeID, nil
		}
	}

	return s.resourceManager.FindBestNode(task.Requirements)
}

func (s *RayScheduler) calculatePriorityScore(node *RayNode, task *TaskRequest) float64 {
	score := 0.0

	available := s.resourceManager.GetAvailableResources(node.ID)
	for resType, required := range task.Requirements {
		if amount, exists := available[resType]; exists {
			score += float64(required) / float64(amount)
		}
	}

	age := time.Since(task.CreatedAt).Seconds()
	score -= age / 100.0

	return score
}

func (s *RayScheduler) canSatisfy(available, required map[ResourceType]float64) bool {
	for resType, amount := range required {
		if avail, exists := available[resType]; !exists || avail < amount {
			return false
		}
	}
	return true
}

func (s *RayScheduler) sortQueue() {
	switch s.policy {
	case PolicyPriority:
		sort.Slice(s.pendingQueue, func(i, j int) bool {
			return s.pendingQueue[i].Priority > s.pendingQueue[j].Priority
		})
	case PolicyFIFO:
		sort.Slice(s.pendingQueue, func(i, j int) bool {
			return s.pendingQueue[i].CreatedAt.Before(s.pendingQueue[j].CreatedAt)
		})
	default:
		sort.Slice(s.pendingQueue, func(i, j int) bool {
			if s.pendingQueue[i].Priority != s.pendingQueue[j].Priority {
				return s.pendingQueue[i].Priority > s.pendingQueue[j].Priority
			}
			return s.pendingQueue[i].CreatedAt.Before(s.pendingQueue[j].CreatedAt)
		})
	}
}

func (s *RayScheduler) GetPendingTasks() []*TaskRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*TaskRequest, len(s.pendingQueue))
	copy(tasks, s.pendingQueue)
	return tasks
}

func (s *RayScheduler) GetScheduledTasks() map[string]*TaskScheduleResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*TaskScheduleResult)
	for k, v := range s.scheduledTasks {
		result[k] = v
	}
	return result
}

func (s *RayScheduler) GetTaskHistory() []*TaskScheduleResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]*TaskScheduleResult, len(s.taskHistory))
	copy(history, s.taskHistory)
	return history
}

func (s *RayScheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.pendingQueue {
		if task.ID == taskID {
			s.pendingQueue = append(s.pendingQueue[:i], s.pendingQueue[i+1:]...)
			log.Printf("Task %s cancelled", taskID)
			return nil
		}
	}

	return fmt.Errorf("task %s not found", taskID)
}

func (s *RayScheduler) RescheduleTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.scheduledTasks, taskID)

	for _, task := range s.pendingQueue {
		if task.ID == taskID {
			s.pendingQueue = append(s.pendingQueue, task)
			s.sortQueue()
			log.Printf("Task %s rescheduled", taskID)
			return nil
		}
	}

	return fmt.Errorf("task %s not found", taskID)
}

func (s *RayScheduler) SetPolicy(policy SchedulerPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.policy = policy
	s.sortQueue()
	log.Printf("Scheduler policy changed to %s", policy)
}

func (s *RayScheduler) GetSchedulerStats() *SchedulerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SchedulerStats{
		PendingTasks:   len(s.pendingQueue),
		ScheduledTasks: len(s.scheduledTasks),
		Policy:         string(s.policy),
		HistorySize:    len(s.taskHistory),
	}

	var totalWaitTime time.Duration
	for _, result := range s.taskHistory {
		totalWaitTime += result.WaitTime
	}
	if len(s.taskHistory) > 0 {
		stats.AverageWaitTime = totalWaitTime / time.Duration(len(s.taskHistory))
	}

	stats.NodeDistribution = make(map[string]int)
	for _, result := range s.taskHistory {
		stats.NodeDistribution[result.AssignedNode]++
	}

	return stats
}

type SchedulerStats struct {
	PendingTasks     int
	ScheduledTasks   int
	Policy           string
	HistorySize      int
	AverageWaitTime  time.Duration
	NodeDistribution map[string]int
}

func (s *RayScheduler) PreemptTask(highPriorityTask *TaskRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pendingQueue) == 0 {
		return fmt.Errorf("no tasks to preempt")
	}

	lowestPriority := s.pendingQueue[len(s.pendingQueue)-1]
	s.pendingQueue = s.pendingQueue[:len(s.pendingQueue)-1]

	highPriorityTask.Priority = lowestPriority.Priority + 1
	s.pendingQueue = append(s.pendingQueue, highPriorityTask)
	s.sortQueue()

	log.Printf("Task %s preempted by task %s", lowestPriority.ID, highPriorityTask.ID)
	return nil
}

func (s *RayScheduler) GetTaskScheduleInfo(taskID string) (*structpb.Struct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

	for _, task := range s.pendingQueue {
		if task.ID == taskID {
			info.Fields["status"] = &structpb.Value{
				Kind: &structpb.Value_StringValue{StringValue: "pending"},
			}
			info.Fields["priority"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: float64(task.Priority)},
			}
			info.Fields["created_at"] = &structpb.Value{
				Kind: &structpb.Value_StringValue{StringValue: task.CreatedAt.Format(time.RFC3339)},
			}
			return info, nil
		}
	}

	if result, exists := s.scheduledTasks[taskID]; exists {
		info.Fields["status"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: "scheduled"},
		}
		info.Fields["node"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: result.AssignedNode},
		}
		info.Fields["schedule_time"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: result.ScheduleTime.Format(time.RFC3339)},
		}
		info.Fields["wait_time_ms"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(result.WaitTime.Milliseconds())},
		}
		return info, nil
	}

	return nil, fmt.Errorf("task %s not found", taskID)
}
