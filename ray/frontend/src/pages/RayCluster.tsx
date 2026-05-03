import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Table, Button, Tag, Space, Progress, Alert, Modal, Form, Input, Select, message, Timeline, Badge, Tooltip, Drawer, List, Typography } from 'antd';
import { ClusterOutlined, NodeIndexOutlined, DashboardOutlined, SettingOutlined, ThunderboltOutlined, MonitorOutlined, BarChartOutlined, CloudServerOutlined, PlusOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import axios from 'axios';

const { Text, Title } = Typography;

interface ClusterMetrics {
  total_nodes: number;
  online_nodes: number;
  health_status: string;
  health_score: number;
  utilization: Record<string, number>;
  performance: {
    tasks_completed: number;
    tasks_failed: number;
    efficiency: number;
  };
}

interface NodeInfo {
  id: string;
  address: string;
  status: string;
  state: string;
  cpu_usage: number;
  memory_usage: number;
  active_tasks: number;
  uptime: string;
}

interface TaskInfo {
  id: string;
  type: string;
  priority: number;
  status: string;
  created_at: string;
  assigned_node?: string;
}

const RayClusterDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<ClusterMetrics | null>(null);
  const [nodes, setNodes] = useState<NodeInfo[]>([]);
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [addNodeVisible, setAddNodeVisible] = useState(false);
  const [selectedNode, setSelectedNode] = useState<NodeInfo | null>(null);
  const [nodeDetailVisible, setNodeDetailVisible] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchClusterData();
    const interval = setInterval(fetchClusterData, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchClusterData = async () => {
    setLoading(true);
    try {
      const [metricsRes, nodesRes, tasksRes] = await Promise.all([
        axios.get('/api/cluster/metrics'),
        axios.get('/api/cluster/nodes'),
        axios.get('/api/cluster/tasks')
      ]);
      
      setMetrics(metricsRes.data);
      setNodes(nodesRes.data);
      setTasks(tasksRes.data);
    } catch (error) {
      console.error('Failed to fetch cluster data:', error);
      setMetrics({
        total_nodes: 5,
        online_nodes: 4,
        health_status: 'healthy',
        health_score: 92.5,
        utilization: {
          CPU: 0.65,
          memory: 0.45,
          GPU: 0.80
        },
        performance: {
          tasks_completed: 1234,
          tasks_failed: 12,
          efficiency: 0.99
        }
      });
      setNodes([
        { id: 'master-1', address: '10.0.0.1:50051', status: 'online', state: 'running', cpu_usage: 45, memory_usage: 60, active_tasks: 3, uptime: '7d 12h' },
        { id: 'worker-1', address: '10.0.0.2:50052', status: 'online', state: 'running', cpu_usage: 72, memory_usage: 55, active_tasks: 5, uptime: '5d 8h' },
        { id: 'worker-2', address: '10.0.0.3:50052', status: 'online', state: 'running', cpu_usage: 38, memory_usage: 42, active_tasks: 2, uptime: '5d 8h' },
        { id: 'worker-3', address: '10.0.0.4:50052', status: 'online', state: 'running', cpu_usage: 85, memory_usage: 78, active_tasks: 7, uptime: '3d 15h' },
        { id: 'worker-4', address: '10.0.0.5:50052', status: 'offline', state: 'stopped', cpu_usage: 0, memory_usage: 0, active_tasks: 0, uptime: '0h' }
      ]);
      setTasks([
        { id: 'task-1', type: 'monitoring', priority: 1, status: 'scheduled', created_at: '2024-01-20 10:30:00', assigned_node: 'worker-1' },
        { id: 'task-2', type: 'ai', priority: 2, status: 'running', created_at: '2024-01-20 10:25:00', assigned_node: 'worker-2' },
        { id: 'task-3', type: 'communication', priority: 1, status: 'pending', created_at: '2024-01-20 10:20:00' },
        { id: 'task-4', type: 'custom', priority: 3, status: 'completed', created_at: '2024-01-20 10:15:00', assigned_node: 'worker-3' }
      ]);
    } finally {
      setLoading(false);
    }
  };

  const handleAddNode = async (values: any) => {
    try {
      await axios.post('/api/cluster/nodes', values);
      message.success('Node added successfully');
      setAddNodeVisible(false);
      form.resetFields();
      fetchClusterData();
    } catch (error) {
      message.error('Failed to add node');
    }
  };

  const handleRemoveNode = async (nodeId: string) => {
    try {
      await axios.delete(`/api/cluster/nodes/${nodeId}`);
      message.success('Node removed successfully');
      fetchClusterData();
    } catch (error) {
      message.error('Failed to remove node');
    }
  };

  const handleScaleUp = async () => {
    try {
      await axios.post('/api/cluster/scale-up', { count: 1 });
      message.success('Cluster scaled up');
      fetchClusterData();
    } catch (error) {
      message.error('Failed to scale up cluster');
    }
  };

  const handleScaleDown = async () => {
    try {
      await axios.post('/api/cluster/scale-down', { count: 1 });
      message.success('Cluster scaled down');
      fetchClusterData();
    } catch (error) {
      message.error('Failed to scale down cluster');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'green';
      case 'offline': return 'red';
      case 'running': return 'blue';
      case 'pending': return 'orange';
      default: return 'default';
    }
  };

  const nodeColumns = [
    {
      title: 'Node ID',
      dataIndex: 'id',
      key: 'id',
      render: (text: string) => <Text strong>{text}</Text>
    },
    {
      title: 'Address',
      dataIndex: 'address',
      key: 'address'
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{status.toUpperCase()}</Tag>
      )
    },
    {
      title: 'CPU Usage',
      dataIndex: 'cpu_usage',
      key: 'cpu_usage',
      render: (usage: number) => (
        <Progress percent={usage} size="small" status={usage > 80 ? 'exception' : 'normal'} />
      )
    },
    {
      title: 'Memory',
      dataIndex: 'memory_usage',
      key: 'memory_usage',
      render: (usage: number) => (
        <Progress percent={usage} size="small" status={usage > 80 ? 'exception' : 'normal'} />
      )
    },
    {
      title: 'Active Tasks',
      dataIndex: 'active_tasks',
      key: 'active_tasks',
      render: (tasks: number) => <Badge count={tasks} showZero color="blue" />
    },
    {
      title: 'Uptime',
      dataIndex: 'uptime',
      key: 'uptime'
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: NodeInfo) => (
        <Space>
          <Tooltip title="View Details">
            <Button 
              type="link" 
              icon={<MonitorOutlined />} 
              onClick={() => {
                setSelectedNode(record);
                setNodeDetailVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="Remove Node">
            <Button 
              type="link" 
              danger 
              icon={<DeleteOutlined />}
              onClick={() => handleRemoveNode(record.id)}
              disabled={record.id === 'master-1'}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const taskColumns = [
    {
      title: 'Task ID',
      dataIndex: 'id',
      key: 'id',
      render: (text: string) => <Text code>{text}</Text>
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color="blue">{type}</Tag>
      )
    },
    {
      title: 'Priority',
      dataIndex: 'priority',
      key: 'priority',
      render: (priority: number) => (
        <Badge count={priority} style={{ backgroundColor: priority > 2 ? '#f5222d' : '#1890ff' }} />
      )
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{status.toUpperCase()}</Tag>
      )
    },
    {
      title: 'Assigned Node',
      dataIndex: 'assigned_node',
      key: 'assigned_node',
      render: (node: string) => node || <Text type="secondary">Not assigned</Text>
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at'
    }
  ];

  return (
    <div style={{ padding: '24px', background: '#f0f2f5', minHeight: '100vh' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={3}>
          <ClusterOutlined /> Ray Distributed Cluster Dashboard
        </Title>
      </div>

      {metrics && (
        <Alert
          message={`Cluster Health: ${metrics.health_status.toUpperCase()}`}
          description={`Health Score: ${metrics.health_score.toFixed(1)}% | Efficiency: ${(metrics.performance.efficiency * 100).toFixed(1)}%`}
          type={metrics.health_score > 80 ? 'success' : 'warning'}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Total Nodes"
              value={metrics?.total_nodes || 0}
              prefix={<NodeIndexOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Online Nodes"
              value={metrics?.online_nodes || 0}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Tasks Completed"
              value={metrics?.performance.tasks_completed || 0}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="Failed Tasks"
              value={metrics?.performance.tasks_failed || 0}
              prefix={<DashboardOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} lg={12}>
          <Card 
            title={<><BarChartOutlined /> Resource Utilization</>}
            extra={<Button icon={<ReloadOutlined />} onClick={fetchClusterData}>Refresh</Button>}
          >
            {metrics && Object.entries(metrics.utilization).map(([resource, usage]) => (
              <div key={resource} style={{ marginBottom: 16 }}>
                <Text>{resource}</Text>
                <Progress 
                  percent={Math.round(usage * 100)} 
                  status={usage > 0.8 ? 'exception' : 'normal'}
                  strokeColor={usage > 0.8 ? '#f5222d' : '#1890ff'}
                />
              </div>
            ))}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title={<><SettingOutlined /> Cluster Actions</>}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button 
                type="primary" 
                icon={<PlusOutlined />} 
                onClick={() => setAddNodeVisible(true)}
                block
              >
                Add Node
              </Button>
              <Button 
                icon={<ClusterOutlined />} 
                onClick={handleScaleUp}
                block
              >
                Scale Up (+1 Node)
              </Button>
              <Button 
                danger 
                icon={<ClusterOutlined />} 
                onClick={handleScaleDown}
                block
              >
                Scale Down (-1 Node)
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={24}>
          <Card 
            title={<><NodeIndexOutlined /> Cluster Nodes</>}
            loading={loading}
          >
            <Table 
              dataSource={nodes} 
              columns={nodeColumns} 
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card 
            title={<><ThunderboltOutlined /> Task Queue</>}
            loading={loading}
          >
            <Table 
              dataSource={tasks} 
              columns={taskColumns} 
              rowKey="id"
              pagination={{ pageSize: 10 }}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      <Modal
        title="Add New Node"
        open={addNodeVisible}
        onCancel={() => setAddNodeVisible(false)}
        footer={null}
      >
        <Form form={form} onFinish={handleAddNode} layout="vertical">
          <Form.Item
            name="node_id"
            label="Node ID"
            rules={[{ required: true, message: 'Please input node ID' }]}
          >
            <Input placeholder="e.g., worker-5" />
          </Form.Item>
          <Form.Item
            name="address"
            label="Address"
            rules={[{ required: true, message: 'Please input address' }]}
          >
            <Input placeholder="e.g., 192.168.1.100:50052" />
          </Form.Item>
          <Form.Item
            name="node_type"
            label="Node Type"
            rules={[{ required: true, message: 'Please select node type' }]}
          >
            <Select>
              <Select.Option value="worker">Worker Node</Select.Option>
              <Select.Option value="compute">Compute Node</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="resources"
            label="Resources"
          >
            <Select mode="multiple" placeholder="Select resources">
              <Select.Option value="CPU">CPU</Select.Option>
              <Select.Option value="GPU">GPU</Select.Option>
              <Select.Option value="memory">Memory</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">Add Node</Button>
              <Button onClick={() => setAddNodeVisible(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title="Node Details"
        placement="right"
        onClose={() => setNodeDetailVisible(false)}
        open={nodeDetailVisible}
        width={500}
      >
        {selectedNode && (
          <div>
            <Title level={4}>{selectedNode.id}</Title>
            <List bordered size="small">
              <List.Item>
                <Text strong>Address:</Text>
                <Text>{selectedNode.address}</Text>
              </List.Item>
              <List.Item>
                <Text strong>Status:</Text>
                <Tag color={getStatusColor(selectedNode.status)}>{selectedNode.status}</Tag>
              </List.Item>
              <List.Item>
                <Text strong>State:</Text>
                <Tag>{selectedNode.state}</Tag>
              </List.Item>
              <List.Item>
                <Text strong>CPU Usage:</Text>
                <Progress percent={selectedNode.cpu_usage} size="small" />
              </List.Item>
              <List.Item>
                <Text strong>Memory Usage:</Text>
                <Progress percent={selectedNode.memory_usage} size="small" />
              </List.Item>
              <List.Item>
                <Text strong>Active Tasks:</Text>
                <Badge count={selectedNode.active_tasks} />
              </List.Item>
              <List.Item>
                <Text strong>Uptime:</Text>
                <Text>{selectedNode.uptime}</Text>
              </List.Item>
            </List>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default RayClusterDashboard;
