import React, { useState, useEffect } from 'react'
import { Card, Table, Button, Modal, Form, Input, Select, Tag, Space, Typography, message, Progress, Descriptions, Alert, Row, Col } from 'antd'
import { ReloadOutlined, ScanOutlined, LockOutlined, SecurityScanOutlined } from '@ant-design/icons'

const { Option } = Select

const { Title, Paragraph } = Typography

interface Vulnerability {
  id: string
  tool: string
  severity: string
  description: string
  status: string
  fix: string
}

interface APIKey {
  id: string
  name: string
  key: string
  createdAt: string
  lastUsed: string
  permissions: string[]
}

const Security: React.FC = () => {
  const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([])
  const [apiKeys, setApiKeys] = useState<APIKey[]>([])
  const [securityScore, setSecurityScore] = useState(85)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    // Mock data for vulnerabilities
    setVulnerabilities([
      {
        id: '1',
        tool: 'Nezha',
        severity: 'medium',
        description: 'Potential XSS vulnerability in dashboard',
        status: 'fixed',
        fix: 'Updated to latest version'
      },
      {
        id: '2',
        tool: 'Hermes Agent',
        severity: 'low',
        description: 'Insecure password storage',
        status: 'fixed',
        fix: 'Implemented proper encryption'
      },
      {
        id: '3',
        tool: 'OpenClaw',
        severity: 'high',
        description: 'Remote code execution vulnerability',
        status: 'pending',
        fix: 'Update to version 1.2.0'
      },
      {
        id: '4',
        tool: 'Ray Platform',
        severity: 'low',
        description: 'Missing rate limiting on API endpoints',
        status: 'fixed',
        fix: 'Implemented rate limiting'
      },
    ])

    // Mock data for API keys
    setApiKeys([
      {
        id: '1',
        name: 'Admin Key',
        key: 'sk_admin_1234567890abcdef',
        createdAt: '2024-01-01 10:00:00',
        lastUsed: '2024-01-01 11:30:00',
        permissions: ['read', 'write', 'admin']
      },
      {
        id: '2',
        name: 'Read-Only Key',
        key: 'sk_read_1234567890abcdef',
        createdAt: '2024-01-01 09:00:00',
        lastUsed: '2024-01-01 10:15:00',
        permissions: ['read']
      },
    ])
  }, [])

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high': return 'red'
      case 'medium': return 'orange'
      case 'low': return 'yellow'
      default: return 'default'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'fixed': return 'green'
      case 'pending': return 'orange'
      case 'ignored': return 'default'
      default: return 'default'
    }
  }

  const vulnerabilityColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id' },
    { title: 'Tool', dataIndex: 'tool', key: 'tool' },
    { 
      title: 'Severity', 
      dataIndex: 'severity', 
      key: 'severity',
      render: (severity: string) => (
        <Tag color={getSeverityColor(severity)}>{severity}</Tag>
      )
    },
    { title: 'Description', dataIndex: 'description', key: 'description', ellipsis: true },
    { 
      title: 'Status', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{status}</Tag>
      )
    },
    { title: 'Fix', dataIndex: 'fix', key: 'fix', ellipsis: true },
  ]

  const apiKeyColumns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Key', dataIndex: 'key', key: 'key', ellipsis: true },
    { title: 'Created At', dataIndex: 'createdAt', key: 'createdAt' },
    { title: 'Last Used', dataIndex: 'lastUsed', key: 'lastUsed' },
    { 
      title: 'Permissions', 
      dataIndex: 'permissions', 
      key: 'permissions',
      render: (permissions: string[]) => (
        <Space direction="vertical" size={4}>
          {permissions.map((permission, index) => (
            <Tag key={index}>{permission}</Tag>
          ))}
        </Space>
      )
    },
  ]

  const handleScan = () => {
    // In a real application, you would call an API to start a security scan
    message.info('Starting security scan...')
    // Simulate scan completion
    setTimeout(() => {
      message.success('Security scan completed')
      // Update security score
      setSecurityScore(92)
    }, 3000)
  }

  const handleGenerateAPIKey = () => {
    setIsModalOpen(true)
  }

  const handleSubmit = () => {
    form.validateFields().then(values => {
      // Create new API key
      const newApiKey: APIKey = {
        id: Math.random().toString(36).substr(2, 9),
        name: values.name,
        key: `sk_${values.name.toLowerCase().replace(/\s+/g, '_')}_${Math.random().toString(36).substr(2, 16)}`,
        createdAt: new Date().toISOString(),
        lastUsed: '',
        permissions: values.permissions
      }
      setApiKeys([...apiKeys, newApiKey])
      message.success('API key generated successfully')
      setIsModalOpen(false)
    })
  }

  const handleRefresh = () => {
    // In a real application, you would call an API to refresh security data
    message.info('Security data refreshed')
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2}>Security</Title>
        <Space>
          <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
          <Button type="primary" icon={<ScanOutlined />} onClick={handleScan}>
            Scan Now
          </Button>
          <Button type="primary" icon={<LockOutlined />} onClick={handleGenerateAPIKey}>
            Generate API Key
          </Button>
        </Space>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <SecurityScanOutlined style={{ fontSize: 48, color: '#1890ff', marginBottom: 16 }} />
              <Title level={3}>{securityScore}%</Title>
              <Paragraph>Security Score</Paragraph>
              <Progress percent={securityScore} status={securityScore > 80 ? 'success' : securityScore > 60 ? 'normal' : 'exception'} />
            </div>
          </Card>
        </Col>
        <Col span={16}>
          <Card title="Security Overview">
            <Descriptions column={2}>
              <Descriptions.Item label="Total Vulnerabilities">4</Descriptions.Item>
              <Descriptions.Item label="Fixed Vulnerabilities">3</Descriptions.Item>
              <Descriptions.Item label="Pending Vulnerabilities">1</Descriptions.Item>
              <Descriptions.Item label="High Severity">1</Descriptions.Item>
              <Descriptions.Item label="Medium Severity">1</Descriptions.Item>
              <Descriptions.Item label="Low Severity">2</Descriptions.Item>
            </Descriptions>
            <Alert 
              message="Security Reminder" 
              description="There is 1 high severity vulnerability that needs attention. Please update OpenClaw to version 1.2.0 as soon as possible." 
              type="warning" 
              style={{ marginTop: 16 }} 
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Card title="Vulnerabilities" bordered={false}>
            <Table 
              columns={vulnerabilityColumns} 
              dataSource={vulnerabilities} 
              rowKey="id"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="API Keys" bordered={false}>
            <Table 
              columns={apiKeyColumns} 
              dataSource={apiKeys} 
              rowKey="id"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>
      </Row>

      <Modal
        title="Generate API Key"
        open={isModalOpen}
        onOk={handleSubmit}
        onCancel={() => setIsModalOpen(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="Key Name"
            rules={[{ required: true, message: 'Please enter key name' }]}
          >
            <Input placeholder="Enter key name" />
          </Form.Item>

          <Form.Item
            name="permissions"
            label="Permissions"
            rules={[{ required: true, message: 'Please select permissions' }]}
          >
            <Select mode="multiple" placeholder="Select permissions">
              <Option value="read">Read</Option>
              <Option value="write">Write</Option>
              <Option value="admin">Admin</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Security
