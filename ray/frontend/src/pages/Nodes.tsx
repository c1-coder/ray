import React, { useState, useEffect } from 'react'
import { Card, Table, Button, Modal, Form, Select, Input, Tag, Space, Typography, message, Progress } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, PoweroffOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'

const { Title } = Typography
const { Option } = Select

interface Node {
  id: string
  type: string
  address: string
  status: string
  load: number
  uptime: string
  capabilities: string[]
  lastHeartbeat: string
}

const Nodes: React.FC = () => {
  const [nodes, setNodes] = useState<Node[]>([])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingNode, setEditingNode] = useState<Node | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    // Mock data for nodes
    setNodes([
      {
        id: 'node-1',
        type: 'worker',
        address: '192.168.1.101:50052',
        status: 'online',
        load: 45,
        uptime: '24h 30m',
        capabilities: ['task_execution', 'monitoring'],
        lastHeartbeat: '2024-01-01 10:30:00'
      },
      {
        id: 'node-2',
        type: 'worker',
        address: '192.168.1.102:50052',
        status: 'online',
        load: 65,
        uptime: '12h 15m',
        capabilities: ['task_execution', 'ai_assistance'],
        lastHeartbeat: '2024-01-01 10:29:00'
      },
      {
        id: 'node-3',
        type: 'worker',
        address: '192.168.1.103:50052',
        status: 'online',
        load: 25,
        uptime: '48h 05m',
        capabilities: ['task_execution', 'communication'],
        lastHeartbeat: '2024-01-01 10:30:00'
      },
      {
        id: 'node-4',
        type: 'worker',
        address: '192.168.1.104:50052',
        status: 'offline',
        load: 0,
        uptime: '0',
        capabilities: ['task_execution'],
        lastHeartbeat: '2024-01-01 08:15:00'
      },
    ])
  }, [])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'green'
      case 'offline': return 'red'
      case 'connecting': return 'orange'
      default: return 'default'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online': return <CheckCircleOutlined /> 
      case 'offline': return <CloseCircleOutlined /> 
      default: return null
    }
  }

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
          {getStatusIcon(status)} {status}
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
    { title: 'Uptime', dataIndex: 'uptime', key: 'uptime' },
    { 
      title: 'Capabilities', 
      dataIndex: 'capabilities', 
      key: 'capabilities',
      render: (capabilities: string[]) => (
        <Space direction="vertical" size={4}>
          {capabilities.map((capability, index) => (
            <Tag key={index}>{capability}</Tag>
          ))}
        </Space>
      )
    },
    { title: 'Last Heartbeat', dataIndex: 'lastHeartbeat', key: 'lastHeartbeat' },
    { 
      title: 'Actions', 
      key: 'actions',
      render: (_: any, record: Node) => (
        <Space size="middle">
          <Button type="primary" icon={<EditOutlined />} size="small" onClick={() => handleEdit(record)}>
            Edit
          </Button>
          <Button danger icon={<PoweroffOutlined />} size="small" onClick={() => handleRestart(record.id)}>
            Restart
          </Button>
          <Button danger icon={<DeleteOutlined />} size="small" onClick={() => handleDelete(record.id)}>
            Delete
          </Button>
        </Space>
      )
    },
  ]

  const handleAdd = () => {
    setEditingNode(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEdit = (node: Node) => {
    setEditingNode(node)
    form.setFieldsValue({
      type: node.type,
      address: node.address,
      capabilities: node.capabilities
    })
    setIsModalOpen(true)
  }

  const handleDelete = (nodeId: string) => {
    // In a real application, you would call an API to delete the node
    setNodes(nodes.filter(node => node.id !== nodeId))
    message.success('Node deleted successfully')
  }

  const handleRestart = (nodeId: string) => {
    // In a real application, you would call an API to restart the node
    setNodes(nodes.map(node => 
      node.id === nodeId 
        ? { ...node, status: 'connecting' }
        : node
    ))
    message.success('Node restarting...')
    // Simulate restart completion
    setTimeout(() => {
      setNodes(nodes.map(node => 
        node.id === nodeId 
          ? { ...node, status: 'online' }
          : node
      ))
      message.success('Node restarted successfully')
    }, 2000)
  }

  const handleSubmit = () => {
    form.validateFields().then(values => {
      if (editingNode) {
        // Update existing node
        const updatedNodes = nodes.map(node => 
          node.id === editingNode.id 
            ? { ...node, ...values }
            : node
        )
        setNodes(updatedNodes)
        message.success('Node updated successfully')
      } else {
        // Create new node
        const newNode: Node = {
          id: Math.random().toString(36).substr(2, 9),
          ...values,
          status: 'connecting',
          load: 0,
          uptime: '0',
          lastHeartbeat: new Date().toISOString()
        }
        setNodes([...nodes, newNode])
        message.success('Node added successfully')
        // Simulate node coming online
        setTimeout(() => {
          setNodes(nodes => nodes.map(node => 
            node.id === newNode.id 
              ? { ...node, status: 'online' }
              : node
          ))
        }, 2000)
      }
      setIsModalOpen(false)
    })
  }

  const handleRefresh = () => {
    // In a real application, you would call an API to refresh the nodes
    message.info('Nodes refreshed')
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2}>Nodes</Title>
        <Space>
          <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            Add Node
          </Button>
        </Space>
      </div>

      <Card bordered={false}>
        <Table 
          columns={nodeColumns} 
          dataSource={nodes} 
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title={editingNode ? 'Edit Node' : 'Add Node'}
        open={isModalOpen}
        onOk={handleSubmit}
        onCancel={() => setIsModalOpen(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="type"
            label="Node Type"
            rules={[{ required: true, message: 'Please select node type' }]}
          >
            <Select placeholder="Select node type">
              <Option value="worker">Worker</Option>
              <Option value="master">Master</Option>
              <Option value="relay">Relay</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="address"
            label="Address"
            rules={[{ required: true, message: 'Please enter node address' }]}
          >
            <Input placeholder="Enter node address (e.g., 192.168.1.100:50052)" />
          </Form.Item>

          <Form.Item
            name="capabilities"
            label="Capabilities"
            rules={[{ required: true, message: 'Please select capabilities' }]}
          >
            <Select mode="multiple" placeholder="Select capabilities">
              <Option value="task_execution">Task Execution</Option>
              <Option value="monitoring">Monitoring</Option>
              <Option value="ai_assistance">AI Assistance</Option>
              <Option value="communication">Communication</Option>
              <Option value="storage">Storage</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Nodes
