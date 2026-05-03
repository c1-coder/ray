import React, { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Progress, Table, Typography, Space, Tag } from 'antd'
import { ArrowUpOutlined, ArrowDownOutlined, CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined } from '@ant-design/icons'

const { Title } = Typography

interface Task {
  id: string
  type: string
  status: string
  priority: string
  assignedTo: string
  createdAt: string
}

interface Node {
  id: string
  type: string
  address: string
  status: string
  load: number
  capabilities: string[]
}

const Dashboard: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([])
  const [nodes, setNodes] = useState<Node[]>([])
  const [stats, setStats] = useState({
    totalTasks: 0,
    completedTasks: 0,
    pendingTasks: 0,
    totalNodes: 0,
    onlineNodes: 0,
    offlineNodes: 0
  })

  useEffect(() => {
    // Mock data for dashboard
    setTasks([
      { id: '1', type: 'Monitoring', status: 'completed', priority: 'high', assignedTo: 'node-1', createdAt: '2024-01-01 10:00:00' },
      { id: '2', type: 'AI', status: 'running', priority: 'medium', assignedTo: 'node-2', createdAt: '2024-01-01 11:00:00' },
      { id: '3', type: 'Communication', status: 'pending', priority: 'low', assignedTo: '', createdAt: '2024-01-01 12:00:00' },
      { id: '4', type: 'Monitoring', status: 'completed', priority: 'medium', assignedTo: 'node-3', createdAt: '2024-01-01 09:00:00' },
      { id: '5', type: 'AI', status: 'pending', priority: 'high', assignedTo: '', createdAt: '2024-01-01 13:00:00' },
    ])

    setNodes([
      { id: 'node-1', type: 'worker', address: '192.168.1.101:50052', status: 'online', load: 45, capabilities: ['task_execution', 'monitoring'] },
      { id: 'node-2', type: 'worker', address: '192.168.1.102:50052', status: 'online', load: 65, capabilities: ['task_execution', 'ai_assistance'] },
      { id: 'node-3', type: 'worker', address: '192.168.1.103:50052', status: 'online', load: 25, capabilities: ['task_execution', 'communication'] },
      { id: 'node-4', type: 'worker', address: '192.168.1.104:50052', status: 'offline', load: 0, capabilities: ['task_execution'] },
    ])

    setStats({
      totalTasks: 150,
      completedTasks: 120,
      pendingTasks: 30,
      totalNodes: 4,
      onlineNodes: 3,
      offlineNodes: 1
    })
  }, [])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'green'
      case 'running': return 'blue'
      case 'pending': return 'orange'
      case 'failed': return 'red'
      case 'online': return 'green'
      case 'offline': return 'red'
      default: return 'default'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircleOutlined />
      case 'running': return <ClockCircleOutlined />
      case 'pending': return <ClockCircleOutlined />
      case 'failed': return <CloseCircleOutlined />
      default: return null
    }
  }

  const taskColumns = [
    { title: 'Task ID', dataIndex: 'id', key: 'id' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    { 
      title: 'Status', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {getStatusIcon(status)} {status}
        </Tag>
      )
    },
    { 
      title: 'Priority', 
      dataIndex: 'priority', 
      key: 'priority',
      render: (priority: string) => (
        <Tag color={priority === 'high' ? 'red' : priority === 'medium' ? 'orange' : 'green'}>
          {priority}
        </Tag>
      )
    },
    { title: 'Assigned To', dataIndex: 'assignedTo', key: 'assignedTo' },
    { title: 'Created At', dataIndex: 'createdAt', key: 'createdAt' },
  ]

  const nodeColumns = [
    { title: 'Node ID', dataIndex: 'id', key: 'id' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    { title: 'Address', dataIndex: 'address', key: 'address' },
    { 
      title: 'Status', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status}
        </Tag>
      )
    },
    { 
      title: 'Load', 
      dataIndex: 'load', 
      key: 'load',
      render: (load: number) => (
        <Progress percent={load} size="small" status={load > 80 ? 'exception' : 'normal'} />
      )
    },
    { 
      title: 'Capabilities', 
      dataIndex: 'capabilities', 
      key: 'capabilities',
      render: (capabilities: string[]) => (
        <Space direction="vertical">
          {capabilities.map((capability, index) => (
            <Tag key={index}>{capability}</Tag>
          ))}
        </Space>
      )
    },
  ]

  return (
    <div>
      <Title level={2}>Dashboard</Title>
      
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Total Tasks" 
              value={stats.totalTasks}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Completed Tasks" 
              value={stats.completedTasks}
              valueStyle={{ color: '#52c41a' }}
              prefix={<ArrowUpOutlined />}
              suffix={`${Math.round((stats.completedTasks / stats.totalTasks) * 100)}%`}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Pending Tasks" 
              value={stats.pendingTasks}
              valueStyle={{ color: '#faad14' }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Task Success Rate" 
              value={Math.round((stats.completedTasks / stats.totalTasks) * 100)}
              valueStyle={{ color: '#52c41a' }}
              suffix="%"
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Total Nodes" 
              value={stats.totalNodes}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Online Nodes" 
              value={stats.onlineNodes}
              valueStyle={{ color: '#52c41a' }}
              prefix={<ArrowUpOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Offline Nodes" 
              value={stats.offlineNodes}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<ArrowDownOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="Node Availability" 
              value={Math.round((stats.onlineNodes / stats.totalNodes) * 100)}
              valueStyle={{ color: '#52c41a' }}
              suffix="%"
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Card title="Recent Tasks" bordered={false}>
            <Table 
              columns={taskColumns} 
              dataSource={tasks} 
              rowKey="id"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="Node Status" bordered={false}>
            <Table 
              columns={nodeColumns} 
              dataSource={nodes} 
              rowKey="id"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard
