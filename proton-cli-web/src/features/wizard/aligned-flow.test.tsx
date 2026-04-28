import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Wizard } from './Wizard'

async function startStandardFlow(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: '标准模式部署' }))
  await user.click(screen.getByRole('button', { name: '新增节点' }))
  await user.type(screen.getByLabelText('Node IPv4'), '192.168.40.11')
}

async function fillRequiredServicePasswords(user: ReturnType<typeof userEvent.setup>) {
  const passwordInputs = screen
    .getAllByDisplayValue('')
    .filter((element): element is HTMLInputElement => element instanceof HTMLInputElement && element.type === 'password')

  await user.type(passwordInputs[0], 'mariadb-pass')
  await user.type(passwordInputs[1], 'redis-pass')
}

describe('aligned flow', () => {
  it('uses old footer button labels across the standard flow', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    expect(screen.getByRole('button', { name: '下一步' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: '下一步' }))
    expect(screen.getByRole('button', { name: '上一步' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '下一步' })).toBeInTheDocument()
  })

  it('only exposes the 完成 action on the final connect step', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await fillRequiredServicePasswords(user)
    expect(screen.queryByRole('button', { name: '完成' })).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: '下一步' }))
    expect(screen.getByRole('button', { name: '完成' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '下一步' })).not.toBeInTheDocument()
  })

  it('keeps the old network, repository and service form groups instead of over-pruning fields', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    expect(screen.queryByText('内部IP')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: '下一步' }))
    expect(screen.getByText('docker IP')).toBeInTheDocument()
    expect(screen.getByText('Pod 网段')).toBeInTheDocument()
    expect(screen.getByText('Serivce 网段')).toBeInTheDocument()
    expect(screen.getByText('etcd 数据路径')).toBeInTheDocument()
    expect(screen.getByText('可选插件')).toBeInTheDocument()
    expect(screen.getByRole('checkbox', { name: 'ingress-nginx' })).toBeChecked()

    await user.click(screen.getByRole('button', { name: '下一步' }))
    expect(screen.getByText('Chart仓库端口')).toBeInTheDocument()
    expect(screen.getByText('Registry端口')).toBeInTheDocument()
    expect(screen.queryByText('RPM仓库端口')).not.toBeInTheDocument()
    expect(screen.getByText('部署节点')).toBeInTheDocument()
    expect(screen.getByText('Chart仓库高可用端口')).toBeInTheDocument()
    expect(screen.getByText('Registry高可用端口')).toBeInTheDocument()
    expect(screen.queryByText('RPM仓库高可用端口')).not.toBeInTheDocument()
    expect(screen.getByText('chart与image的存储路径')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: '下一步' }))
    expect(screen.getAllByText('Innodb_buffer_size').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Requests.Memory').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Requests.CPU').length).toBeGreaterThan(0)
    expect(screen.getAllByText('数据路径').length).toBeGreaterThan(0)
    expect(screen.getAllByText('JVM配置').length).toBeGreaterThan(0)
    expect(screen.getByText('低警戒水位线')).toBeInTheDocument()
    expect(screen.getByText('日志保留字节数')).toBeInTheDocument()
    expect(screen.getByText('日志保留小时数')).toBeInTheDocument()
    expect(screen.getByText('日志段最大小时数')).toBeInTheDocument()
  })

  it('starts with an empty node list instead of a prefilled default node', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await user.click(screen.getByRole('button', { name: '标准模式部署' }))

    expect(screen.queryByDisplayValue('node1')).not.toBeInTheDocument()
    expect(screen.queryByDisplayValue('192.168.40.11')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Node IPv4')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: '新增节点' })).toBeInTheDocument()
  })

  it('keeps the managed kubernetes template disabled on the home screen', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    const managedButton = screen.getByRole('button', { name: '托管Kubernetes部署' })
    expect(managedButton).toBeDisabled()
    expect(screen.getByText('功能持续完善中')).toBeInTheDocument()

    await user.click(managedButton)

    expect(screen.getByRole('heading', { name: '初始化模板' })).toBeInTheDocument()
    expect(screen.queryByText('命名空间')).not.toBeInTheDocument()
    expect(screen.queryByText('ServiceAccount')).not.toBeInTheDocument()
  })

  it('shows kafka external service configuration and redis resource fields in the service step', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getAllByText('Requests.CPU').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Requests.Memory').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Limits.CPU').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Limits.Memory').length).toBeGreaterThan(0)
    expect(screen.queryByText('对外暴露端口')).not.toBeInTheDocument()
    expect(screen.queryByText('服务名称')).not.toBeInTheDocument()
    expect(screen.queryByText('服务IP')).not.toBeInTheDocument()
    expect(screen.queryByText('服务端口')).not.toBeInTheDocument()
    expect(screen.queryAllByText('存储卷容量')).toHaveLength(0)
    expect(screen.queryByText('Kafka 存储卷容量')).not.toBeInTheDocument()
    expect(screen.queryByText('ZooKeeper 存储卷容量')).not.toBeInTheDocument()
    expect(screen.getAllByText('密码*').length).toBe(2)
  })

  it('blocks leaving the service step until internal mariadb and redis passwords are filled', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getAllByText('密码*').length).toBe(2)

    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByText('MariaDB 密码为必填项。')).toBeInTheDocument()
    expect(screen.getByText('Redis 密码为必填项。')).toBeInTheDocument()
    expect(screen.queryByText('RDS 类型')).not.toBeInTheDocument()

    const passwordInputs = screen
      .getAllByDisplayValue('')
      .filter((element): element is HTMLInputElement => element instanceof HTMLInputElement && element.type === 'password')

    expect(passwordInputs[0]).toHaveFocus()
  })

  it('supports selecting multiple kubernetes master nodes from declared nodes', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '新增节点' }))
    await user.type(screen.getAllByLabelText('Node IPv4')[1], '192.168.40.12')
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByText('Kubernetes Master节点')).toBeInTheDocument()
    expect(screen.getByRole('checkbox', { name: 'node1' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'node2' })).not.toBeChecked()

    await user.click(screen.getByRole('checkbox', { name: 'node2' }))

    expect(screen.getByRole('checkbox', { name: 'node1' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'node2' })).toBeChecked()
  })

  it('limits repository deployment nodes to at most two selected nodes', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '新增节点' }))
    await user.type(screen.getAllByLabelText('Node IPv4')[1], '192.168.40.12')
    await user.click(screen.getByRole('button', { name: '新增节点' }))
    await user.type(screen.getAllByLabelText('Node IPv4')[2], '192.168.40.13')

    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByRole('checkbox', { name: 'node1' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'node2' })).not.toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'node3' })).not.toBeChecked()
    expect(screen.getByText('部署节点最多选择2个')).toBeInTheDocument()

    await user.click(screen.getByRole('checkbox', { name: 'node2' }))
    expect(screen.getByRole('checkbox', { name: 'node1' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'node2' })).toBeChecked()

    await user.click(screen.getByRole('checkbox', { name: 'node3' }))
    expect(screen.getByRole('checkbox', { name: 'node3' })).not.toBeChecked()
  })

  it('keeps per-service internal external selection in the service step', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByLabelText('MariaDB source')).toHaveValue('internal')
    expect(screen.getByLabelText('Redis source')).toHaveValue('internal')
    expect(screen.getByLabelText('OpenSearch source')).toHaveValue('internal')
    expect(screen.getByLabelText('MQ source')).toHaveValue('internal')

    await user.selectOptions(screen.getByLabelText('MariaDB source'), 'external')
    expect(screen.getByLabelText('MariaDB source')).toHaveValue('external')
    expect(screen.getByText('改为外置后，请在连接配置中补充 MariaDB 连接信息。')).toBeInTheDocument()
    expect(screen.queryByLabelText('MariaDB hosts')).not.toBeInTheDocument()
  })

  it('restores old conditional connect forms for external services', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await fillRequiredServicePasswords(user)

    await user.selectOptions(screen.getByLabelText('MariaDB source'), 'external')
    await user.selectOptions(screen.getByLabelText('Redis source'), 'external')
    await user.selectOptions(screen.getByLabelText('OpenSearch source'), 'external')
    await user.selectOptions(screen.getByLabelText('MQ source'), 'external')

    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByRole('heading', { name: 'RDS 类型' })).toBeInTheDocument()
    expect(screen.getByText('Redis 连接模式')).toBeInTheDocument()
    expect(screen.getByText('SearchEngine版本')).toBeInTheDocument()
    expect(screen.getByText('MQ 类型')).toBeInTheDocument()
    expect(screen.getByText('自动化创建数据库')).toBeInTheDocument()
    expect(screen.getByText('发行版')).toBeInTheDocument()

    await user.selectOptions(screen.getByLabelText('Redis connect type'), 'sentinel')
    expect(screen.getByText('哨兵连接信息')).toBeInTheDocument()
    expect(screen.getByLabelText('Sentinel hosts')).toBeInTheDocument()
    expect(screen.getByLabelText('Master group name')).toBeInTheDocument()

    await user.selectOptions(screen.getByLabelText('MQ type'), 'nsq')
    expect(screen.getByText('lookupd 连接信息')).toBeInTheDocument()
    expect(screen.getByLabelText('NSQ lookupd hosts')).toBeInTheDocument()
  })

  it('reveals admin credentials when auto create database is enabled for external rds', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await fillRequiredServicePasswords(user)
    await user.selectOptions(screen.getByLabelText('MariaDB source'), 'external')
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByLabelText('RDS admin username')).toBeInTheDocument()
    expect(screen.getByLabelText('RDS admin password')).toBeInTheDocument()
    await user.click(screen.getByRole('checkbox', { name: '自动化创建数据库' }))
    expect(screen.queryByLabelText('RDS admin username')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('RDS admin password')).not.toBeInTheDocument()
  })

  it('does not render duplicate RDS type text in connect configuration', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await fillRequiredServicePasswords(user)
    await user.selectOptions(screen.getByLabelText('MariaDB source'), 'external')
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getAllByText('RDS 类型')).toHaveLength(1)
  })

  it('does not expose kafka and zookeeper optional toggles in internal mq mode', async () => {
    const user = userEvent.setup()
    render(<Wizard />)

    await startStandardFlow(user)
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))
    await user.click(screen.getByRole('button', { name: '下一步' }))

    expect(screen.getByLabelText('MQ source')).toHaveValue('internal')
    expect(screen.queryByRole('checkbox', { name: '启用 Kafka' })).not.toBeInTheDocument()
    expect(screen.queryByRole('checkbox', { name: '启用 ZooKeeper' })).not.toBeInTheDocument()
    expect(screen.getByText('Kafka 部署节点')).toBeInTheDocument()
    expect(screen.getByText('ZooKeeper 部署节点')).toBeInTheDocument()
  })
})
