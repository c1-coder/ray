import React, { useState, useEffect } from 'react'
import { Card, Table, Button, Modal, Form, Select, Input, Tag, Space, Typography, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'

const { Title } = Typography
const { Option } = Select
const { TextArea } = Input

interface Task {
  id: string
  type: string
  status: string
  priority: string
  assignedTo: string
  payload: string
  createdAt: string
  updatedAt: string
}

const Tasks: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([])
  const [nodes, setNodes] = useState<{id: string, address: string}[]>([])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    // Mock data for tasks
    setTasks([
      {
        id: '1',
        type: 'Monitoring',
        status: 'completed',
        priority: 'high',
        assignedTo: 'node-1',
        payload: '{"target": "server-1", "metrics": ["cpu", "memory", "disk"]}',
        createdAt: '2024-01-01 10:00:00',
        updatedAt: '2024-01-01 10:30:00'
      },
      {
        id: '2',
        type: 'AI',
        status: 'running',
        priority: 'medium',
        assignedTo: 'node-2',
        payload: '{"prompt": "Generate a summary of the latest news", "model": "gpt-4"}',
        createdAt: '2024-01-01 11:00:00',
        updatedAt: '2024-01-01 11:15:00'
      },
      {
        id: '3',
        type: 'Communication',
        status: 'pending',
        priority: 'low',
        assignedTo: '',
        payload: '{"recipient": "user@example.com", "message": "Hello from Ray Platform"}',
        createdAt: '2024-01-01 12:00:00',
        updatedAt: '2024-01-01 12:00:00'
      },
      {
        id: '4',
        type: 'Monitoring',
        status: 'completed',
        priority: 'medium',
        assignedTo: 'node-3',
        payload: '{"target": "server-2", "metrics": ["cpu", "memory"]}',
        createdAt: '2024-01-01 09:00:00',
        updatedAt: '2024-01-01 09:30:00'
      },
      {
        id: '5',
        type: 'AI',
        status: 'pending',
        priority: 'high',
        assignedTo: '',
        payload: '{"prompt": "Analyze the sales data", "model": "gpt-4"}',
        createdAt: '2024-01-01 13:00:00',
        updatedAt: '2024-01-01 13:00:00'
      },
    ])

    // Mock data for nodes
    setNodes([
      { id: 'node-1', address: '192.168.1.101:50052' },
      { id: 'node-2', address: '192.168.1.102:50052' },
      { id: 'node-3', address: '192.168.1.103:50052' },
      { id: 'node-4', address: '192.168.1.104:50052' },
    ])
  }, [])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'green'
      case 'running': return 'blue'
      case 'pending': return 'orange'
      case 'failed': return 'red'
      default: return 'default'
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'red'
      case 'medium': return 'orange'
      case 'low': return 'green'
      default: return 'default'
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
        <Tag color={getStatusColor(status)}>{status}</Tag>
      )
    },
    { 
      title: 'Priority', 
      dataIndex: 'priority', 
      key: 'priority',
      render: (priority: string) => (
        <Tag color={getPriorityColor(priority)}>{priority}</Tag>
      )
    },
    { title: 'Assigned To', dataIndex: 'assignedTo', key: 'assignedTo' },
    { title: 'Payload', dataIndex: 'payload', key: 'payload', ellipsis: true },
    { title: 'Created At', dataIndex: 'createdAt', key: 'createdAt' },
    { title: 'Updated At', dataIndex: 'updatedAt', key: 'updatedAt' },
    { 
      title: 'Actions', 
      key: 'actions',
      render: (_: any, record: Task) => (
        <Space size="middle">
          <Button type="primary" icon={<EditOutlined />} size="small" onClick={() => handleEdit(record)}>
            Edit
          </Button>
          <Button danger icon={<DeleteOutlined />} size="small" onClick={() => handleDelete(record.id)}>
            Delete
          </Button>
        </Space>
      )
    },
  ]

  const handleAdd = () => {
    setEditingTask(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEdit = (task: Task) => {
    setEditingTask(task)
    form.setFieldsValue({
      type: task.type,
      priority: task.priority,
      assignedTo: task.assignedTo,
      payload: task.payload
    })
    setIsModalOpen(true)
  }

  const handleDelete = (taskId: string) => {
    // In a real application, you would call an API to delete the task
    setTasks(tasks.filter(task => task.id !== taskId))
    message.success('Task deleted successfully')
  }

  const handleSubmit = () => {
    form.validateFields().then(values => {
      if (editingTask) {
        // Update existing task
        const updatedTasks = tasks.map(task => 
          task.id === editingTask.id 
            ? { ...task, ...values, updatedAt: new Date().toISOString() }
            : task
        )
        setTasks(updatedTasks)
        message.success('Task updated successfully')
      } else {
        // Create new task
        const newTask: Task = {
          id: Math.random().toString(36).substr(2, 9),
          ...values,
          status: 'pending',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
        setTasks([...tasks, newTask])
        message.success('Task created successfully')
      }
      setIsModalOpen(false)
    })
  }

  const handleRefresh = () => {
    // In a real application, you would call an API to refresh the tasks
    message.info('Tasks refreshed')
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2}>Tasks</Title>
        <Space>
          <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            Add Task
          </Button>
        </Space>
      </div>

      <Card bordered={false}>
        <Table 
          columns={taskColumns} 
          dataSource={tasks} 
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title={editingTask ? 'Edit Task' : 'Add Task'}
        open={isModalOpen}
        onOk={handleSubmit}
        onCancel={() => setIsModalOpen(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="type"
            label="Task Type"
            rules={[{ required: true, message: 'Please select task type' }]}
          >
            <Select placeholder="Select task type">
              <Option value="Monitoring">Monitoring</Option>
              <Option value="AI">AI</Option>
              <Option value="Communication">Communication</Option>
              <Option value="Custom">Custom</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="priority"
            label="Priority"
            rules={[{ required: true, message: 'Please select priority' }]}
          >
            <Select placeholder="Select priority">
              <Option value="high">High</Option>
              <Option value="medium">Medium</Option>
              <Option value="low">Low</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="assignedTo"
            label="Assign To"
          >
            <Select placeholder="Select node (optional)" allowClear>
              {nodes.map(node => (
                <Option key={node.id} value={node.id}>{node.id} ({node.address})</Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="payload"
            label="Payload"
            rules={[{ required: true, message: 'Please enter payload' }]}
          >
            <TextArea rows={4} placeholder="Enter JSON payload" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Tasks
