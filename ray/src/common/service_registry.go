package common

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ServiceRegistry manages the registration and discovery of nodes

type ServiceRegistry struct {
	nodes      map[string]*NodeInfo
	mu         sync.RWMutex
	heartbeatTimeout time.Duration
}

// NewServiceRegistry creates a new ServiceRegistry instance
func NewServiceRegistry(heartbeatTimeout time.Duration) *ServiceRegistry {
	return &ServiceRegistry{
		nodes:      make(map[string]*NodeInfo),
		heartbeatTimeout: heartbeatTimeout,
	}
}

// Register registers a node with the registry
func (r *ServiceRegistry) Register(node *NodeInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[node.Id]; exists {
		return errors.New("node already registered")
	}

	r.nodes[node.Id] = node
	return nil
}

// Unregister removes a node from the registry
func (r *ServiceRegistry) Unregister(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; !exists {
		return errors.New("node not found")
	}

	delete(r.nodes, nodeID)
	return nil
}

// UpdateHeartbeat updates the heartbeat timestamp for a node
func (r *ServiceRegistry) UpdateHeartbeat(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return errors.New("node not found")
	}

	node.LastHeartbeat = timestamppb.New(time.Now())
	node.Status = NodeStatus_NODE_STATUS_ONLINE
	return nil
}

// GetNode retrieves a node by its ID
func (r *ServiceRegistry) GetNode(nodeID string) (*NodeInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return nil, errors.New("node not found")
	}

	// Check if node is alive
	if time.Since(node.LastHeartbeat.AsTime()) > r.heartbeatTimeout {
		return nil, errors.New("node is offline")
	}

	return node, nil
}

// GetAllNodes returns all nodes in the registry
func (r *ServiceRegistry) GetAllNodes() []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*NodeInfo, 0, len(r.nodes))
	for _, node := range r.nodes {
		// Check if node is alive
		if time.Since(node.LastHeartbeat.AsTime()) <= r.heartbeatTimeout {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetWorkerNodes returns all worker nodes in the registry
func (r *ServiceRegistry) GetWorkerNodes() []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*NodeInfo, 0)
	for _, node := range r.nodes {
		if node.Type == NodeType_WORKER_NODE && time.Since(node.LastHeartbeat.AsTime()) <= r.heartbeatTimeout {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// CleanupOfflineNodes removes nodes that have not sent a heartbeat within the timeout period
func (r *ServiceRegistry) CleanupOfflineNodes() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for id, node := range r.nodes {
		if time.Since(node.LastHeartbeat.AsTime()) > r.heartbeatTimeout {
			delete(r.nodes, id)
			count++
		}
	}

	return count
}
