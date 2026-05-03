package ray_integration

import (
	"fmt"
	"math"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
)

type ResourceType string

const (
	ResourceCPU     ResourceType = "CPU"
	ResourceMemory  ResourceType = "memory"
	ResourceGPU     ResourceType = "GPU"
	ResourceStorage ResourceType = "storage"
	ResourceNetwork ResourceType = "network"
)

type Resource struct {
	Type  ResourceType
	Total float64
	Used  float64
}

type ResourceAllocation struct {
	NodeID     string
	Resources  map[ResourceType]float64
	TaskID     string
	AllocatedAt time.Time
}

type RayResourceManager struct {
	cluster       *RayCluster
	resources    map[string]map[ResourceType]*Resource
	allocations  map[string]*ResourceAllocation
	mu           sync.RWMutex
	totalResources map[ResourceType]float64
}

func NewRayResourceManager(cluster *RayCluster) *RayResourceManager {
	return &RayResourceManager{
		cluster:       cluster,
		resources:     make(map[string]map[ResourceType]*Resource),
		allocations:   make(map[string]*ResourceAllocation),
		totalResources: make(map[ResourceType]float64),
	}
}

func (rm *RayResourceManager) InitializeNode(nodeID string, resources map[ResourceType]float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	nodeResources := make(map[ResourceType]*Resource)
	for resType, amount := range resources {
		nodeResources[resType] = &Resource{
			Type:  resType,
			Total: amount,
			Used:  0,
		}
		rm.totalResources[resType] += amount
	}
	rm.resources[nodeID] = nodeResources
}

func (rm *RayResourceManager) AllocateResources(nodeID, taskID string, required map[ResourceType]float64) (*ResourceAllocation, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	nodeResources, exists := rm.resources[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	allocation := &ResourceAllocation{
		NodeID:     nodeID,
		Resources:  make(map[ResourceType]float64),
		TaskID:     taskID,
		AllocatedAt: time.Now(),
	}

	for resType, amount := range required {
		resource, exists := nodeResources[resType]
		if !exists {
			return nil, fmt.Errorf("resource %s not available on node %s", resType, nodeID)
		}

		if resource.Total-resource.Used < amount {
			return nil, fmt.Errorf("insufficient %s on node %s: need %.2f, have %.2f", 
				resType, nodeID, amount, resource.Total-resource.Used)
		}

		resource.Used += amount
		allocation.Resources[resType] = amount
	}

	rm.allocations[taskID] = allocation
	return allocation, nil
}

func (rm *RayResourceManager) ReleaseResources(taskID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	allocation, exists := rm.allocations[taskID]
	if !exists {
		return fmt.Errorf("allocation for task %s not found", taskID)
	}

	nodeResources, exists := rm.resources[allocation.NodeID]
	if !exists {
		return fmt.Errorf("node %s not found", allocation.NodeID)
	}

	for resType, amount := range allocation.Resources {
		if resource, exists := nodeResources[resType]; exists {
			resource.Used = math.Max(0, resource.Used-amount)
		}
	}

	delete(rm.allocations, taskID)
	return nil
}

func (rm *RayResourceManager) GetAvailableResources(nodeID string) map[ResourceType]float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	available := make(map[ResourceType]float64)
	nodeResources, exists := rm.resources[nodeID]
	if !exists {
		return available
	}

	for resType, resource := range nodeResources {
		available[resType] = resource.Total - resource.Used
	}
	return available
}

func (rm *RayResourceManager) GetTotalResources() map[ResourceType]float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	total := make(map[ResourceType]float64)
	for resType, amount := range rm.totalResources {
		total[resType] = amount
	}
	return total
}

func (rm *RayResourceManager) GetClusterUtilization() map[ResourceType]float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	utilization := make(map[ResourceType]float64)
	for resType, total := range rm.totalResources {
		if total > 0 {
			var used float64
			for _, nodeResources := range rm.resources {
				if resource, exists := nodeResources[resType]; exists {
					used += resource.Used
				}
			}
			utilization[resType] = used / total
		}
	}
	return utilization
}

func (rm *RayResourceManager) FindBestNode(required map[ResourceType]float64) (string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	bestNode := ""
	bestScore := -1.0

	for nodeID, nodeResources := range rm.resources {
		score := rm.calculateNodeScore(nodeResources, required)
		if score > bestScore {
			bestScore = score
			bestNode = nodeID
		}
	}

	if bestNode == "" {
		return "", fmt.Errorf("no suitable node found")
	}
	return bestNode, nil
}

func (rm *RayResourceManager) calculateNodeScore(nodeResources map[ResourceType]*Resource, required map[ResourceType]float64) float64 {
	score := 0.0
	for resType, amount := range required {
		if resource, exists := nodeResources[resType]; exists {
			available := resource.Total - resource.Used
			if available >= amount {
				margin := available - amount
				score += margin / resource.Total
			} else {
				return -1.0
			}
		} else {
			return -1.0
		}
	}
	return score
}

func (rm *RayResourceManager) GetResourceStats() *ResourceStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := &ResourceStats{
		TotalNodes:    len(rm.resources),
		TotalAllocations: len(rm.allocations),
		NodeStats:    make(map[string]*NodeResourceStats),
	}

	for nodeID, nodeResources := range rm.resources {
		nodeStat := &NodeResourceStats{
			Resources: make(map[string]*ResourceStat),
		}

		for resType, resource := range nodeResources {
			nodeStat.Resources[string(resType)] = &ResourceStat{
				Total:      resource.Total,
				Used:       resource.Used,
				Available:  resource.Total - resource.Used,
				Utilization: (resource.Used / resource.Total) * 100,
			}
		}

		stats.NodeStats[nodeID] = nodeStat
	}

	return stats
}

type ResourceStats struct {
	TotalNodes       int
	TotalAllocations int
	NodeStats        map[string]*NodeResourceStats
}

type NodeResourceStats struct {
	Resources map[string]*ResourceStat
}

type ResourceStat struct {
	Total        float64
	Used         float64
	Available    float64
	Utilization  float64
}

func (rm *RayResourceManager) AutoScale(demand map[ResourceType]float64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	available := rm.GetTotalResources()
	
	for resType, required := range demand {
		if available[resType] < required {
			return fmt.Errorf("insufficient %s: need %.2f, available %.2f", 
				resType, required, available[resType])
		}
	}

	return nil
}

func (rm *RayResourceManager) GetAllocationInfo(taskID string) (*structpb.Struct, error) {
	rm.mu.RLock()
	allocation, exists := rm.allocations[taskID]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("allocation not found")
	}

	fields := make(map[string]*structpb.Value)
	fields["node_id"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: allocation.NodeID},
	}
	fields["task_id"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: allocation.TaskID},
	}
	fields["allocated_at"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: allocation.AllocatedAt.Format(time.RFC3339)},
	}

	resources := make(map[string]*structpb.Value)
	for resType, amount := range allocation.Resources {
		resources[string(resType)] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: amount},
		}
	}
	fields["resources"] = &structpb.Value{
		Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: resources}},
	}

	return &structpb.Struct{Fields: fields}, nil
}
