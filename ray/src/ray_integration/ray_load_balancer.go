package ray_integration

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
)

type LoadBalanceStrategy string

const (
	StrategyRoundRobin    LoadBalanceStrategy = "round_robin"
	StrategyLeastConn    LoadBalanceStrategy = "least_conn"
	StrategyIPHash       LoadBalanceStrategy = "ip_hash"
	StrategyWeighted     LoadBalanceStrategy = "weighted"
	StrategyRandom       LoadBalanceStrategy = "random"
	StrategyAdaptive     LoadBalanceStrategy = "adaptive"
)

type LoadBalancerConfig struct {
	Strategy         LoadBalanceStrategy
	HealthCheckInterval time.Duration
	MaxRetries       int
	Timeout          time.Duration
	RetryDelay       time.Duration
	CircuitBreakerThreshold float64
}

type NodeConnection struct {
	NodeID          string
	ActiveConns     int
	TotalConns      int
	FailedConns     int
	SuccessRate     float64
	AvgResponseTime time.Duration
	LastUsed        time.Time
	HealthScore     float64
}

type RayLoadBalancer struct {
	cluster    *RayCluster
	config     *LoadBalancerConfig
	nodes      map[string]*NodeConnection
	algorithm  LoadBalanceAlgorithm
	circuitBreaker map[string]*CircuitBreaker
	mu         sync.RWMutex
	requestCount uint64
}

type LoadBalanceAlgorithm interface {
	SelectNode(*RayLoadBalancer) (string, error)
}

type RoundRobinAlgorithm struct {
	currentIndex int
	mu           sync.Mutex
}

type LeastConnAlgorithm struct{}

type WeightedAlgorithm struct {
	weights map[string]int
	mu     sync.RWMutex
}

type CircuitBreaker struct {
	NodeID           string
	Failures         int
	Threshold        float64
	State            CircuitBreakerState
	LastFailureTime  time.Time
	RecoveryTimeout  time.Duration
}

type CircuitBreakerState string

const (
	CircuitBreakerClosed CircuitBreakerState = "closed"
	CircuitBreakerOpen   CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

func NewRayLoadBalancer(cluster *RayCluster, config *LoadBalancerConfig) *RayLoadBalancer {
	lb := &RayLoadBalancer{
		cluster:    cluster,
		config:     config,
		nodes:      make(map[string]*NodeConnection),
		circuitBreaker: make(map[string]*CircuitBreaker),
	}

	switch config.Strategy {
	case StrategyRoundRobin:
		lb.algorithm = &RoundRobinAlgorithm{currentIndex: 0}
	case StrategyLeastConn:
		lb.algorithm = &LeastConnAlgorithm{}
	case StrategyWeighted:
		lb.algorithm = &WeightedAlgorithm{weights: make(map[string]int)}
	case StrategyAdaptive:
		lb.algorithm = &AdaptiveAlgorithm{}
	default:
		lb.algorithm = &RoundRobinAlgorithm{currentIndex: 0}
	}

	return lb
}

func (lb *RayLoadBalancer) RegisterNode(nodeID string, weight int) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if _, exists := lb.nodes[nodeID]; exists {
		return fmt.Errorf("node %s already registered", nodeID)
	}

	conn := &NodeConnection{
		NodeID:      nodeID,
		ActiveConns: 0,
		TotalConns:  0,
		HealthScore: 100.0,
		LastUsed:    time.Now(),
	}

	lb.nodes[nodeID] = conn

	if wa, ok := lb.algorithm.(*WeightedAlgorithm); ok {
		wa.mu.Lock()
		wa.weights[nodeID] = weight
		wa.mu.Unlock()
	}

	lb.circuitBreaker[nodeID] = &CircuitBreaker{
		NodeID:    nodeID,
		Threshold: lb.config.CircuitBreakerThreshold,
		State:    CircuitBreakerClosed,
	}

	log.Printf("Node %s registered with load balancer (weight: %d)", nodeID, weight)
	return nil
}

func (lb *RayLoadBalancer) UnregisterNode(nodeID string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if _, exists := lb.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	delete(lb.nodes, nodeID)

	if wa, ok := lb.algorithm.(*WeightedAlgorithm); ok {
		wa.mu.Lock()
		delete(wa.weights, nodeID)
		wa.mu.Unlock()
	}

	delete(lb.circuitBreaker, nodeID)

	log.Printf("Node %s unregistered from load balancer", nodeID)
	return nil
}

func (lb *RayLoadBalancer) SelectNode() (string, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.nodes) == 0 {
		return "", fmt.Errorf("no nodes available")
	}

	if lb.circuitBreaker == nil {
		return lb.algorithm.SelectNode(lb)
	}

	for attempts := 0; attempts < 3; attempts++ {
		nodeID, err := lb.algorithm.SelectNode(lb)
		if err != nil {
			return "", err
		}

		if cb, exists := lb.circuitBreaker[nodeID]; exists {
			if cb.State == CircuitBreakerOpen {
				if time.Since(cb.LastFailureTime) > cb.RecoveryTimeout {
					cb.State = CircuitBreakerHalfOpen
					log.Printf("Circuit breaker for node %s moved to half-open", nodeID)
					return nodeID, nil
				}
				continue
			}
		}

		return nodeID, nil
	}

	return "", fmt.Errorf("all nodes are unavailable")
}

func (rr *RoundRobinAlgorithm) SelectNode(lb *RayLoadBalancer) (string, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	nodes := lb.getAvailableNodes()
	if len(nodes) == 0 {
		return "", fmt.Errorf("no available nodes")
	}

	index := rr.currentIndex % len(nodes)
	rr.currentIndex++

	return nodes[index], nil
}

func (lc *LeastConnAlgorithm) SelectNode(lb *RayLoadBalancer) (string, error) {
	nodes := lb.getAvailableNodes()
	if len(nodes) == 0 {
		return "", fmt.Errorf("no available nodes")
	}

	minConns := math.MaxInt32
	selectedNode := ""

	for _, nodeID := range nodes {
		conn := lb.nodes[nodeID]
		if conn.ActiveConns < minConns {
			minConns = conn.ActiveConns
			selectedNode = nodeID
		}
	}

	return selectedNode, nil
}

func (wa *WeightedAlgorithm) SelectNode(lb *RayLoadBalancer) (string, error) {
	wa.mu.RLock()
	defer wa.mu.RUnlock()

	if len(wa.weights) == 0 {
		return "", fmt.Errorf("no weighted nodes")
	}

	var totalWeight int
	for _, weight := range wa.weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return "", fmt.Errorf("total weight is zero")
	}

	nodes := lb.getAvailableNodes()
	if len(nodes) == 0 {
		return "", fmt.Errorf("no available nodes")
	}

	selectedWeight := time.Now().UnixNano() % int64(totalWeight)
	currentWeight := 0

	for _, nodeID := range nodes {
		if weight, exists := wa.weights[nodeID]; exists {
			currentWeight += weight
			if int64(currentWeight) > selectedWeight {
				return nodeID, nil
			}
		}
	}

	return nodes[0], nil
}

type AdaptiveAlgorithm struct{}

func (aa *AdaptiveAlgorithm) SelectNode(lb *RayLoadBalancer) (string, error) {
	nodes := lb.getAvailableNodes()
	if len(nodes) == 0 {
		return "", fmt.Errorf("no available nodes")
	}

	bestScore := -1.0
	selectedNode := ""

	for _, nodeID := range nodes {
		conn := lb.nodes[nodeID]
		score := lb.calculateHealthScore(conn)
		if score > bestScore {
			bestScore = score
			selectedNode = nodeID
		}
	}

	return selectedNode, nil
}

func (lb *RayLoadBalancer) getAvailableNodes() []string {
	nodes := make([]string, 0)
	for nodeID := range lb.nodes {
		nodes = append(nodes, nodeID)
	}
	return nodes
}

func (lb *RayLoadBalancer) calculateHealthScore(conn *NodeConnection) float64 {
	score := 100.0

	score -= float64(conn.ActiveConns) * 0.5

	if conn.SuccessRate < 1.0 {
		score -= (1.0 - conn.SuccessRate) * 50
	}

	score -= conn.AvgResponseTime.Seconds() * 10

	if conn.Failures > 0 {
		score -= float64(conn.Failures) * 5
	}

	return math.Max(0, score)
}

func (lb *RayLoadBalancer) RecordConnection(nodeID string, success bool, responseTime time.Duration) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	conn, exists := lb.nodes[nodeID]
	if !exists {
		return
	}

	conn.TotalConns++
	conn.LastUsed = time.Now()

	if success {
		conn.SuccessRate = float64(conn.TotalConns-conn.Failures) / float64(conn.TotalConns)
		conn.AvgResponseTime = (conn.AvgResponseTime*time.Duration(conn.TotalConns-1) + responseTime) / time.Duration(conn.TotalConns)
		lb.recordCircuitBreakerSuccess(nodeID)
	} else {
		conn.Failures++
		lb.recordCircuitBreakerFailure(nodeID)
	}

	conn.HealthScore = lb.calculateHealthScore(conn)
}

func (lb *RayLoadBalancer) recordCircuitBreakerSuccess(nodeID string) {
	cb, exists := lb.circuitBreaker[nodeID]
	if !exists {
		return
	}

	if cb.State == CircuitBreakerHalfOpen {
		cb.Failures = 0
		cb.State = CircuitBreakerClosed
		log.Printf("Circuit breaker for node %s closed", nodeID)
	}
}

func (lb *RayLoadBalancer) recordCircuitBreakerFailure(nodeID string) {
	cb, exists := lb.circuitBreaker[nodeID]
	if !exists {
		return
	}

	cb.Failures++
	cb.LastFailureTime = time.Now()

	if cb.State == CircuitBreakerHalfOpen || float64(cb.Failures)/float64(cb.Threshold) >= 1.0 {
		cb.State = CircuitBreakerOpen
		log.Printf("Circuit breaker for node %s opened", nodeID)
	}
}

func (lb *RayLoadBalancer) IncrementActiveConns(nodeID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if conn, exists := lb.nodes[nodeID]; exists {
		conn.ActiveConns++
	}
}

func (lb *RayLoadBalancer) DecrementActiveConns(nodeID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if conn, exists := lb.nodes[nodeID]; exists {
		conn.ActiveConns = math.Max(0, float64(conn.ActiveConns-1))
	}
}

func (lb *RayLoadBalancer) GetNodeStats(nodeID string) (*structpb.Struct, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	conn, exists := lb.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	fields := make(map[string]*structpb.Value)
	fields["node_id"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: conn.NodeID},
	}
	fields["active_conns"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(conn.ActiveConns)},
	}
	fields["total_conns"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(conn.TotalConns)},
	}
	fields["failed_conns"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(conn.Failures)},
	}
	fields["success_rate"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: conn.SuccessRate * 100},
	}
	fields["avg_response_time_ms"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(conn.AvgResponseTime.Milliseconds())},
	}
	fields["health_score"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: conn.HealthScore},
	}
	fields["last_used"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: conn.LastUsed.Format(time.RFC3339)},
	}

	if cb, exists := lb.circuitBreaker[nodeID]; exists {
		cbFields := make(map[string]*structpb.Value)
		cbFields["state"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: string(cb.State)},
		}
		cbFields["failures"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(cb.Failures)},
		}
		fields["circuit_breaker"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: cbFields}},
		}
	}

	return &structpb.Struct{Fields: fields}, nil
}

func (lb *RayLoadBalancer) GetLoadBalancerStats() *LoadBalancerStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := &LoadBalancerStats{
		Strategy:     string(lb.config.Strategy),
		TotalNodes:   len(lb.nodes),
		RequestCount: lb.requestCount,
		NodeStats:    make(map[string]*NodeConnectionStats),
	}

	var totalConns, activeConns int
	var totalSuccessRate float64

	for nodeID, conn := range lb.nodes {
		stats.NodeStats[nodeID] = &NodeConnectionStats{
			ActiveConns:     conn.ActiveConns,
			TotalConns:      conn.TotalConns,
			SuccessRate:     conn.SuccessRate,
			AvgResponseTime: conn.AvgResponseTime,
			HealthScore:     conn.HealthScore,
		}
		totalConns += conn.TotalConns
		activeConns += conn.ActiveConns
		totalSuccessRate += conn.SuccessRate
	}

	if len(lb.nodes) > 0 {
		stats.AverageSuccessRate = totalSuccessRate / float64(len(lb.nodes))
	}

	stats.TotalConnections = totalConns
	stats.ActiveConnections = activeConns

	return stats
}

type LoadBalancerStats struct {
	Strategy           string
	TotalNodes         int
	RequestCount       uint64
	TotalConnections   int
	ActiveConnections  int
	AverageSuccessRate float64
	NodeStats          map[string]*NodeConnectionStats
}

type NodeConnectionStats struct {
	ActiveConns     int
	TotalConns      int
	SuccessRate     float64
	AvgResponseTime time.Duration
	HealthScore     float64
}

func (lb *RayLoadBalancer) SetStrategy(strategy LoadBalanceStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.config.Strategy = strategy

	switch strategy {
	case StrategyRoundRobin:
		lb.algorithm = &RoundRobinAlgorithm{currentIndex: 0}
	case StrategyLeastConn:
		lb.algorithm = &LeastConnAlgorithm{}
	case StrategyWeighted:
		lb.algorithm = &WeightedAlgorithm{weights: make(map[string]int)}
	case StrategyAdaptive:
		lb.algorithm = &AdaptiveAlgorithm{}
	}

	log.Printf("Load balancer strategy changed to %s", strategy)
}

func (lb *RayLoadBalancer) StartHealthCheck() {
	ticker := time.NewTicker(lb.config.HealthCheckInterval)
	go func() {
		for range ticker.C {
			lb.performHealthCheck()
		}
	}()
}

func (lb *RayLoadBalancer) performHealthCheck() {
	lb.mu.RLock()
	nodes := make([]string, 0, len(lb.nodes))
	for nodeID := range lb.nodes {
		nodes = append(nodes, nodeID)
	}
	lb.mu.RUnlock()

	for _, nodeID := range nodes {
		lb.mu.RLock()
		conn := lb.nodes[nodeID]
		lb.mu.RUnlock()

		healthScore := lb.calculateHealthScore(conn)

		if healthScore < 50 {
			log.Printf("Node %s health check failed (score: %.2f)", nodeID, healthScore)
			lb.recordCircuitBreakerFailure(nodeID)
		}
	}
}

func (lb *RayLoadBalancer) ResetNodeStats(nodeID string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	conn, exists := lb.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	conn.ActiveConns = 0
	conn.TotalConns = 0
	conn.Failures = 0
	conn.SuccessRate = 1.0
	conn.AvgResponseTime = 0
	conn.HealthScore = 100.0

	if cb, exists := lb.circuitBreaker[nodeID]; exists {
		cb.Failures = 0
		cb.State = CircuitBreakerClosed
	}

	log.Printf("Stats reset for node %s", nodeID)
	return nil
}
