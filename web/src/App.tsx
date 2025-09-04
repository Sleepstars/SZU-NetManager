import React, { useEffect, useRef, useState } from 'react'
import { Layout, Menu, Button, Typography, Space, Form, Input, Select, message, Upload, Tag, Divider, Drawer, Switch, Dropdown } from 'antd'
import { MenuOutlined, SunOutlined, MoonOutlined, SettingOutlined, BulbOutlined } from '@ant-design/icons'
import { getJSON, postJSON, postRaw, wsURL } from './api'
import { useTheme } from './contexts/ThemeContext'
import { useIsMobile } from './hooks/useResponsive'

const { Header, Content, Sider } = Layout
const { Title, Paragraph, Text } = Typography

type MemberMap = Record<string, string>

export default function App() {
  const [tab, setTab] = useState('wizard')
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const { themeMode, isDark, setThemeMode } = useTheme()
  const isMobile = useIsMobile()

  const menuItems = [
    { key: 'wizard', label: '设置向导' },
    { key: 'accounts', label: '账号池' },
    { key: 'status', label: '接口状态' },
    { key: 'logs', label: '实时日志' },
    { key: 'backup', label: '备份/恢复' },
  ]

  const handleMenuClick = (e: any) => {
    setTab(e.key)
    if (isMobile) {
      setMobileMenuOpen(false)
    }
  }

  const themeMenuItems = [
    {
      key: 'light',
      label: (
        <Space>
          <SunOutlined />
          浅色模式
        </Space>
      ),
    },
    {
      key: 'dark',
      label: (
        <Space>
          <MoonOutlined />
          深色模式
        </Space>
      ),
    },
    {
      key: 'auto',
      label: (
        <Space>
          <BulbOutlined />
          跟随系统
        </Space>
      ),
    },
  ]

  const handleThemeChange = ({ key }: { key: string }) => {
    setThemeMode(key as any)
  }

  const siderContent = (
    <Menu
      theme="dark"
      mode="inline"
      selectedKeys={[tab]}
      onClick={handleMenuClick}
      items={menuItems}
      style={{ height: '100%', borderRight: 0 }}
    />
  )

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 桌面端侧边栏 */}
      {!isMobile && (
        <Sider
          width={200}
          className="desktop-only"
          style={{
            background: isDark ? '#001529' : '#722ed1',
          }}
        >
          {siderContent}
        </Sider>
      )}

      {/* 移动端抽屉菜单 */}
      {isMobile && (
        <>
          <div
            className={`mobile-overlay ${mobileMenuOpen ? 'show' : ''}`}
            onClick={() => setMobileMenuOpen(false)}
          />
          <Drawer
            title="菜单"
            placement="left"
            onClose={() => setMobileMenuOpen(false)}
            open={mobileMenuOpen}
            bodyStyle={{ padding: 0 }}
            width={250}
          >
            {siderContent}
          </Drawer>
        </>
      )}

      <Layout>
        <Header style={{
          background: isDark ? '#001529' : '#fff',
          padding: '0 16px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: `1px solid ${isDark ? '#434343' : '#f0f0f0'}`
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {isMobile && (
              <Button
                type="text"
                icon={<MenuOutlined />}
                onClick={() => setMobileMenuOpen(true)}
                style={{ color: isDark ? '#fff' : '#000' }}
              />
            )}
            <Title level={4} style={{
              margin: 0,
              color: isDark ? '#fff' : '#000',
              fontSize: isMobile ? '18px' : '20px'
            }}>
              SZU-NetManager
            </Title>
          </div>

          <Dropdown
            menu={{
              items: themeMenuItems,
              onClick: handleThemeChange,
              selectedKeys: [themeMode],
            }}
            trigger={['click']}
          >
            <Button
              type="text"
              icon={<SettingOutlined />}
              style={{ color: isDark ? '#fff' : '#000' }}
            />
          </Dropdown>
        </Header>

        <Content className="responsive-space">
          {tab === 'wizard' && <Wizard />}
          {tab === 'accounts' && <Accounts />}
          {tab === 'status' && <Status />}
          {tab === 'logs' && <Logs />}
          {tab === 'backup' && <BackupRestore />}
        </Content>
      </Layout>
    </Layout>
  )
}

function Wizard() {
  const [members, setMembers] = useState<MemberMap>({})
  const [mapping, setMapping] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)
  const isMobile = useIsMobile()

  const load = async () => {
    setLoading(true)
    try {
      const res = await getJSON<{ member_map: MemberMap }>("/api/mwan/interfaces")
      setMembers(res.member_map)
    } catch (e: any) { message.error(e.message) }
    finally { setLoading(false) }
  }

  const login = async (wan: string) => {
    const hide = message.loading(`正在为 ${wan} 触发登录...`)
    try {
      await fetch(`/api/login/start?wan=${encodeURIComponent(wan)}`, { method: 'POST' })
      message.success(`${wan} 登录任务已触发，请查看实时日志`)
    } catch (e: any) { message.error(e.message) }
    finally { hide() }
  }

  const onSave = async () => {
    try {
      await Promise.all(Object.entries(mapping).map(([wan, nic]) => postJSON("/api/iface-map", { WanIface: wan, Nic: nic })))
      message.success('已保存接口映射')
    } catch (e: any) { message.error(e.message) }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Button onClick={load} loading={loading} block={isMobile}>
        加载接口
      </Button>
      {Object.keys(members).length > 0 && (
        <div>
          <Title level={5}>选择要用于多拨的接口并设置 NIC</Title>
          <Space direction="vertical" size="middle" style={{ width: '100%', maxWidth: isMobile ? '100%' : 640 }}>
            {Object.keys(members).map((wan) => (
              <div key={wan} style={{
                display: 'flex',
                gap: 8,
                alignItems: 'center',
                flexDirection: isMobile ? 'column' : 'row',
                width: '100%'
              }}>
                <Input
                  style={{ width: isMobile ? '100%' : 120 }}
                  value={wan}
                  disabled
                  addonBefore={isMobile ? "接口" : undefined}
                />
                <Input
                  style={{ width: isMobile ? '100%' : 260 }}
                  placeholder="如 eth0/eth1"
                  onChange={(e) => setMapping({ ...mapping, [wan]: e.target.value })}
                  addonBefore={isMobile ? "NIC" : undefined}
                />
                <Button
                  onClick={() => login(wan)}
                  block={isMobile}
                  style={{ minWidth: isMobile ? '100%' : 'auto' }}
                >
                  立即登录
                </Button>
              </div>
            ))}
          </Space>
          <div style={{ marginTop: 16 }}>
            <Button type="primary" onClick={onSave} block={isMobile}>
              保存映射
            </Button>
          </div>
        </div>
      )}
    </Space>
  )
}

type Account = { id: number, username: string, password: string, bandwidth: number, status: string }

function Accounts() {
  const [list, setList] = useState<Account[]>([])
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const isMobile = useIsMobile()

  const load = async () => {
    setLoading(true)
    setError(null)
    try {
      const accounts = await getJSON<Account[]>("/api/accounts")
      setList(accounts)
    } catch (e: any) {
      console.error('Failed to load accounts:', e)
      setError(e.message || '加载账号列表失败')
      setList([]) // 确保有默认值
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const onAdd = async () => {
    try {
      const v = await form.validateFields()
      await postJSON("/api/accounts", { Username: v.username, Password: v.password, Bandwidth: Number(v.bandwidth) })
      form.resetFields()
      message.success('账号添加成功')
      load()
    } catch (e: any) {
      message.error(e.message || '添加账号失败')
    }
  }

  if (error) {
    return (
      <Space direction="vertical" style={{ width: '100%', textAlign: 'center', padding: 40 }}>
        <Text type="danger">加载失败: {error}</Text>
        <Button onClick={load} loading={loading}>重新加载</Button>
      </Space>
    )
  }

  return (
    <Space direction="vertical" style={{ width: '100%', maxWidth: isMobile ? '100%' : 600 }}>
      <div>
        <Title level={4}>账号池管理</Title>
        <Paragraph type="secondary">
          管理用于网络登录的账号信息
        </Paragraph>
      </div>

      <Form
        form={form}
        layout={isMobile ? "vertical" : "inline"}
        className="responsive-form"
      >
        <Form.Item name="username" label="账号" rules={[{ required: true }]}>
          <Input placeholder="请输入账号" />
        </Form.Item>
        <Form.Item name="password" label="密码" rules={[{ required: true }]}>
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        <Form.Item name="bandwidth" label="带宽" rules={[{ required: true }]}>
          <Select
            style={{ width: isMobile ? '100%' : 120 }}
            placeholder="选择带宽"
            options={[20,50,100,200].map(v=>({value:String(v), label:`${v}M`}))}
          />
        </Form.Item>
        <Form.Item>
          <Button type="primary" onClick={onAdd} block={isMobile} loading={loading}>
            添加账号
          </Button>
        </Form.Item>
      </Form>

      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <Title level={5} style={{ margin: 0 }}>账号列表</Title>
          <Button size="small" onClick={load} loading={loading}>刷新</Button>
        </div>

        <div className="responsive-table">
          {loading ? (
            <div style={{ textAlign: 'center', padding: 20 }}>
              <Text type="secondary">加载中...</Text>
            </div>
          ) : list.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 20 }}>
              <Text type="secondary">暂无账号，请添加第一个账号</Text>
            </div>
          ) : (
            <Space direction="vertical" style={{ width: '100%' }}>
              {list.map(a => (
                <div key={a.id} style={{
                  padding: 12,
                  border: '1px solid var(--border-color)',
                  borderRadius: 6,
                  background: 'var(--bg-secondary)'
                }}>
                  <Space direction={isMobile ? "vertical" : "horizontal"} style={{ width: '100%' }}>
                    <Text strong>{a.username}</Text>
                    <Tag color="purple">{a.bandwidth}M</Tag>
                    <Tag color={a.status === 'active' ? 'green' : 'default'}>{a.status}</Tag>
                  </Space>
                </div>
              ))}
            </Space>
          )}
        </div>
      </div>
    </Space>
  )
}

type IfaceState = { name: string; status: 'online' | 'offline' | 'unknown'; line: string }

function parseStatus(raw: string): IfaceState[] {
  const out: IfaceState[] = []
  raw.split(/\n+/).forEach((ln) => {
    const m = ln.match(/^interface\s+(\S+)\s+is\s+(online|offline)/i)
    if (m) {
      out.push({ name: m[1], status: m[2].toLowerCase() as any, line: ln })
    }
  })
  return out
}

function Status() {
  const [raw, setRaw] = useState('')
  const [ifaces, setIfaces] = useState<IfaceState[]>([])
  const isMobile = useIsMobile()

  const load = async () => {
    const r = await getJSON<{ status: string }>("/api/mwan/status")
    setRaw(r.status)
    setIfaces(parseStatus(r.status))
  }
  useEffect(() => { load() }, [])

  const login = async (wan: string) => {
    const hide = message.loading(`正在为 ${wan} 触发登录...`)
    try {
      await fetch(`/api/login/start?wan=${encodeURIComponent(wan)}`, { method: 'POST' })
      message.success(`${wan} 登录任务已触发，请查看实时日志`)
    } catch (e: any) { message.error(e.message) }
    finally { hide() }
  }

  return (
    <Space direction="vertical" style={{ width: '100%', maxWidth: isMobile ? '100%' : 900 }}>
      <div>
        <Button onClick={load} block={isMobile}>刷新状态</Button>
      </div>
      <div>
        <Title level={5}>接口状态</Title>
        <Space direction="vertical" style={{ width: '100%' }}>
          {ifaces.map(i => (
            <div key={i.name} style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              marginBottom: 8,
              flexDirection: isMobile ? 'column' : 'row',
              padding: isMobile ? 12 : 8,
              border: isMobile ? '1px solid var(--border-color)' : 'none',
              borderRadius: isMobile ? 6 : 0,
              background: isMobile ? 'var(--bg-secondary)' : 'transparent'
            }}>
              <Tag
                color={i.status === 'online' ? 'green' : (i.status === 'offline' ? 'red' : 'default')}
                style={{ fontSize: isMobile ? 14 : 12 }}
              >
                {i.name}: {i.status}
              </Tag>
              <Button
                size={isMobile ? "middle" : "small"}
                onClick={() => login(i.name)}
                block={isMobile}
              >
                尝试登录
              </Button>
            </div>
          ))}
        </Space>
      </div>
      <Divider />
      <div>
        <Title level={5}>详细状态</Title>
        <pre style={{
          maxWidth: '100%',
          whiteSpace: 'pre-wrap',
          fontSize: isMobile ? 12 : 14,
          padding: 12,
          background: 'var(--bg-secondary)',
          border: '1px solid var(--border-color)',
          borderRadius: 6,
          overflow: 'auto'
        }}>
          {raw}
        </pre>
      </div>
    </Space>
  )
}

function Logs() {
  const [logs, setLogs] = useState<string[]>([])
  const boxRef = useRef<HTMLDivElement>(null)
  const isMobile = useIsMobile()

  useEffect(() => {
    const ws = new WebSocket(wsURL('/ws'))
    ws.onmessage = (ev) => setLogs(prev => [...prev.slice(-200), String(ev.data)])
    return () => ws.close()
  }, [])

  useEffect(() => {
    boxRef.current?.scrollTo(0, boxRef.current.scrollHeight)
  }, [logs])

  return (
    <div style={{ width: '100%' }}>
      <Title level={5}>实时日志</Title>
      <div
        ref={boxRef}
        style={{
          border: '1px solid var(--border-color)',
          height: isMobile ? 300 : 400,
          overflow: 'auto',
          padding: 12,
          background: 'var(--bg-secondary)',
          borderRadius: 6,
          fontSize: isMobile ? 12 : 14,
          fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace'
        }}
      >
        {logs.length === 0 ? (
          <Text type="secondary">等待日志输出...</Text>
        ) : (
          logs.map((l, i) => (
            <div key={i} style={{
              marginBottom: 4,
              wordBreak: 'break-all',
              lineHeight: 1.4
            }}>
              {l}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

function BackupRestore() {
  const isMobile = useIsMobile()

  const onBackup = () => {
    window.open('/api/backup', '_blank')
  }

  const onRestore = async (file: File) => {
    const buf = await file.arrayBuffer()
    await postRaw('/api/restore', buf)
    message.success('恢复完成，重启服务后生效')
  }

  return (
    <Space direction={isMobile ? "vertical" : "horizontal"} style={{ width: '100%' }}>
      <Title level={5}>备份与恢复</Title>
      <Space direction={isMobile ? "vertical" : "horizontal"} style={{ width: '100%' }}>
        <Button onClick={onBackup} block={isMobile} type="primary">
          备份配置
        </Button>
        <Upload
          beforeUpload={(f) => { onRestore(f); return false }}
          showUploadList={false}
        >
          <Button block={isMobile}>恢复配置</Button>
        </Upload>
      </Space>
      {isMobile && (
        <div style={{ marginTop: 16 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            备份：下载当前配置文件<br/>
            恢复：选择配置文件进行恢复
          </Text>
        </div>
      )}
    </Space>
  )
}
