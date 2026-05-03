package ray_integration

import (
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"ray/src/common"
)

type NodeState string

const (
	NodeStatePending   NodeState = "pending"
	NodeStateStarting  NodeState = "starting"
	NodeStateRunning   NodeState = "running"
	NodeStateStopping  NodeState = "stopping"
	NodeStateStopped  NodeState = "stopped"
	NodeStateFailed    NodeState = "failed"
	NodeStateScalingUp   NodeState = "scaling_up"
	NodeStateScalingDown NodeState = "scaling_down"
)

type NodeConfig struct {
	ID          string
	Address     string
	Port        int
	NodeType    string
	Resources   map[ResourceType]float64
	Labels      map[string]string
	Annotations map[string]string
	MaxRetries  int
	Timeout     time.Duration
}

type NodeMetrics struct {
	NodeID            string
	CPUUsage          float64
	MemoryUsage       float64
	DiskUsage         float64
	NetworkIn         float64
	NetworkOut        float64
	ActiveTasks       int
	CompletedTasks    int
	FailedTasks       int
	Uptime            time.Duration
	LastMetricUpdated time.Time
}

type RayNodeManager struct {
	cluster         *RayCluster
	nodes           map[string]*ManagedNode
	nodeConfigs     map[string]*NodeConfig
	nodeMetrics     map[string]*NodeMetrics
	state           NodeState
	mu              sync.RWMutex
	eventHandlers   []NodeEventHandler
}

type ManagedNode struct {
	ID            string
	Config        *NodeConfig
	State         NodeState
	Status        string
	Resources     map[ResourceType]float64
	Labels        map[string]string
	LastHeartbeat time.Time
	Metrics       *NodeMetrics
	RetryCount    int
	CreatedAt     time.Time
	StartedAt     time.Time
}

type NodeEventHandler func(event *NodeEvent)

type NodeEvent struct {
	Type      string
	NodeID    string
	NodeState NodeState
	Message   string
	Timestamp time.Time
	Data      map[string]interface{}
}

func NewRayNodeManager(cluster *RayCluster) *RayNodeManager {
	return &RayNodeManager{
		cluster:     cluster,
		nodes:       make(map[string]*ManagedNode),
		nodeConfigs: make(map[string]*NodeConfig),
		nodeMetrics: make(map[string]*NodeMetrics),
		state:       NodeStateStopped,
	}
}

func (nm *RayNodeManager) RegisterNode(config *NodeConfig) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.nodes[config.ID]; exists {
		return fmt.Errorf("node %s already registered", config.ID)
	}

	node := &ManagedNode{
		ID:        config.ID,
		Config:    config,
		State:     NodeStatePending,
		Status:    "registered",
		Resources: config.Resources,
		Labels:    config.Labels,
		CreatedAt: time.Now(),
		Metrics: &NodeMetrics{
			NodeID: config.ID,
		},
	}

	nm.nodes[config.ID] = node
	nm.nodeConfigs[config.ID] = config
	nm.nodeMetrics[config.ID] = node.Metrics

	nm.emitEvent(&NodeEvent{
		Type:      "node_registered",
		NodeID:    config.ID,
		NodeState: NodeStatePending,
		Message:   fmt.Sprintf("Node %s registered", config.ID),
		Timestamp: time.Now(),
	})

	log.Printf("Node %s registered", config.ID)
	return nil
}

func (nm *RayNodeManager) StartNode(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	if node.State == NodeStateRunning {
		return fmt.Errorf("node %s already running", nodeID)
	}

	node.State = NodeStateStarting
	node.Status = "starting"
	node.StartedAt = time.Now()

	nm.emitEvent(&NodeEvent{
		Type:      "node_starting",
		NodeID:    nodeID,
		NodeState: NodeStateStarting,
		Message:   fmt.Sprintf("Node %s starting", nodeID),
		Timestamp: time.Now(),
	})

	go nm.startNodeAsync(nodeID)

	log.Printf("Node %s starting", nodeID)
	return nil
}

func (nm *RayNodeManager) startNodeAsync(nodeID string) {
	time.Sleep(100 * time.Millisecond)

	nm.mu.Lock()
	if node, exists := nm.nodes[nodeID]; exists {
		node.State = NodeStateRunning
		node.Status = "online"
		node.LastHeartbeat = time.Now()
	}
	nm.mu.Unlock()

	nm.emitEvent(&NodeEvent{
		Type:      "node_started",
		NodeID:    nodeID,
		NodeState: NodeStateRunning,
		Message:   fmt.Sprintf("Node %s started successfully", nodeID),
		Timestamp: time.Now(),
	})

	log.Printf("Node %s started successfully", nodeID)
}

func (nm *RayNodeManager) StopNode(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	if node.State == NodeStateStopped {
		return fmt.Errorf("node %s already stopped", nodeID)
	}

	node.State = NodeStateStopping
	node.Status = "stopping"

	nm.emitEvent(&NodeEvent{
		Type:      "node_stopping",
		NodeID:    nodeID,
		NodeState: NodeStateStopping,
		Message:   fmt.Sprintf("Node %s stopping", nodeID),
		Timestamp: time.Now(),
	})

	go func() {
		time.Sleep(100 * time.Millisecond)

		nm.mu.Lock()
		if n, exists := nm.nodes[nodeID]; exists {
			n.State = NodeStateStopped
			n.Status = "offline"
		}
		nm.mu.Unlock()

		nm.emitEvent(&NodeEvent{
			Type:      "node_stopped",
			NodeID:    nodeID,
			NodeState: NodeStateStopped,
			Message:   fmt.Sprintf("Node %s stopped", nodeID),
			Timestamp: time.Now(),
		})
	}()

	log.Printf("Node %s stopping", nodeID)
	return nil
}

func (nm *RayNodeManager) RestartNode(nodeID string) error {
	if err := nm.StopNode(nodeID); err != nil {
		return err
	}

	time.Sleep(200 * time.Millisecond)
	return nm.StartNode(nodeID)
}

func (nm *RayNodeManager) UnregisterNode(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	if node.State == NodeStateRunning {
		return fmt.Errorf("cannot unregister running node %s", nodeID)
	}

	delete(nm.nodes, nodeID)
	delete(nm.nodeConfigs, nodeID)
	delete(nm.nodeMetrics, nodeID)

	nm.emitEvent(&NodeEvent{
		Type:      "node_unregistered",
		NodeID:    nodeID,
		NodeState: NodeStateStopped,
		Message:   fmt.Sprintf("Node %s unregistered", nodeID),
		Timestamp: time.Now(),
	})

	log.Printf("Node %s unregistered", nodeID)
	return nil
}

func (nm *RayNodeManager) GetNode(nodeID string) (*ManagedNode, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node, nil
}

func (nm *RayNodeManager) GetAllNodes() []*ManagedNode {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodes := make([]*ManagedNode, 0, len(nm.nodes))
	for _, node := range nm.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (nm *RayNodeManager) GetNodesByState(state NodeState) []*ManagedNode {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodes := make([]*ManagedNode, 0)
	for _, node := range nm.nodes {
		if node.State == state {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (nm *RayNodeManager) GetNodesByLabel(labelKey, labelValue string) []*ManagedNode {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodes := make([]*ManagedNode, 0)
	for _, node := range nm.nodes {
		if value, exists := node.Labels[labelKey]; exists && value == labelValue {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (nm *RayNodeManager) UpdateNodeMetrics(nodeID string, metrics *NodeMetrics) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	metrics.NodeID = nodeID
	metrics.LastMetricUpdated = time.Now()
	node.Metrics = metrics
	nm.nodeMetrics[nodeID] = metrics

	return nil
}

func (nm *RayNodeManager) UpdateHeartbeat(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.LastHeartbeat = time.Now()
	if node.State != NodeStateRunning {
		node.State = NodeStateRunning
		node.Status = "online"
	}

	return nil
}

func (nm *RayNodeManager) ScaleUp(count int) error {
	nm.mu.Lock()
	nm.state = NodeStateScalingUp
	nm.mu.Unlock()

	nm.emitEvent(&NodeEvent{
		Type:    "scale_up_started",
		Message: fmt.Sprintf("Scaling up by %d nodes", count),
		Data: map[string]interface{}{
			"count": count,
		},
		Timestamp: time.Now(),
	})

	for i := 0; i < count; i++ {
		nodeID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
		config := &NodeConfig{
			ID: nodeID,
			Resources: map[ResourceType]float64{
				ResourceCPU:    2.0,
				ResourceMemory: 4096.0,
			},
			Labels: map[string]string{
				"role": "worker",
			},
		}

		if err := nm.RegisterNode(config); err != nil {
			log.Printf("Failed to register node during scale up: %v", err)
			continue
		}

		if err := nm.StartNode(nodeID); err != nil {
			log.Printf("Failed to start node during scale up: %v", err)
			continue
		}
	}

	nm.mu.Lock()
	nm.state = NodeStateRunning
	nm.mu.Unlock()

	nm.emitEvent(&NodeEvent{
		Type:    "scale_up_completed",
		Message: fmt.Sprintf("Scaled up by %d nodes", count),
		Data: map[string]interface{}{
			"count": count,
		},
		Timestamp: time.Now(),
	})

	log.Printf("Scaled up by %d nodes", count)
	return nil
}

func (nm *RayNodeManager) ScaleDown(count int) error {
	nm.mu.Lock()
	nm.state = NodeStateScalingDown
	nm.mu.Unlock()

	nm.emitEvent(&NodeEvent{
		Type:    "scale_down_started",
		Message: fmt.Sprintf("Scaling down by %d nodes", count),
		Data: map[string]interface{}{
			"count": count,
		},
		Timestamp: time.Now(),
	})

	runningNodes := nm.GetNodesByState(NodeStateRunning)
	workers := make([]*ManagedNode, 0)
	for _, node := range runningNodes {
		if node.Config.NodeType == "worker" {
			workers = append(workers, node)
		}
	}

	stopped := 0
	for i := 0; i < count && i < len(workers); i++ {
		if err := nm.StopNode(workers[i].ID); err != nil {
			log.Printf("Failed to stop node during scale down: %v", err)
			continue
		}
		stopped++
	}

	nm.mu.Lock()
	nm.state = NodeStateRunning
	nm.mu.Unlock()

	nm.emitEvent(&NodeEvent{
		Type:    "scale_down_completed",
		Message: fmt.Sprintf("Scaled down by %d nodes", stopped),
		Data: map[string]interface{}{
			"count": stopped,
		},
		Timestamp: time.Now(),
	})

	log.Printf("Scaled down by %d nodes", stopped)
	return nil
}

func (nm *RayNodeManager) RegisterEventHandler(handler NodeEventHandler) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.eventHandlers = append(nm.eventHandlers, handler)
}

func (nm *RayNodeManager) emitEvent(event *NodeEvent) {
	nm.mu.RLock()
	handlers := make([]NodeEventHandler, len(nm.eventHandlers))
	copy(handlers, nm.eventHandlers)
	nm.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

func (nm *RayNodeManager) GetNodeStats() *NodeManagerStats {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	stats := &NodeManagerStats{
		TotalNodes:      len(nm.nodes),
		NodeStats:       make(map[string]*NodeDetailStats),
	}

	stateCounts := make(map[NodeState]int)
	for _, node := range nm.nodes {
		stateCounts[node.State]++
	}
	stats.StateDistribution = stateCounts

	for nodeID, node := range nm.nodes {
		detail := &NodeDetailStats{
			ID:         node.ID,
			State:      string(node.State),
			Status:     node.Status,
			Uptime:     time.Since(node.StartedAt),
			Labels:     node.Labels,
			RetryCount: node.RetryCount,
		}

		if node.Metrics != nil {
			detail.CPUUsage = node.Metrics.CPUUsage
			detail.MemoryUsage = node.Metrics.MemoryUsage
			detail.DiskUsage = node.Metrics.DiskUsage
			detail.ActiveTasks = node.Metrics.ActiveTasks
			detail.CompletedTasks = node.Metrics.CompletedTasks
			detail.FailedTasks = node.Metrics.FailedTasks
		}

		stats.NodeStats[nodeID] = detail
	}

	return stats
}

type NodeManagerStats struct {
	TotalNodes        int
	StateDistribution map[NodeState]int
	NodeStats         map[string]*NodeDetailStats
}

type NodeDetailStats struct {
	ID             string
	State          string
	Status         string
	Uptime         time.Duration
	CPUUsage       float64
	MemoryUsage    float64
	DiskUsage      float64
	ActiveTasks    int
	CompletedTasks int
	FailedTasks    int
	Labels         map[string]string
	RetryCount     int
}

func (nm *RayNodeManager) GetNodeInfo(nodeID string) (*structpb.Struct, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	fields := make(map[string]*structpb.Value)

	fields["id"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: node.ID},
	}
	fields["state"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: string(node.State)},
	}
	fields["status"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: node.Status},
	}
	fields["created_at"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: node.CreatedAt.Format(time.RFC3339)},
	}
	fields["uptime_seconds"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: time.Since(node.StartedAt).Seconds()},
	}

	if node.Config != nil {
		fields["address"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: node.Config.Address},
		}
		fields["port"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(node.Config.Port)},
		}
		fields["type"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: node.Config.NodeType},
		}
	}

	if node.Metrics != nil {
		metricsFields := make(map[string]*structpb.Value)
		metricsFields["cpu_usage"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: node.Metrics.CPUUsage},
		}
		metricsFields["memory_usage"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: node.Metrics.MemoryUsage},
		}
		metricsFields["disk_usage"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: node.Metrics.DiskUsage},
		}
		metricsFields["active_tasks"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(node.Metrics.ActiveTasks)},
		}
		metricsFields["completed_tasks"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(node.Metrics.CompletedTasks)},
		}
		metricsFields["failed_tasks"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(node.Metrics.FailedTasks)},
		}
		fields["metrics"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: metricsFields}},
		}
	}

	return &structpb.Struct{Fields: fields}, nil
}

func (nm *RayNodeManager) DrainNode(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.Status = "draining"
	node.Config.Annotations["draining"] = "true"

	nm.emitEvent(&NodeEvent{
		Type:      "node_draining",
		NodeID:    nodeID,
		NodeState: node.State,
		Message:   fmt.Sprintf("Node %s is draining", nodeID),
		Timestamp: time.Now(),
	})

	log.Printf("Node %s is draining", nodeID)
	return nil
}

func (nm *RayNodeManager) CordonNode(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.Labels["node.kubernetes.io/unschedulable"] = "true"

	nm.emitEvent(&NodeEvent{
		Type:      "node_cordoned",
		NodeID:    nodeID,
		NodeState: node.State,
		Message:   fmt.Sprintf("Node %s cordoned", nodeID),
		Timestamp: time.Now(),
	})

	log.Printf("Node %s cordoned", nodeID)
	return nil
}

func (nm *RayNodeManager) UncordonNode(nodeID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	delete(node.Labels, "node.kubernetes.io/unschedulable")

	nm.emitEvent(&NodeEvent{
		Type:      "node_uncordoned",
		NodeID:    nodeID,
		NodeState: node.State,
		Message:   fmt.Sprintf("Node %s uncordoned", nodeID),
		Timestamp: time.Now(),
	})

	log.Printf("Node %s uncordoned", nodeID)
	return nil
}
