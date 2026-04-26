import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { Layout, Menu, ConfigProvider } from 'antd'
import { 
  DashboardOutlined, 
  ProjectOutlined, 
  AppstoreOutlined, 
  SettingOutlined,
  SecurityScanOutlined
} from '@ant-design/icons'

import Dashboard from './pages/Dashboard'
import Tasks from './pages/Tasks'
import Nodes from './pages/Nodes'
import Settings from './pages/Settings'
import Security from './pages/Security'

const { Header, Content, Sider } = Layout

function App() {
  return (
    <ConfigProvider theme={{ token: { colorPrimary: '#1890ff' } }}>
      <Router>
        <Layout style={{ minHeight: '100vh' }}>
          <Sider theme="dark" width={200}>
            <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontSize: 18, fontWeight: 'bold' }}>
              Ray Platform
            </div>
            <Menu
              theme="dark"
              mode="inline"
              defaultSelectedKeys={['dashboard']}
              items={[
                {
                  key: 'dashboard',
                  icon: <DashboardOutlined />,
                  label: <a href="/dashboard">Dashboard</a>,
                },
                {
                  key: 'tasks',
                  icon: <ProjectOutlined />,
                  label: <a href="/tasks">Tasks</a>,
                },
                {
                  key: 'nodes',
                  icon: <AppstoreOutlined />,
                  label: <a href="/nodes">Nodes</a>,
                },
                {
                  key: 'security',
                  icon: <SecurityScanOutlined />,
                  label: <a href="/security">Security</a>,
                },
                {
                  key: 'settings',
                  icon: <SettingOutlined />,
                  label: <a href="/settings">Settings</a>,
                },
              ]}
            />
          </Sider>
          <Layout>
            <Header style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', backgroundColor: 'white' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <span>Admin</span>
              </div>
            </Header>
            <Content style={{ margin: '24px 16px', padding: 24, background: 'white', minHeight: 280 }}>
              <Routes>
                <Route path="/dashboard" element={<Dashboard />} />
                <Route path="/tasks" element={<Tasks />} />
                <Route path="/nodes" element={<Nodes />} />
                <Route path="/security" element={<Security />} />
                <Route path="/settings" element={<Settings />} />
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
              </Routes>
            </Content>
          </Layout>
        </Layout>
      </Router>
    </ConfigProvider>
  )
}

export default App
