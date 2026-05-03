package ray_integration

import (
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
)

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

type MetricValue struct {
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

type Metric struct {
	Name      string
	Type      MetricType
	Help      string
	Values    []*MetricValue
	Aggregates *MetricAggregates
}

type MetricAggregates struct {
	Count    int
	Sum      float64
	Min      float64
	Max      float64
	Avg      float64
	P50      float64
	P90      float64
	P99      float64
}

type RayMetricsCollector struct {
	cluster      *RayCluster
	metrics      map[string]*Metric
	eventLogs    []*EventLog
	performance  *ClusterPerformance
	mu           sync.RWMutex
	collectors   []MetricsCollector
}

type EventLog struct {
	Timestamp time.Time
	Level     string
	Source    string
	Message   string
	Data      map[string]interface{}
}

type ClusterPerformance struct {
	TasksPerSecond      float64
	AvgTaskDuration     time.Duration
	TotalTasksCompleted int
	TotalTasksFailed    int
	ClusterEfficiency   float64
	ThroughputHistory   []float64
}

type MetricsCollector interface {
	Collect() error
	GetName() string
}

func NewRayMetricsCollector(cluster *RayCluster) *RayMetricsCollector {
	return &RayMetricsCollector{
		cluster:     cluster,
		metrics:     make(map[string]*Metric),
		eventLogs:   make([]*EventLog, 0),
		performance: &ClusterPerformance{
			ThroughputHistory: make([]float64, 0),
		},
	}
}

func (mc *RayMetricsCollector) RegisterMetric(name string, metricType MetricType, help string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics[name] = &Metric{
		Name:      name,
		Type:      metricType,
		Help:      help,
		Values:    make([]*MetricValue, 0),
		Aggregates: &MetricAggregates{},
	}

	log.Printf("Registered metric: %s (type: %s)", name, metricType)
}

func (mc *RayMetricsCollector) RecordMetric(name string, value float64, labels map[string]string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metric, exists := mc.metrics[name]
	if !exists {
		return fmt.Errorf("metric %s not found", name)
	}

	metricValue := &MetricValue{
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}

	metric.Values = append(metric.Values, metricValue)
	mc.updateAggregates(metric)

	if len(metric.Values) > 1000 {
		metric.Values = metric.Values[1:]
	}

	return nil
}

func (mc *RayMetricsCollector) updateAggregates(metric *Metric) {
	if len(metric.Values) == 0 {
		return
	}

	metric.Aggregates.Count = len(metric.Values)
	metric.Aggregates.Sum = 0
	metric.Aggregates.Min = metric.Values[0].Value
	metric.Aggregates.Max = metric.Values[0].Value

	values := make([]float64, len(metric.Values))
	for i, v := range metric.Values {
		values[i] = v.Value
		metric.Aggregates.Sum += v.Value
		if v.Value < metric.Aggregates.Min {
			metric.Aggregates.Min = v.Value
		}
		if v.Value > metric.Aggregates.Max {
			metric.Aggregates.Max = v.Value
		}
	}

	metric.Aggregates.Avg = metric.Aggregates.Sum / float64(metric.Aggregates.Count)

	n := len(values)
	if n > 0 {
		metric.Aggregates.P50 = values[n/2]
		metric.Aggregates.P90 = values[int(float64(n)*0.9)]
		metric.Aggregates.P99 = values[int(float64(n)*0.99)]
	}
}

func (mc *RayMetricsCollector) GetMetric(name string) (*Metric, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metric, exists := mc.metrics[name]
	if !exists {
		return nil, fmt.Errorf("metric %s not found", name)
	}
	return metric, nil
}

func (mc *RayMetricsCollector) GetAllMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*Metric)
	for name, metric := range mc.metrics {
		result[name] = metric
	}
	return result
}

func (mc *RayMetricsCollector) LogEvent(level, source, message string, data map[string]interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	event := &EventLog{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
		Data:      data,
	}

	mc.eventLogs = append(mc.eventLogs, event)

	if len(mc.eventLogs) > 10000 {
		mc.eventLogs = mc.eventLogs[1:]
	}

	log.Printf("[%s] %s: %s", level, source, message)
}

func (mc *RayMetricsCollector) GetEventLogs(limit int) []*EventLog {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if limit <= 0 || limit > len(mc.eventLogs) {
		limit = len(mc.eventLogs)
	}

	logs := make([]*EventLog, limit)
	copy(logs, mc.eventLogs[len(mc.eventLogs)-limit:])
	return logs
}

func (mc *RayMetricsCollector) GetClusterMetrics() *ClusterMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := &ClusterMetrics{
		Timestamp: time.Now(),
		Metrics:   make(map[string]*Metric),
	}

	for name, metric := range mc.metrics {
		metrics.Metrics[name] = metric
	}

	nodes := mc.cluster.GetNodes()
	metrics.TotalNodes = len(nodes)
	metrics.OnlineNodes = 0

	for _, node := range nodes {
		if node.Status == "online" {
			metrics.OnlineNodes++
		}
	}

	metrics.Performance = mc.performance

	return metrics
}

type ClusterMetrics struct {
	Timestamp     time.Time
	TotalNodes    int
	OnlineNodes   int
	Metrics       map[string]*Metric
	Performance   *ClusterPerformance
}

func (mc *RayMetricsCollector) UpdatePerformance(tasksCompleted, tasksFailed int, avgDuration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.performance.TotalTasksCompleted += tasksCompleted
	mc.performance.TotalTasksFailed += tasksFailed
	mc.performance.AvgTaskDuration = avgDuration

	if mc.performance.TotalTasksCompleted+mc.performance.TotalTasksFailed > 0 {
		mc.performance.ClusterEfficiency = float64(mc.performance.TotalTasksCompleted) / 
			float64(mc.performance.TotalTasksCompleted+mc.performance.TotalTasksFailed)
	}

	throughput := float64(tasksCompleted) / 60.0
	mc.performance.TasksPerSecond = throughput / 60.0
	mc.performance.ThroughputHistory = append(mc.performance.ThroughputHistory, throughput)

	if len(mc.performance.ThroughputHistory) > 60 {
		mc.performance.ThroughputHistory = mc.performance.ThroughputHistory[1:]
	}
}

func (mc *RayMetricsCollector) CalculateClusterHealth() *ClusterHealth {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	health := &ClusterHealth{
		Timestamp: time.Now(),
		Score:     100.0,
		Factors:   make(map[string]float64),
	}

	nodes := mc.cluster.GetNodes()
	totalNodes := len(nodes)
	if totalNodes == 0 {
		health.Status = "unknown"
		health.Score = 0
		return health
	}

	onlineNodes := 0
	for _, node := range nodes {
		if node.Status == "online" {
			onlineNodes++
		}
	}

	nodeHealthFactor := float64(onlineNodes) / float64(totalNodes) * 100
	health.Factors["node_availability"] = nodeHealthFactor
	health.Score *= nodeHealthFactor / 100

	if mc.performance.ClusterEfficiency > 0 {
		health.Factors["task_success_rate"] = mc.performance.ClusterEfficiency * 100
		health.Score *= mc.performance.ClusterEfficiency
	}

	if health.Score >= 90 {
		health.Status = "healthy"
	} else if health.Score >= 70 {
		health.Status = "degraded"
	} else if health.Score >= 50 {
		health.Status = "unhealthy"
	} else {
		health.Status = "critical"
	}

	return health
}

type ClusterHealth struct {
	Timestamp time.Time
	Status    string
	Score     float64
	Factors   map[string]float64
}

func (mc *RayMetricsCollector) ExportMetrics() (*structpb.Struct, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := mc.GetClusterMetrics()
	health := mc.CalculateClusterHealth()

	fields := make(map[string]*structpb.Value)

	fields["timestamp"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: metrics.Timestamp.Format(time.RFC3339)},
	}
	fields["total_nodes"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(metrics.TotalNodes)},
	}
	fields["online_nodes"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(metrics.OnlineNodes)},
	}

	metricsFields := make(map[string]*structpb.Value)
	for name, metric := range metrics.Metrics {
		if metric.Aggregates != nil {
			metricFields := make(map[string]*structpb.Value)
			metricFields["count"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: float64(metric.Aggregates.Count)},
			}
			metricFields["sum"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.Sum},
			}
			metricFields["avg"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.Avg},
			}
			metricFields["min"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.Min},
			}
			metricFields["max"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.Max},
			}
			metricFields["p50"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.P50},
			}
			metricFields["p90"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.P90},
			}
			metricFields["p99"] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{NumberValue: metric.Aggregates.P99},
			}
			metricsFields[name] = &structpb.Value{
				Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: metricFields}},
			}
		}
	}
	fields["metrics"] = &structpb.Value{
		Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: metricsFields}},
	}

	healthFields := make(map[string]*structpb.Value)
	healthFields["status"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: health.Status},
	}
	healthFields["score"] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: health.Score},
	}
	factorsFields := make(map[string]*structpb.Value)
	for name, value := range health.Factors {
		factorsFields[name] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: value},
		}
	}
	healthFields["factors"] = &structpb.Value{
		Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: factorsFields}},
	}
	fields["health"] = &structpb.Value{
		Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: healthFields}},
	}

	if metrics.Performance != nil {
		perfFields := make(map[string]*structpb.Value)
		perfFields["tasks_completed"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(metrics.Performance.TotalTasksCompleted)},
		}
		perfFields["tasks_failed"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: float64(metrics.Performance.TotalTasksFailed)},
		}
		perfFields["efficiency"] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{NumberValue: metrics.Performance.ClusterEfficiency},
		}
		fields["performance"] = &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: perfFields}},
		}
	}

	return &structpb.Struct{Fields: fields}, nil
}

func (mc *RayMetricsCollector) RegisterCollector(collector MetricsCollector) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.collectors = append(mc.collectors, collector)
	log.Printf("Registered metrics collector: %s", collector.GetName())
}

func (mc *RayMetricsCollector) StartCollection(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			mc.collect()
		}
	}()
}

func (mc *RayMetricsCollector) collect() {
	mc.mu.RLock()
	collectors := make([]MetricsCollector, len(mc.collectors))
	copy(collectors, mc.collectors)
	mc.mu.RUnlock()

	for _, collector := range collectors {
		if err := collector.Collect(); err != nil {
			mc.LogEvent("error", collector.GetName(), fmt.Sprintf("Collection failed: %v", err), nil)
		}
	}

	mc.updateClusterMetrics()
}

func (mc *RayMetricsCollector) updateClusterMetrics() {
	nodes := mc.cluster.GetNodes()

	totalCPU := 0.0
	totalMemory := 0.0
	usedCPU := 0.0
	usedMemory := 0.0

	for _, node := range nodes {
		if cpu, ok := node.Resources["CPU"]; ok {
			totalCPU += cpu
		}
		if mem, ok := node.Resources["memory"]; ok {
			totalMemory += mem
		}
	}

	mc.RecordMetric("cluster_nodes_total", float64(len(nodes)), nil)
	mc.RecordMetric("cluster_cpu_total", totalCPU, nil)
	mc.RecordMetric("cluster_memory_total", totalMemory, nil)
	mc.RecordMetric("cluster_cpu_usage", usedCPU, nil)
	mc.RecordMetric("cluster_memory_usage", usedMemory, nil)
}

func (mc *RayMetricsCollector) GetMetricsSummary() *MetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	summary := &MetricsSummary{
		TotalMetrics: len(mc.metrics),
		MetricNames: make([]string, 0),
	}

	for name := range mc.metrics {
		summary.MetricNames = append(summary.MetricNames, name)
	}

	summary.EventLogsCount = len(mc.eventLogs)
	summary.CollectorsCount = len(mc.collectors)

	return summary
}

type MetricsSummary struct {
	TotalMetrics     int
	MetricNames      []string
	EventLogsCount   int
	CollectorsCount  int
}

type NodeMetricsCollector struct {
	nodeID string
}

func NewNodeMetricsCollector(nodeID string) *NodeMetricsCollector {
	return &NodeMetricsCollector{nodeID: nodeID}
}

func (nmc *NodeMetricsCollector) GetName() string {
	return fmt.Sprintf("node_%s", nmc.nodeID)
}

func (nmc *NodeMetricsCollector) Collect() error {
	return nil
}
