import React, { useEffect, useRef, useState } from 'react'
import { Layout, Menu, Button, Typography, Space, Form, Input, Select, message, Upload, Tag, Divider } from 'antd'
import { getJSON, postJSON, postRaw, wsURL } from './api'

const { Header, Content, Sider } = Layout
const { Title, Paragraph, Text } = Typography

type MemberMap = Record<string, string>

export default function App() {
  const [tab, setTab] = useState('wizard')
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider>
        <Menu theme="dark" mode="inline" selectedKeys={[tab]} onClick={(e) => setTab(e.key)}
          items={[
            { key: 'wizard', label: '设置向导' },
            { key: 'accounts', label: '账号池' },
            { key: 'status', label: '接口状态' },
            { key: 'logs', label: '实时日志' },
            { key: 'backup', label: '备份/恢复' },
          ]}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff' }}>
          <Title level={4} style={{ margin: 0 }}>SZU-NetManager</Title>
        </Header>
        <Content style={{ padding: 16 }}>
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
      <Button onClick={load} loading={loading}>加载接口</Button>
      {Object.keys(members).length > 0 && (
        <div>
          <Title level={5}>选择要用于多拨的接口并设置 NIC</Title>
          <Space direction="vertical" size="middle" style={{ width: 640 }}>
            {Object.keys(members).map((wan) => (
              <div key={wan} style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <Input style={{ width: 120 }} value={wan} disabled />
                <Input style={{ width: 260 }} placeholder="如 eth0/eth1" onChange={(e) => setMapping({ ...mapping, [wan]: e.target.value })} />
                <Button onClick={() => login(wan)}>立即登录</Button>
              </div>
            ))}
          </Space>
          <div style={{ marginTop: 16 }}>
            <Button type="primary" onClick={onSave}>保存映射</Button>
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

  const load = async () => { setList(await getJSON<Account[]>("/api/accounts")) }
  useEffect(() => { load() }, [])

  const onAdd = async () => {
    const v = await form.validateFields()
    await postJSON("/api/accounts", { Username: v.username, Password: v.password, Bandwidth: Number(v.bandwidth) })
    form.resetFields(); load()
  }

  return (
    <Space direction="vertical" style={{ width: 600 }}>
      <Form form={form} layout="inline">
        <Form.Item name="username" label="账号" rules={[{ required: true }]}><Input /></Form.Item>
        <Form.Item name="password" label="密码" rules={[{ required: true }]}><Input.Password /></Form.Item>
        <Form.Item name="bandwidth" label="带宽" rules={[{ required: true }]}>
          <Select style={{ width: 120 }} options={[20,50,100,200].map(v=>({value:String(v), label:`${v}M`}))} />
        </Form.Item>
        <Form.Item><Button type="primary" onClick={onAdd}>添加</Button></Form.Item>
      </Form>
      <div>
        <Title level={5}>账号列表</Title>
        <ul>
          {list.map(a => (
            <li key={a.id}><Text strong>{a.username}</Text> - {a.bandwidth}M - {a.status}</li>
          ))}
        </ul>
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
    <Space direction="vertical" style={{ width: 900 }}>
      <div>
        <Button onClick={load}>刷新</Button>
      </div>
      <div>
        {ifaces.map(i => (
          <div key={i.name} style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
            <Tag color={i.status === 'online' ? 'green' : (i.status === 'offline' ? 'red' : 'default')}>{i.name}: {i.status}</Tag>
            <Button size="small" onClick={() => login(i.name)}>尝试登录</Button>
          </div>
        ))}
      </div>
      <Divider />
      <pre style={{ maxWidth: 900, whiteSpace: 'pre-wrap' }}>{raw}</pre>
    </Space>
  )
}

function Logs() {
  const [logs, setLogs] = useState<string[]>([])
  const boxRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const ws = new WebSocket(wsURL('/ws'))
    ws.onmessage = (ev) => setLogs(prev => [...prev.slice(-200), String(ev.data)])
    return () => ws.close()
  }, [])
  useEffect(() => { boxRef.current?.scrollTo(0, boxRef.current.scrollHeight) }, [logs])
  return (
    <div ref={boxRef} style={{ border: '1px solid #ddd', height: 400, overflow: 'auto', padding: 8 }}>
      {logs.map((l, i) => <div key={i}>{l}</div>)}
    </div>
  )
}

function BackupRestore() {
  const onBackup = () => { window.open('/api/backup', '_blank') }
  const onRestore = async (file: File) => {
    const buf = await file.arrayBuffer()
    await postRaw('/api/restore', buf)
    message.success('恢复完成，重启服务后生效')
  }
  return (
    <Space>
      <Button onClick={onBackup}>备份配置</Button>
      <Upload beforeUpload={(f) => { onRestore(f); return false }}><Button>恢复配置</Button></Upload>
    </Space>
  )
}
