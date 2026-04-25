package common

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewNodeInfo creates a new NodeInfo instance
func NewNodeInfo(id string, nodeType NodeType, address string, port int, capabilities []string) *NodeInfo {
	now := time.Now()
	timestampNow := timestamppb.New(now)
	return &NodeInfo{
		Id:          id,
		Type:        nodeType,
		Address:     address,
		Port:        int32(port),
		Status:      NodeStatus_NODE_STATUS_ONLINE,
		Capabilities: capabilities,
		Load:        0.0,
		LastHeartbeat: timestampNow,
		CreatedAt:    timestampNow,
	}
}
