import React, { useState, useEffect } from 'react'
import { Card, Form, Input, Select, Switch, Button, Space, Typography, message, Divider, Descriptions, Alert } from 'antd'
import { SaveOutlined, ReloadOutlined } from '@ant-design/icons'

const { Title } = Typography
const { Option } = Select

interface Settings {
  masterAddress: string
  heartbeatInterval: number
  taskTimeout: number
  maxRetries: number
  enableTLS: boolean
  enableAuthentication: boolean
  logLevel: string
  nodeCapabilities: string[]
}

const Settings: React.FC = () => {
  const [settings, setSettings] = useState<Settings>({
    masterAddress: 'localhost:50051',
    heartbeatInterval: 30,
    taskTimeout: 300,
    maxRetries: 3,
    enableTLS: false,
    enableAuthentication: true,
    logLevel: 'info',
    nodeCapabilities: ['task_execution', 'monitoring']
  })
  const [form] = Form.useForm()

  useEffect(() => {
    form.setFieldsValue(settings)
  }, [settings, form])

  const handleSubmit = () => {
    form.validateFields().then(values => {
      // In a real application, you would call an API to save the settings
      setSettings(values)
      message.success('Settings saved successfully')
    })
  }

  const handleReset = () => {
    form.resetFields()
  }

  const handleRefresh = () => {
    // In a real application, you would call an API to refresh the settings
    message.info('Settings refreshed')
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2}>Settings</Title>
        <Space>
          <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
        </Space>
      </div>

      <Card title="Platform Settings" bordered={false} style={{ marginBottom: 24 }}>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="masterAddress"
            label="Master Address"
            rules={[{ required: true, message: 'Please enter master address' }]}
          >
            <Input placeholder="Enter master address (e.g., localhost:50051)" />
          </Form.Item>

          <Form.Item
            name="heartbeatInterval"
            label="Heartbeat Interval (seconds)"
            rules={[{ required: true, message: 'Please enter heartbeat interval' }]}
          >
            <Input type="number" min={1} max={300} placeholder="Enter heartbeat interval" />
          </Form.Item>

          <Form.Item
            name="taskTimeout"
            label="Task Timeout (seconds)"
            rules={[{ required: true, message: 'Please enter task timeout' }]}
          >
            <Input type="number" min={10} max={3600} placeholder="Enter task timeout" />
          </Form.Item>

          <Form.Item
            name="maxRetries"
            label="Max Retries"
            rules={[{ required: true, message: 'Please enter max retries' }]}
          >
            <Input type="number" min={0} max={10} placeholder="Enter max retries" />
          </Form.Item>

          <Form.Item
            name="logLevel"
            label="Log Level"
            rules={[{ required: true, message: 'Please select log level' }]}
          >
            <Select placeholder="Select log level">
              <Option value="debug">Debug</Option>
              <Option value="info">Info</Option>
              <Option value="warn">Warn</Option>
              <Option value="error">Error</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="enableTLS"
            label="Enable TLS"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item
            name="enableAuthentication"
            label="Enable Authentication"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item
            name="nodeCapabilities"
            label="Node Capabilities"
            rules={[{ required: true, message: 'Please select node capabilities' }]}
          >
            <Select mode="multiple" placeholder="Select node capabilities">
              <Option value="task_execution">Task Execution</Option>
              <Option value="monitoring">Monitoring</Option>
              <Option value="ai_assistance">AI Assistance</Option>
              <Option value="communication">Communication</Option>
              <Option value="storage">Storage</Option>
            </Select>
          </Form.Item>

          <Space style={{ marginTop: 24 }}>
            <Button type="primary" icon={<SaveOutlined />} htmlType="submit">
              Save Settings
            </Button>
            <Button onClick={handleReset}>
              Reset
            </Button>
          </Space>
        </Form>
      </Card>

      <Card title="Tool Integrations" bordered={false}>
        <Alert 
          message="Tool Integration Settings" 
          description="Configure how the platform integrates with external tools like Nezha, Hermes Agent, and OpenClaw." 
          type="info" 
          style={{ marginBottom: 24 }} 
        />

        <Descriptions column={2}>
          <Descriptions.Item label="Nezha Path">/workspace/nezha</Descriptions.Item>
          <Descriptions.Item label="Status">Connected</Descriptions.Item>
          <Descriptions.Item label="Hermes Agent Path">/workspace/hermes-agent</Descriptions.Item>
          <Descriptions.Item label="Status">Connected</Descriptions.Item>
          <Descriptions.Item label="OpenClaw Path">/workspace/openclaw</Descriptions.Item>
          <Descriptions.Item label="Status">Connected</Descriptions.Item>
        </Descriptions>

        <Divider />

        <Title level={4}>Integration Settings</Title>
        <Form layout="vertical">
          <Form.Item
            name="nezhaApiKey"
            label="Nezha API Key"
          >
            <Input.Password placeholder="Enter Nezha API key" />
          </Form.Item>

          <Form.Item
            name="hermesApiKey"
            label="Hermes Agent API Key"
          >
            <Input.Password placeholder="Enter Hermes Agent API key" />
          </Form.Item>

          <Form.Item
            name="openclawApiKey"
            label="OpenClaw API Key"
          >
            <Input.Password placeholder="Enter OpenClaw API key" />
          </Form.Item>

          <Button type="primary" icon={<SaveOutlined />}>
            Save Integration Settings
          </Button>
        </Form>
      </Card>
    </div>
  )
}

export default Settings
