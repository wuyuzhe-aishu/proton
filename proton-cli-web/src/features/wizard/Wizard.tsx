import { type Dispatch, type RefObject, type SetStateAction, useMemo, useRef, useState } from 'react'

import { PreviewPanel } from '../preview/PreviewPanel'
import { useSubmit } from '../submit/useSubmit'
import type { WizardState } from '../../schema/config'
import { defaultWizardState } from '../../schema/defaults'
import { validateWizardState } from '../../validation/wizard'

type StorageMode = 'standard' | 'deposit-kubernetes'
type WizardScreen = 'template' | 'steps' | 'success'

type LocalStep = 'node' | 'network' | 'repository' | 'service' | 'connect'
type ManagedStep = 'network' | 'repository' | 'service' | 'connect'
type StepKey = LocalStep | ManagedStep

const localSteps: Array<{ key: LocalStep; title: string; description: string }> = [
  { key: 'node', title: '节点配置', description: '配置集群节点' },
  { key: 'network', title: 'kubernetes配置', description: '设置kubernetes及docker' },
  { key: 'repository', title: '仓库配置', description: '配置容器仓库' },
  { key: 'service', title: '基础服务配置', description: '设置存储服务、消息中间件等' },
  { key: 'connect', title: '连接配置', description: '设置本地、第三方连接配置' },
]

const managedSteps: Array<{ key: ManagedStep; title: string; description: string }> = [
  { key: 'network', title: 'kubernetes配置', description: '设置kubernetes及docker' },
  { key: 'repository', title: '仓库配置', description: '配置容器仓库' },
  { key: 'service', title: '基础服务配置', description: '设置存储服务、消息中间件等' },
  { key: 'connect', title: '连接配置', description: '设置本地、第三方连接配置' },
]

const csAddons = [
  { key: 'ingress-nginx', label: 'ingress-nginx' },
  { key: 'node-exporter', label: 'node-exporter' },
  { key: 'kube-state-metrics', label: 'kube-state-metrics' },
]

function createNode(index: number) {
  return {
    name: `node${index + 1}`,
    ip4: '',
    ip6: '',
  }
}

function replaceNodeReference(hosts: string[], previousName: string, nextName: string) {
  return hosts.map((host) => (host === previousName ? nextName : host))
}

function removeNodeReference(hosts: string[], removedName: string, fallbackName?: string) {
  const filtered = hosts.filter((host) => host !== removedName)

  if (filtered.length > 0 || !fallbackName) {
    return filtered
  }

  return [fallbackName]
}

function syncNodeReferences(current: WizardState, previousName: string, nextName: string) {
  if (!previousName || previousName === nextName) {
    return current
  }

  return {
    ...current,
    chrony:
      current.chrony.mode === 'localmaster'
        ? {
            ...current.chrony,
            server: replaceNodeReference(current.chrony.server, previousName, nextName),
          }
        : current.chrony,
    cs: {
      ...current.cs,
      local: {
        ...current.cs.local,
        master: replaceNodeReference(current.cs.local.master, previousName, nextName),
      },
    },
    cr: {
      ...current.cr,
      local: {
        ...current.cr.local,
        hosts: replaceNodeReference(current.cr.local.hosts, previousName, nextName),
      },
    },
    services: {
      mariadb: {
        ...current.services.mariadb,
        local: {
          ...current.services.mariadb.local,
          hosts: replaceNodeReference(current.services.mariadb.local.hosts, previousName, nextName),
        },
      },
      redis: {
        ...current.services.redis,
        local: {
          ...current.services.redis.local,
          hosts: replaceNodeReference(current.services.redis.local.hosts, previousName, nextName),
        },
      },
      opensearch: {
        ...current.services.opensearch,
        local: {
          ...current.services.opensearch.local,
          hosts: replaceNodeReference(current.services.opensearch.local.hosts, previousName, nextName),
        },
      },
      kafka: {
        ...current.services.kafka,
        local: {
          ...current.services.kafka.local,
          hosts: replaceNodeReference(current.services.kafka.local.hosts, previousName, nextName),
        },
      },
      zookeeper: {
        ...current.services.zookeeper,
        local: {
          ...current.services.zookeeper.local,
          hosts: replaceNodeReference(current.services.zookeeper.local.hosts, previousName, nextName),
        },
      },
      prometheus: {
        ...current.services.prometheus,
        local: {
          ...current.services.prometheus.local,
          hosts: replaceNodeReference(current.services.prometheus.local.hosts, previousName, nextName),
        },
      },
      grafana: {
        ...current.services.grafana,
        local: {
          ...current.services.grafana.local,
          hosts: replaceNodeReference(current.services.grafana.local.hosts, previousName, nextName),
        },
      },
    },
  }
}

function detachNodeReferences(current: WizardState, removedName: string, fallbackName?: string) {
  return {
    ...current,
    chrony:
      current.chrony.mode === 'localmaster'
        ? {
            ...current.chrony,
            server:
              current.chrony.server[0] === removedName && fallbackName
                ? [fallbackName]
                : current.chrony.server.filter((server) => server !== removedName),
          }
        : current.chrony,
    cs: {
      ...current.cs,
      local: {
        ...current.cs.local,
        master: removeNodeReference(current.cs.local.master, removedName, fallbackName),
      },
    },
    cr: {
      ...current.cr,
      local: {
        ...current.cr.local,
        hosts: removeNodeReference(current.cr.local.hosts, removedName, fallbackName),
      },
    },
    services: {
      mariadb: {
        ...current.services.mariadb,
        local: {
          ...current.services.mariadb.local,
          hosts: removeNodeReference(current.services.mariadb.local.hosts, removedName, fallbackName),
        },
      },
      redis: {
        ...current.services.redis,
        local: {
          ...current.services.redis.local,
          hosts: removeNodeReference(current.services.redis.local.hosts, removedName, fallbackName),
        },
      },
      opensearch: {
        ...current.services.opensearch,
        local: {
          ...current.services.opensearch.local,
          hosts: removeNodeReference(current.services.opensearch.local.hosts, removedName, fallbackName),
        },
      },
      kafka: {
        ...current.services.kafka,
        local: {
          ...current.services.kafka.local,
          hosts: removeNodeReference(current.services.kafka.local.hosts, removedName, fallbackName),
        },
      },
      zookeeper: {
        ...current.services.zookeeper,
        local: {
          ...current.services.zookeeper.local,
          hosts: removeNodeReference(current.services.zookeeper.local.hosts, removedName, fallbackName),
        },
      },
      prometheus: {
        ...current.services.prometheus,
        local: {
          ...current.services.prometheus.local,
          hosts: removeNodeReference(current.services.prometheus.local.hosts, removedName, fallbackName),
        },
      },
      grafana: {
        ...current.services.grafana,
        local: {
          ...current.services.grafana.local,
          hosts: removeNodeReference(current.services.grafana.local.hosts, removedName, fallbackName),
        },
      },
    },
  }
}

function splitValues(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function joinValues(values: string[]) {
  return values.join(', ')
}

function hasIngressNginx(addons: string[]) {
  return addons.includes('ingress-nginx')
}

function getCurrentStepIssues(step: StepKey, issues: ReturnType<typeof validateWizardState>['issues']) {
  const prefixes: Record<StepKey, string[]> = {
    node: ['nodes', 'chrony', 'firewall'],
    network: ['cs'],
    repository: ['cr'],
    service: ['services'],
    connect: ['resource_connect_info'],
  }

  return issues.filter((issue) => prefixes[step].some((prefix) => issue.field.startsWith(prefix)))
}

function updateState(setState: Dispatch<SetStateAction<WizardState>>, updater: (current: WizardState) => WizardState) {
  setState((current) => updater(current))
}

function isInternalSource(sourceType: string) {
  return sourceType === 'internal'
}

function setResourceSource(
  current: WizardState,
  resource: 'rds' | 'redis' | 'opensearch' | 'mq',
  sourceType: 'internal' | 'external',
) {
  return {
    ...current,
    resource_connect_info: {
      ...current.resource_connect_info,
      [resource]: {
        ...current.resource_connect_info[resource],
        source_type: sourceType,
        ...(resource === 'mq' && sourceType === 'internal' ? { mq_type: 'kafka' } : {}),
      },
    },
  }
}

function PasswordField({
  value,
  visible,
  onChange,
  onToggleVisibility,
  inputRef,
}: {
  value: string
  visible: boolean
  onChange: (value: string) => void
  onToggleVisibility: () => void
  inputRef?: RefObject<HTMLInputElement | null>
}) {
  return (
    <div className="legacy-password-field">
      <input ref={inputRef} type={visible ? 'text' : 'password'} value={value} onChange={(event) => onChange(event.target.value)} />
      <button type="button" className="legacy-password-toggle" onClick={onToggleVisibility}>
        {visible ? '隐藏' : '显示'}
      </button>
    </div>
  )
}

function TemplateChooser({
  onChoose,
}: {
  onChoose: (mode: StorageMode) => void
}) {
  return (
    <section className="legacy-card legacy-template-card">
      <h2>初始化模板</h2>
      <p className="legacy-muted">保留原部署工具的步骤感，只缩减当前真正还在使用的配置面。</p>
      <div className="legacy-template-grid">
        <button type="button" className="legacy-template-option" onClick={() => onChoose('standard')}>
          <span className="legacy-template-option__title" aria-hidden="true">
            标准模式部署
          </span>
          <span className="legacy-template-option__desc" aria-hidden="true">
            本地 Kubernetes 分支，包含节点配置。
          </span>
          <span className="legacy-sr-only">标准模式部署</span>
        </button>
        <button type="button" className="legacy-template-option" disabled>
          <span className="legacy-template-option__title" aria-hidden="true">
            托管Kubernetes部署
          </span>
          <span className="legacy-template-option__desc" aria-hidden="true">
            功能持续完善中
          </span>
          <span className="legacy-sr-only">托管Kubernetes部署</span>
        </button>
      </div>
    </section>
  )
}

function NodeStep({
  state,
  setState,
}: {
  state: WizardState
  setState: Dispatch<SetStateAction<WizardState>>
}) {
  const canDelete = state.nodes.length > 1

  function addNode() {
    updateState(setState, (current) => ({
      ...current,
      nodes: [...current.nodes, createNode(current.nodes.length)],
    }))
  }

  function updateNodeField(index: number, field: 'name' | 'ip4' | 'ip6', value: string) {
    updateState(setState, (current) => {
      const previousNode = current.nodes[index]
      const nextNode = { ...previousNode, [field]: value }
      const nodes = current.nodes.map((node, nodeIndex) => (nodeIndex === index ? nextNode : node))
      const nextState = { ...current, nodes }

      if (field === 'name') {
        return syncNodeReferences(nextState, previousNode.name, value)
      }

      return nextState
    })
  }

  function removeNode(index: number) {
    updateState(setState, (current) => {
      if (current.nodes.length <= 1) {
        return current
      }

      const removedNode = current.nodes[index]
      const nodes = current.nodes.filter((_, nodeIndex) => nodeIndex !== index)
      const fallbackName = nodes[0]?.name

      return detachNodeReferences(
        {
          ...current,
          nodes,
        },
        removedNode.name,
        fallbackName,
      )
    })
  }

  return (
    <div className="legacy-stack">
      <section className="legacy-card">
        <h3>节点通用配置</h3>
        <div className="legacy-form-grid legacy-form-grid--compact">
          <label>
            <span>时间同步模式</span>
            <select
              value={state.chrony.mode}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  chrony: {
                    ...current.chrony,
                    mode: event.target.value as WizardState['chrony']['mode'],
                    server: event.target.value === 'localmaster' ? [current.cs.local.master[0] ?? current.nodes[0]?.name ?? ''] : [],
                  },
                }))
              }
            >
              <option value="usermanaged">usermanaged</option>
              <option value="localmaster">localmaster</option>
              <option value="externalntp">externalntp</option>
            </select>
          </label>
          <label>
            <span>防火墙模式</span>
            <select
              value={state.firewall.mode}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  firewall: {
                    mode: event.target.value as WizardState['firewall']['mode'],
                  },
                }))
              }
            >
              <option value="usermanaged">usermanaged</option>
              <option value="firewalld">firewalld</option>
            </select>
          </label>
          {state.chrony.mode === 'externalntp' ? (
            <label>
              <span>外部NTP服务器</span>
              <input
                value={joinValues(state.chrony.server)}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    chrony: {
                      ...current.chrony,
                      server: splitValues(event.target.value),
                    },
                  }))
                }
              />
            </label>
          ) : null}
        </div>
      </section>

      <section className="legacy-card">
        <div className="legacy-section-header">
          <div>
            <h3>节点列表</h3>
            <p className="legacy-muted">节点可以连续添加，通用策略放在上方统一配置。</p>
          </div>
          <button type="button" className="legacy-button" onClick={addNode}>
            新增节点
          </button>
        </div>

        <div className="legacy-node-table" role="table" aria-label="节点列表">
          <div className="legacy-node-table__head" role="rowgroup">
            <div className="legacy-node-table__row legacy-node-table__row--head" role="row">
              <div role="columnheader">节点名称</div>
              <div role="columnheader">IPv4地址</div>
              <div role="columnheader">IPv6地址</div>
              <div role="columnheader">操作</div>
            </div>
          </div>
          <div className="legacy-node-table__body" role="rowgroup">
            {state.nodes.map((node, index) => (
              <div key={`${index}-${node.name}`} className="legacy-node-table__row" role="row">
                <div role="cell">
                  <label className="legacy-node-cell">
                    <span className="legacy-sr-only">节点名称</span>
                    <input
                      aria-label="节点名称"
                      value={node.name}
                      onChange={(event) => updateNodeField(index, 'name', event.target.value)}
                    />
                  </label>
                </div>
                <div role="cell">
                  <label className="legacy-node-cell">
                    <span className="legacy-sr-only">Node IPv4</span>
                    <input
                      aria-label="Node IPv4"
                      value={node.ip4}
                      onChange={(event) => updateNodeField(index, 'ip4', event.target.value)}
                    />
                  </label>
                </div>
                <div role="cell">
                  <label className="legacy-node-cell">
                    <span className="legacy-sr-only">IPv6地址</span>
                    <input
                      aria-label="IPv6地址"
                      value={node.ip6}
                      onChange={(event) => updateNodeField(index, 'ip6', event.target.value)}
                    />
                  </label>
                </div>
                <div role="cell">
                  <button type="button" className="legacy-link-button" onClick={() => removeNode(index)} disabled={!canDelete}>
                    删除
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}

function NetworkStep({
  state,
  setState,
}: {
  state: WizardState
  setState: Dispatch<SetStateAction<WizardState>>
}) {
  const isLocal = state.deploymentKind === 'local'

  return (
    <section className="legacy-card">
      <h3>kubernetes配置表单</h3>
      {isLocal ? (
        <>
          <div className="legacy-form-grid">
            <label>
              <span>Kubernetes Master节点</span>
              <span className="legacy-checkbox-panel">
                {state.nodes.map((node) => {
                  const checked = state.cs.local.master.includes(node.name)

                  return (
                    <label key={node.name} className="legacy-checkbox-panel__item">
                      <input
                        type="checkbox"
                        aria-label={node.name}
                        checked={checked}
                        onChange={(event) =>
                          updateState(setState, (current) => {
                            const nextMaster = event.target.checked
                              ? [...current.cs.local.master, node.name]
                              : current.cs.local.master.filter((item) => item !== node.name)

                            const uniqueMaster = Array.from(new Set(nextMaster))
                            const nextPrimaryMaster = uniqueMaster[0]

                            return {
                              ...current,
                              cs: {
                                ...current.cs,
                                local: {
                                  ...current.cs.local,
                                  master: uniqueMaster,
                                },
                              },
                              chrony:
                                current.chrony.mode === 'localmaster'
                                  ? {
                                      ...current.chrony,
                                      server: nextPrimaryMaster ? [nextPrimaryMaster] : [],
                                    }
                                  : current.chrony,
                            }
                          })
                        }
                      />
                      {node.name}
                    </label>
                  )
                })}
              </span>
            </label>
            <label>
              <span>IP协议栈</span>
              <select
                value={state.cs.local.enableDualStack ? 'dual' : state.cs.local.ipFamilies[0] ?? 'IPv4'}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        enableDualStack: event.target.value === 'dual',
                        ipFamilies: event.target.value === 'dual' ? ['IPv4', 'IPv6'] : [event.target.value],
                      },
                    },
                  }))
                }
              >
                <option value="IPv4">IPv4</option>
                <option value="IPv6">IPv6</option>
                <option value="dual">Dual stack</option>
              </select>
            </label>
            <label>
              <span>ha_port</span>
              <input
                type="number"
                value={state.cs.local.ha_port}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        ha_port: Number(event.target.value),
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>docker IP</span>
              <input
                value={state.cs.local.host_network.bip}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        host_network: {
                          ...current.cs.local.host_network,
                          bip: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Pod 网段</span>
              <input
                value={state.cs.local.host_network.pod_network_cidr}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        host_network: {
                          ...current.cs.local.host_network,
                          pod_network_cidr: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Serivce 网段</span>
              <input
                value={state.cs.local.host_network.service_cidr}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        host_network: {
                          ...current.cs.local.host_network,
                          service_cidr: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            {state.cs.local.enableDualStack ? (
              <>
                <label>
                  <span>IPv4 网卡</span>
                  <input
                    value={state.cs.local.host_network.ipv4_interface}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          local: {
                            ...current.cs.local,
                            host_network: {
                              ...current.cs.local.host_network,
                              ipv4_interface: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>IPv6 网卡</span>
                  <input
                    value={state.cs.local.host_network.ipv6_interface}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          local: {
                            ...current.cs.local,
                            host_network: {
                              ...current.cs.local.host_network,
                              ipv6_interface: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
              </>
            ) : null}
            <label>
              <span>etcd 数据路径</span>
              <input
                value={state.cs.local.etcd_data_dir}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        etcd_data_dir: event.target.value,
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>docker 数据路径</span>
              <input
                value={state.cs.local.docker_data_dir}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      local: {
                        ...current.cs.local,
                        docker_data_dir: event.target.value,
                      },
                    },
                  }))
                }
              />
            </label>
          </div>

          <div className="legacy-service-block">
            <h4>可选插件</h4>
            <div className="legacy-checkbox-row">
              {csAddons.map((addon) => (
                <label key={addon.key}>
                  <input
                    type="checkbox"
                    checked={state.cs.local.addons.includes(addon.key)}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          local: {
                            ...current.cs.local,
                            addons: event.target.checked
                              ? [...current.cs.local.addons, addon.key]
                              : current.cs.local.addons.filter((item) => item !== addon.key),
                          },
                        },
                      }))
                    }
                  />
                  {addon.label}
                </label>
              ))}
            </div>
            {hasIngressNginx(state.cs.local.addons) ? (
              <div className="legacy-form-grid">
                <label>
                  <span>ingress-nginx HTTP 端口</span>
                  <input
                    type="number"
                    value={state.cs.local.ingressNginx.httpPort}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          local: {
                            ...current.cs.local,
                            ingressNginx: {
                              ...current.cs.local.ingressNginx,
                              httpPort: Number(event.target.value),
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>ingress-nginx HTTPS 端口</span>
                  <input
                    type="number"
                    value={state.cs.local.ingressNginx.httpsPort}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          local: {
                            ...current.cs.local,
                            ingressNginx: {
                              ...current.cs.local.ingressNginx,
                              httpsPort: Number(event.target.value),
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
              </div>
            ) : null}
          </div>
        </>
      ) : (
        <>
          <div className="legacy-form-grid">
            <label>
              <span>命名空间</span>
              <input
                value={state.cs.managed.namespace}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      managed: {
                        ...current.cs.managed,
                        namespace: event.target.value,
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>ServiceAccount</span>
              <input
                value={state.cs.managed.serviceaccount}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cs: {
                      ...current.cs,
                      managed: {
                        ...current.cs.managed,
                        serviceaccount: event.target.value,
                      },
                    },
                  }))
                }
              />
            </label>
          </div>

          <div className="legacy-service-block">
            <h4>可选插件</h4>
            <div className="legacy-checkbox-row">
              {csAddons.map((addon) => (
                <label key={addon.key}>
                  <input
                    type="checkbox"
                    checked={state.cs.managed.addons.includes(addon.key)}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          managed: {
                            ...current.cs.managed,
                            addons: event.target.checked
                              ? [...current.cs.managed.addons, addon.key]
                              : current.cs.managed.addons.filter((item) => item !== addon.key),
                          },
                        },
                      }))
                    }
                  />
                  {addon.label}
                </label>
              ))}
            </div>
            {hasIngressNginx(state.cs.managed.addons) ? (
              <div className="legacy-form-grid">
                <label>
                  <span>ingress-nginx HTTP 端口</span>
                  <input
                    type="number"
                    value={state.cs.managed.ingressNginx.httpPort}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          managed: {
                            ...current.cs.managed,
                            ingressNginx: {
                              ...current.cs.managed.ingressNginx,
                              httpPort: Number(event.target.value),
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>ingress-nginx HTTPS 端口</span>
                  <input
                    type="number"
                    value={state.cs.managed.ingressNginx.httpsPort}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cs: {
                          ...current.cs,
                          managed: {
                            ...current.cs.managed,
                            ingressNginx: {
                              ...current.cs.managed.ingressNginx,
                              httpsPort: Number(event.target.value),
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
              </div>
            ) : null}
          </div>
        </>
      )}
    </section>
  )
}

function RepositoryStep({
  state,
  setState,
}: {
  state: WizardState
  setState: Dispatch<SetStateAction<WizardState>>
}) {
  const isLocal = state.deploymentKind === 'local'

  return (
    <section className="legacy-card">
      <h3>仓库配置表单</h3>
      {isLocal ? (
        <>
          <div className="legacy-form-grid">
            <label>
              <span>部署节点</span>
              <span className="legacy-checkbox-panel">
                {state.nodes.map((node) => {
                  const checked = state.cr.local.hosts.includes(node.name)
                  const limitReached = state.cr.local.hosts.length >= 2 && !checked

                  return (
                    <label key={node.name} className="legacy-checkbox-panel__item">
                      <input
                        type="checkbox"
                        aria-label={node.name}
                        checked={checked}
                        disabled={limitReached}
                        onChange={(event) =>
                          updateState(setState, (current) => {
                            const nextHosts = event.target.checked
                              ? [...current.cr.local.hosts, node.name]
                              : current.cr.local.hosts.filter((item) => item !== node.name)

                            return {
                              ...current,
                              cr: {
                                ...current.cr,
                                local: {
                                  ...current.cr.local,
                                  hosts: Array.from(new Set(nextHosts)).slice(0, 2),
                                },
                              },
                            }
                          })
                        }
                      />
                      {node.name}
                    </label>
                  )
                })}
              </span>
              <span className="legacy-field-hint">部署节点最多选择2个</span>
            </label>
            <label>
              <span>chart与image的存储路径</span>
              <input
                value={state.cr.local.storage}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cr: {
                      ...current.cr,
                      local: {
                        ...current.cr.local,
                        storage: event.target.value,
                      },
                    },
                  }))
                }
              />
            </label>
          </div>

          <div className="legacy-service-block">
            <h4>端口配置</h4>
            <div className="legacy-form-grid">
              <label>
                <span>Chart仓库端口</span>
                <input
                  type="number"
                  value={state.cr.local.ports.chartmuseum}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      cr: {
                        ...current.cr,
                        local: {
                          ...current.cr.local,
                          ports: {
                            ...current.cr.local.ports,
                            chartmuseum: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              <label>
                <span>Registry端口</span>
                <input
                  type="number"
                  value={state.cr.local.ports.registry}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      cr: {
                        ...current.cr,
                        local: {
                          ...current.cr.local,
                          ports: {
                            ...current.cr.local.ports,
                            registry: Number(event.target.value),
                            cr_manager: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            </div>
          </div>

          <div className="legacy-service-block">
            <h4>高可用端口配置</h4>
            <div className="legacy-form-grid">
              <label>
                <span>Chart仓库高可用端口</span>
                <input
                  type="number"
                  value={state.cr.local.ha_ports.chartmuseum}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      cr: {
                        ...current.cr,
                        local: {
                          ...current.cr.local,
                          ha_ports: {
                            ...current.cr.local.ha_ports,
                            chartmuseum: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              <label>
                <span>Registry高可用端口</span>
                <input
                  type="number"
                  value={state.cr.local.ha_ports.registry}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      cr: {
                        ...current.cr,
                        local: {
                          ...current.cr.local,
                          ha_ports: {
                            ...current.cr.local.ha_ports,
                            registry: Number(event.target.value),
                            cr_manager: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            </div>
          </div>
        </>
      ) : (
        <>
          <div className="legacy-form-grid">
            <label>
              <span>镜像仓库</span>
              <select
                value={state.cr.external.image_repository}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cr: {
                      ...current.cr,
                      external: {
                        ...current.cr.external,
                        image_repository: event.target.value as WizardState['cr']['external']['image_repository'],
                      },
                    },
                  }))
                }
              >
                <option value="registry">registry</option>
                <option value="oci">oci</option>
              </select>
            </label>
            <label>
              <span>Chart仓库</span>
              <select
                value={state.cr.external.chart_repository}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    cr: {
                      ...current.cr,
                      external: {
                        ...current.cr.external,
                        chart_repository: event.target.value as WizardState['cr']['external']['chart_repository'],
                      },
                    },
                  }))
                }
              >
                <option value="chartmuseum">chartmuseum</option>
                <option value="oci">oci</option>
              </select>
            </label>
          </div>

          {state.cr.external.image_repository === 'registry' ? (
            <div className="legacy-service-block">
              <h4>registry</h4>
              <div className="legacy-form-grid">
                <label>
                  <span>Registry地址</span>
                  <input
                    value={state.cr.external.registry.host}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            registry: {
                              ...current.cr.external.registry,
                              host: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>Registry用户名</span>
                  <input
                    value={state.cr.external.registry.username}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            registry: {
                              ...current.cr.external.registry,
                              username: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>Registry密码</span>
                  <input
                    type="password"
                    value={state.cr.external.registry.password}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            registry: {
                              ...current.cr.external.registry,
                              password: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
              </div>
            </div>
          ) : null}

          {state.cr.external.chart_repository === 'chartmuseum' ? (
            <div className="legacy-service-block">
              <h4>chartmuseum</h4>
              <div className="legacy-form-grid">
                <label>
                  <span>Chartmuseum地址</span>
                  <input
                    value={state.cr.external.chartmuseum.host}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            chartmuseum: {
                              ...current.cr.external.chartmuseum,
                              host: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>Chartmuseum用户名</span>
                  <input
                    value={state.cr.external.chartmuseum.username}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            chartmuseum: {
                              ...current.cr.external.chartmuseum,
                              username: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>Chartmuseum密码</span>
                  <input
                    type="password"
                    value={state.cr.external.chartmuseum.password}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            chartmuseum: {
                              ...current.cr.external.chartmuseum,
                              password: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
              </div>
            </div>
          ) : null}

          {state.cr.external.image_repository === 'oci' || state.cr.external.chart_repository === 'oci' ? (
            <div className="legacy-service-block">
              <h4>oci</h4>
              <div className="legacy-form-grid">
                <label>
                  <span>OCI registry</span>
                  <input
                    value={state.cr.external.oci.registry}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            oci: {
                              ...current.cr.external.oci,
                              registry: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>OCI用户名</span>
                  <input
                    value={state.cr.external.oci.username}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            oci: {
                              ...current.cr.external.oci,
                              username: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label>
                  <span>OCI密码</span>
                  <input
                    type="password"
                    value={state.cr.external.oci.password}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            oci: {
                              ...current.cr.external.oci,
                              password: event.target.value,
                            },
                          },
                        },
                      }))
                    }
                  />
                </label>
                <label className="legacy-inline-checkbox">
                  <input
                    type="checkbox"
                    checked={state.cr.external.oci.plain_http}
                    onChange={(event) =>
                      updateState(setState, (current) => ({
                        ...current,
                        cr: {
                          ...current.cr,
                          external: {
                            ...current.cr.external,
                            oci: {
                              ...current.cr.external.oci,
                              plain_http: event.target.checked,
                            },
                          },
                        },
                      }))
                    }
                  />
                  plain_http
                </label>
              </div>
            </div>
          ) : null}
        </>
      )}
    </section>
  )
}

function ServiceStep({
  state,
  setState,
  passwordVisibility,
  onTogglePasswordVisibility,
  mariadbPasswordRef,
  redisPasswordRef,
}: {
  state: WizardState
  setState: Dispatch<SetStateAction<WizardState>>
  passwordVisibility: {
    mariadb: boolean
    redis: boolean
  }
  onTogglePasswordVisibility: (field: 'mariadb' | 'redis') => void
  mariadbPasswordRef: RefObject<HTMLInputElement | null>
  redisPasswordRef: RefObject<HTMLInputElement | null>
}) {
  const isLocal = state.deploymentKind === 'local'
  const serviceMode = isLocal ? 'local' : 'managed'
  const mariadb = state.services.mariadb[serviceMode]
  const redis = state.services.redis[serviceMode]
  const opensearch = state.services.opensearch[serviceMode]
  const kafka = state.services.kafka[serviceMode]
  const zookeeper = state.services.zookeeper[serviceMode]
  const rdsInternal = isInternalSource(state.resource_connect_info.rds.source_type)
  const redisInternal = isInternalSource(state.resource_connect_info.redis.source_type)
  const searchInternal = isInternalSource(state.resource_connect_info.opensearch.source_type)
  const mqInternal = isInternalSource(state.resource_connect_info.mq.source_type)

  function renderExternalHint(label: string) {
    return (
      <div className="legacy-inline-note">
        <strong>{label} 使用第三方资源</strong>
        <p>改为外置后，请在连接配置中补充 {label} 连接信息。</p>
      </div>
    )
  }

  return (
    <section className="legacy-card">
      <h3>基础服务配置表单</h3>

      <div className="legacy-service-block">
        <h4>MariaDB</h4>
        <div className="legacy-form-grid">
          <label>
            <span>资源类型</span>
            <select
              aria-label="MariaDB source"
              value={state.resource_connect_info.rds.source_type}
              onChange={(event) =>
                updateState(setState, (current) => setResourceSource(current, 'rds', event.target.value as 'internal' | 'external'))
              }
            >
              <option value="internal">internal</option>
              <option value="external">external</option>
            </select>
          </label>
        </div>
        {rdsInternal ? (
          <div className="legacy-form-grid">
            {isLocal ? (
              <label>
                <span>部署节点</span>
                <span className="legacy-checkbox-panel">
                  {state.nodes.map((node) => {
                    const checked = state.services.mariadb.local.hosts.includes(node.name)

                    return (
                      <label key={node.name} className="legacy-checkbox-panel__item">
                        <input
                          type="checkbox"
                          aria-label={node.name}
                          checked={checked}
                          onChange={(event) =>
                            updateState(setState, (current) => {
                              const nextHosts = event.target.checked
                                ? [...current.services.mariadb.local.hosts, node.name]
                                : current.services.mariadb.local.hosts.filter((item) => item !== node.name)

                              return {
                                ...current,
                                services: {
                                  ...current.services,
                                  mariadb: {
                                    ...current.services.mariadb,
                                    local: {
                                      ...current.services.mariadb.local,
                                      hosts: Array.from(new Set(nextHosts)),
                                    },
                                  },
                                },
                              }
                            })
                          }
                        />
                        {node.name}
                      </label>
                    )
                  })}
                </span>
              </label>
            ) : (
              <label>
                <span>副本数</span>
                <input
                  aria-label="MariaDB replicas"
                  type="number"
                  min={1}
                  value={state.services.mariadb.managed.replica_count}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        mariadb: {
                          ...current.services.mariadb,
                          managed: {
                            ...current.services.mariadb.managed,
                            replica_count: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            )}
            <label>
              <span>Innodb_buffer_size</span>
              <input
                value={mariadb.config.innodb_buffer_pool_size}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      mariadb: {
                        ...current.services.mariadb,
                        [serviceMode]: {
                          ...current.services.mariadb[serviceMode],
                          config: {
                            ...current.services.mariadb[serviceMode].config,
                            innodb_buffer_pool_size: event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Requests.Memory</span>
              <input
                value={mariadb.config.resource_requests_memory}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      mariadb: {
                        ...current.services.mariadb,
                        [serviceMode]: {
                          ...current.services.mariadb[serviceMode],
                          config: {
                            ...current.services.mariadb[serviceMode].config,
                            resource_requests_memory: event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Limits.Memory</span>
              <input
                value={mariadb.config.resource_limits_memory}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      mariadb: {
                        ...current.services.mariadb,
                        [serviceMode]: {
                          ...current.services.mariadb[serviceMode],
                          config: {
                            ...current.services.mariadb[serviceMode].config,
                            resource_limits_memory: event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>用户名（管理权限）</span>
              <input
                value={mariadb.admin_user}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      mariadb: {
                        ...current.services.mariadb,
                        [serviceMode]: {
                          ...current.services.mariadb[serviceMode],
                          admin_user: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>密码*</span>
              <PasswordField
                inputRef={mariadbPasswordRef}
                value={mariadb.admin_passwd}
                visible={passwordVisibility.mariadb}
                onToggleVisibility={() => onTogglePasswordVisibility('mariadb')}
                onChange={(value) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      mariadb: {
                        ...current.services.mariadb,
                        [serviceMode]: {
                          ...current.services.mariadb[serviceMode],
                          admin_passwd: value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            {isLocal ? (
              <label>
                <span>数据路径</span>
                <input
                  value={state.services.mariadb.local.data_path}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        mariadb: {
                          ...current.services.mariadb,
                          local: {
                            ...current.services.mariadb.local,
                            data_path: event.target.value,
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            ) : (
              <label>
                <span>storageClassName</span>
                <input
                  value={state.services.mariadb.managed.storageClassName}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        mariadb: {
                          ...current.services.mariadb,
                          managed: {
                            ...current.services.mariadb.managed,
                            storageClassName: event.target.value,
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            )}
          </div>
        ) : (
          renderExternalHint('MariaDB')
        )}
      </div>

      <div className="legacy-service-block">
        <h4>Redis</h4>
        <div className="legacy-form-grid">
          <label>
            <span>资源类型</span>
            <select
              aria-label="Redis source"
              value={state.resource_connect_info.redis.source_type}
              onChange={(event) =>
                updateState(setState, (current) =>
                  setResourceSource(current, 'redis', event.target.value as 'internal' | 'external'),
                )
              }
            >
              <option value="internal">internal</option>
              <option value="external">external</option>
            </select>
          </label>
        </div>
        {redisInternal ? (
          <div className="legacy-form-grid">
            {isLocal ? (
              <label>
                <span>部署节点</span>
                <span className="legacy-checkbox-panel">
                  {state.nodes.map((node) => {
                    const checked = state.services.redis.local.hosts.includes(node.name)

                    return (
                      <label key={node.name} className="legacy-checkbox-panel__item">
                        <input
                          type="checkbox"
                          aria-label={node.name}
                          checked={checked}
                          onChange={(event) =>
                            updateState(setState, (current) => {
                              const nextHosts = event.target.checked
                                ? [...current.services.redis.local.hosts, node.name]
                                : current.services.redis.local.hosts.filter((item) => item !== node.name)

                              return {
                                ...current,
                                services: {
                                  ...current.services,
                                  redis: {
                                    ...current.services.redis,
                                    local: {
                                      ...current.services.redis.local,
                                      hosts: Array.from(new Set(nextHosts)),
                                    },
                                  },
                                },
                              }
                            })
                          }
                        />
                        {node.name}
                      </label>
                    )
                  })}
                </span>
              </label>
            ) : (
              <label>
                <span>副本数</span>
                <input
                  type="number"
                  min={1}
                  value={state.services.redis.managed.replica_count}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        redis: {
                          ...current.services.redis,
                          managed: {
                            ...current.services.redis.managed,
                            replica_count: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            )}
            <label>
              <span>用户名（管理权限）</span>
              <input
                value={redis.admin_user}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      redis: {
                        ...current.services.redis,
                        [serviceMode]: {
                          ...current.services.redis[serviceMode],
                          admin_user: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>密码*</span>
              <PasswordField
                inputRef={redisPasswordRef}
                value={redis.admin_passwd}
                visible={passwordVisibility.redis}
                onToggleVisibility={() => onTogglePasswordVisibility('redis')}
                onChange={(value) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      redis: {
                        ...current.services.redis,
                        [serviceMode]: {
                          ...current.services.redis[serviceMode],
                          admin_passwd: value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Requests.CPU</span>
              <input
                value={redis.resources?.requests?.cpu ?? ''}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      redis: {
                        ...current.services.redis,
                        [serviceMode]: {
                          ...current.services.redis[serviceMode],
                          resources: {
                            ...current.services.redis[serviceMode].resources,
                            requests: {
                              ...current.services.redis[serviceMode].resources?.requests,
                              cpu: event.target.value,
                            },
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Requests.Memory</span>
              <input
                value={redis.resources?.requests?.memory ?? ''}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      redis: {
                        ...current.services.redis,
                        [serviceMode]: {
                          ...current.services.redis[serviceMode],
                          resources: {
                            ...current.services.redis[serviceMode].resources,
                            requests: {
                              ...current.services.redis[serviceMode].resources?.requests,
                              memory: event.target.value,
                            },
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Limits.CPU</span>
              <input
                value={redis.resources?.limits?.cpu ?? ''}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      redis: {
                        ...current.services.redis,
                        [serviceMode]: {
                          ...current.services.redis[serviceMode],
                          resources: {
                            ...current.services.redis[serviceMode].resources,
                            limits: {
                              ...current.services.redis[serviceMode].resources?.limits,
                              cpu: event.target.value,
                            },
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>Limits.Memory</span>
              <input
                value={redis.resources?.limits?.memory ?? ''}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      redis: {
                        ...current.services.redis,
                        [serviceMode]: {
                          ...current.services.redis[serviceMode],
                          resources: {
                            ...current.services.redis[serviceMode].resources,
                            limits: {
                              ...current.services.redis[serviceMode].resources?.limits,
                              memory: event.target.value,
                            },
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            {isLocal ? (
              <label>
                <span>数据路径</span>
                <input
                  value={state.services.redis.local.data_path}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        redis: {
                          ...current.services.redis,
                          local: {
                            ...current.services.redis.local,
                            data_path: event.target.value,
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            ) : (
              <label>
                <span>storageClassName</span>
                <input
                  value={state.services.redis.managed.storageClassName}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        redis: {
                          ...current.services.redis,
                          managed: {
                            ...current.services.redis.managed,
                            storageClassName: event.target.value,
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            )}
          </div>
        ) : (
          renderExternalHint('Redis')
        )}
      </div>

      <div className="legacy-service-block">
        <h4>OpenSearch</h4>
        <div className="legacy-form-grid">
          <label>
            <span>资源类型</span>
            <select
              aria-label="OpenSearch source"
              value={state.resource_connect_info.opensearch.source_type}
              onChange={(event) =>
                updateState(setState, (current) =>
                  setResourceSource(current, 'opensearch', event.target.value as 'internal' | 'external'),
                )
              }
            >
              <option value="internal">internal</option>
              <option value="external">external</option>
            </select>
          </label>
        </div>
        {searchInternal ? (
          <div className="legacy-form-grid">
            {isLocal ? (
              <label>
                <span>部署节点</span>
                <span className="legacy-checkbox-panel">
                  {state.nodes.map((node) => {
                    const checked = state.services.opensearch.local.hosts.includes(node.name)

                    return (
                      <label key={node.name} className="legacy-checkbox-panel__item">
                        <input
                          type="checkbox"
                          aria-label={node.name}
                          checked={checked}
                          onChange={(event) =>
                            updateState(setState, (current) => {
                              const nextHosts = event.target.checked
                                ? [...current.services.opensearch.local.hosts, node.name]
                                : current.services.opensearch.local.hosts.filter((item) => item !== node.name)

                              return {
                                ...current,
                                services: {
                                  ...current.services,
                                  opensearch: {
                                    ...current.services.opensearch,
                                    local: {
                                      ...current.services.opensearch.local,
                                      hosts: Array.from(new Set(nextHosts)),
                                    },
                                  },
                                },
                              }
                            })
                          }
                        />
                        {node.name}
                      </label>
                    )
                  })}
                </span>
              </label>
            ) : (
              <label>
                <span>副本数</span>
                <input
                  type="number"
                  min={1}
                  value={state.services.opensearch.managed.replica_count}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        opensearch: {
                          ...current.services.opensearch,
                          managed: {
                            ...current.services.opensearch.managed,
                            replica_count: Number(event.target.value),
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            )}
            <label>
              <span>模式</span>
              <input
                value={opensearch.mode}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          mode: event.target.value as typeof opensearch.mode,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>JVM配置</span>
              <input
                value={opensearch.config.jvmOptions}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          config: {
                            ...current.services.opensearch[serviceMode].config,
                            jvmOptions: event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            {isLocal ? (
              <label>
                <span>数据路径</span>
                <input
                  value={state.services.opensearch.local.data_path}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        opensearch: {
                          ...current.services.opensearch,
                          local: {
                            ...current.services.opensearch.local,
                            data_path: event.target.value,
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            ) : (
              <label>
                <span>storageClassName</span>
                <input
                  value={state.services.opensearch.managed.storageClassName}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        opensearch: {
                          ...current.services.opensearch,
                          managed: {
                            ...current.services.opensearch.managed,
                            storageClassName: event.target.value,
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
            )}
            <label>
              <span>远程词库</span>
              <input
                value={opensearch.config.hanlpRemoteextDict ?? ''}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          config: {
                            ...current.services.opensearch[serviceMode].config,
                            hanlpRemoteextDict: event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>去停词</span>
              <input
                value={opensearch.config.hanlpRemoteextStopwords ?? ''}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          config: {
                            ...current.services.opensearch[serviceMode].config,
                            hanlpRemoteextStopwords: event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>低警戒水位线</span>
              <input
                value={opensearch.settings['cluster.routing.allocation.disk.watermark.low']}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          settings: {
                            ...current.services.opensearch[serviceMode].settings,
                            'cluster.routing.allocation.disk.watermark.low': event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>高警戒水位线</span>
              <input
                value={opensearch.settings['cluster.routing.allocation.disk.watermark.high']}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          settings: {
                            ...current.services.opensearch[serviceMode].settings,
                            'cluster.routing.allocation.disk.watermark.high': event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
            <label>
              <span>洪泛警戒水位线</span>
              <input
                value={opensearch.settings['cluster.routing.allocation.disk.watermark.flood_stage']}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      opensearch: {
                        ...current.services.opensearch,
                        [serviceMode]: {
                          ...current.services.opensearch[serviceMode],
                          settings: {
                            ...current.services.opensearch[serviceMode].settings,
                            'cluster.routing.allocation.disk.watermark.flood_stage': event.target.value,
                          },
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          </div>
        ) : (
          renderExternalHint('OpenSearch')
        )}
      </div>

      <div className="legacy-service-block">
        <h4>Kafka / ZooKeeper</h4>
        <div className="legacy-form-grid">
          <label>
            <span>资源类型</span>
            <select
              aria-label="MQ source"
              value={state.resource_connect_info.mq.source_type}
              onChange={(event) =>
                updateState(setState, (current) => setResourceSource(current, 'mq', event.target.value as 'internal' | 'external'))
              }
            >
              <option value="internal">internal</option>
              <option value="external">external</option>
            </select>
          </label>
        </div>
        {mqInternal ? (
          <>
            <div className="legacy-form-grid legacy-service-grid">
              {isLocal ? (
                <>
                  <label>
                    <span>Kafka 部署节点</span>
                    <span className="legacy-checkbox-panel">
                      {state.nodes.map((node) => {
                        const checked = state.services.kafka.local.hosts.includes(node.name)

                        return (
                          <label key={node.name} className="legacy-checkbox-panel__item">
                            <input
                              type="checkbox"
                              aria-label={node.name}
                              checked={checked}
                              onChange={(event) =>
                                updateState(setState, (current) => {
                                  const nextHosts = event.target.checked
                                    ? [...current.services.kafka.local.hosts, node.name]
                                    : current.services.kafka.local.hosts.filter((item) => item !== node.name)

                                  return {
                                    ...current,
                                    services: {
                                      ...current.services,
                                      kafka: {
                                        ...current.services.kafka,
                                        local: {
                                          ...current.services.kafka.local,
                                          hosts: Array.from(new Set(nextHosts)),
                                        },
                                      },
                                    },
                                  }
                                })
                              }
                            />
                            {node.name}
                          </label>
                        )
                      })}
                    </span>
                  </label>
                  <label>
                    <span>ZooKeeper 部署节点</span>
                    <span className="legacy-checkbox-panel">
                      {state.nodes.map((node) => {
                        const checked = state.services.zookeeper.local.hosts.includes(node.name)

                        return (
                          <label key={node.name} className="legacy-checkbox-panel__item">
                            <input
                              type="checkbox"
                              aria-label={node.name}
                              checked={checked}
                              onChange={(event) =>
                                updateState(setState, (current) => {
                                  const nextHosts = event.target.checked
                                    ? [...current.services.zookeeper.local.hosts, node.name]
                                    : current.services.zookeeper.local.hosts.filter((item) => item !== node.name)

                                  return {
                                    ...current,
                                    services: {
                                      ...current.services,
                                      zookeeper: {
                                        ...current.services.zookeeper,
                                        local: {
                                          ...current.services.zookeeper.local,
                                          hosts: Array.from(new Set(nextHosts)),
                                        },
                                      },
                                    },
                                  }
                                })
                              }
                            />
                            {node.name}
                          </label>
                        )
                      })}
                    </span>
                  </label>
                </>
              ) : (
                <>
                  <label>
                    <span>Kafka 副本数</span>
                    <input
                      type="number"
                      min={1}
                      value={state.services.kafka.managed.replica_count}
                      onChange={(event) =>
                        updateState(setState, (current) => ({
                          ...current,
                          services: {
                            ...current.services,
                            kafka: {
                              ...current.services.kafka,
                              managed: {
                                ...current.services.kafka.managed,
                                replica_count: Number(event.target.value),
                              },
                            },
                          },
                        }))
                      }
                    />
                  </label>
                  <label>
                    <span>ZooKeeper 副本数</span>
                    <input
                      type="number"
                      min={1}
                      value={state.services.zookeeper.managed.replica_count}
                      onChange={(event) =>
                        updateState(setState, (current) => ({
                          ...current,
                          services: {
                            ...current.services,
                            zookeeper: {
                              ...current.services.zookeeper,
                              managed: {
                                ...current.services.zookeeper.managed,
                                replica_count: Number(event.target.value),
                              },
                            },
                          },
                        }))
                      }
                    />
                  </label>
                </>
              )}
              <label>
                <span>Kafka JVM配置</span>
                <input
                  value={kafka.env.KAFKA_HEAP_OPTS}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        kafka: {
                          ...current.services.kafka,
                          [serviceMode]: {
                            ...current.services.kafka[serviceMode],
                            env: {
                              ...current.services.kafka[serviceMode].env,
                              KAFKA_HEAP_OPTS: event.target.value,
                            },
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              <label>
                <span>日志保留字节数</span>
                <input
                  value={kafka.env.KAFKA_LOG_RETENTION_BYTES ?? ''}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        kafka: {
                          ...current.services.kafka,
                          [serviceMode]: {
                            ...current.services.kafka[serviceMode],
                            env: {
                              ...current.services.kafka[serviceMode].env,
                              KAFKA_LOG_RETENTION_BYTES: event.target.value,
                            },
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              <label>
                <span>日志保留小时数</span>
                <input
                  value={kafka.env.KAFKA_LOG_RETENTION_HOURS ?? ''}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        kafka: {
                          ...current.services.kafka,
                          [serviceMode]: {
                            ...current.services.kafka[serviceMode],
                            env: {
                              ...current.services.kafka[serviceMode].env,
                              KAFKA_LOG_RETENTION_HOURS: event.target.value,
                            },
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              <label>
                <span>日志段最大小时数</span>
                <input
                  value={kafka.env.KAFKA_LOG_ROLL_HOURS ?? ''}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        kafka: {
                          ...current.services.kafka,
                          [serviceMode]: {
                            ...current.services.kafka[serviceMode],
                            env: {
                              ...current.services.kafka[serviceMode].env,
                              KAFKA_LOG_ROLL_HOURS: event.target.value,
                            },
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              <label>
                <span>ZooKeeper JVM配置</span>
                <input
                  value={zookeeper.env.JVMFLAGS}
                  onChange={(event) =>
                    updateState(setState, (current) => ({
                      ...current,
                      services: {
                        ...current.services,
                        zookeeper: {
                          ...current.services.zookeeper,
                          [serviceMode]: {
                            ...current.services.zookeeper[serviceMode],
                            env: {
                              ...current.services.zookeeper[serviceMode].env,
                              JVMFLAGS: event.target.value,
                            },
                          },
                        },
                      },
                    }))
                  }
                />
              </label>
              {isLocal ? (
                <>
                  <label>
                    <span>Kafka 数据路径</span>
                    <input
                      value={state.services.kafka.local.data_path}
                      onChange={(event) =>
                        updateState(setState, (current) => ({
                          ...current,
                          services: {
                            ...current.services,
                            kafka: {
                              ...current.services.kafka,
                              local: {
                                ...current.services.kafka.local,
                                data_path: event.target.value,
                              },
                            },
                          },
                        }))
                      }
                    />
                  </label>
                  <label>
                    <span>ZooKeeper 数据路径</span>
                    <input
                      value={state.services.zookeeper.local.data_path}
                      onChange={(event) =>
                        updateState(setState, (current) => ({
                          ...current,
                          services: {
                            ...current.services,
                            zookeeper: {
                              ...current.services.zookeeper,
                              local: {
                                ...current.services.zookeeper.local,
                                data_path: event.target.value,
                              },
                            },
                          },
                        }))
                      }
                    />
                  </label>
                </>
              ) : (
                <>
                  <label>
                    <span>Kafka storageClassName</span>
                    <input
                      value={state.services.kafka.managed.storageClassName}
                      onChange={(event) =>
                        updateState(setState, (current) => ({
                          ...current,
                          services: {
                            ...current.services,
                            kafka: {
                              ...current.services.kafka,
                              managed: {
                                ...current.services.kafka.managed,
                                storageClassName: event.target.value,
                              },
                            },
                          },
                        }))
                      }
                    />
                  </label>
                  <label>
                    <span>ZooKeeper storageClassName</span>
                    <input
                      value={state.services.zookeeper.managed.storageClassName}
                      onChange={(event) =>
                        updateState(setState, (current) => ({
                          ...current,
                          services: {
                            ...current.services,
                            zookeeper: {
                              ...current.services.zookeeper,
                              managed: {
                                ...current.services.zookeeper.managed,
                                storageClassName: event.target.value,
                              },
                            },
                          },
                        }))
                      }
                    />
                  </label>
                </>
              )}
            </div>

          </>
        ) : (
          renderExternalHint('MQ')
        )}
      </div>

      <div className="legacy-service-block">
        <h4>Prometheus</h4>
        <div className="legacy-form-grid">
          {isLocal ? (
            <label>
              <span>部署节点</span>
              <span className="legacy-checkbox-panel">
                {state.nodes.map((node) => {
                  const checked = state.services.prometheus.local.hosts.includes(node.name)

                  return (
                    <label key={node.name} className="legacy-checkbox-panel__item">
                      <input
                        type="checkbox"
                        aria-label={node.name}
                        checked={checked}
                        onChange={(event) =>
                          updateState(setState, (current) => {
                            const nextHosts = event.target.checked
                              ? [...current.services.prometheus.local.hosts, node.name]
                              : current.services.prometheus.local.hosts.filter((item) => item !== node.name)

                            return {
                              ...current,
                              services: {
                                ...current.services,
                                prometheus: {
                                  ...current.services.prometheus,
                                  local: {
                                    ...current.services.prometheus.local,
                                    hosts: Array.from(new Set(nextHosts)),
                                  },
                                },
                              },
                            }
                          })
                        }
                      />
                      {node.name}
                    </label>
                  )
                })}
              </span>
            </label>
          ) : (
            <label>
              <span>副本数</span>
              <input
                type="number"
                min={1}
                value={state.services.prometheus.managed.replica_count}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      prometheus: {
                        ...current.services.prometheus,
                        managed: {
                          ...current.services.prometheus.managed,
                          replica_count: Number(event.target.value),
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          )}
          <label>
            <span>Requests.CPU</span>
            <input
              value={state.services.prometheus[serviceMode].resources?.requests?.cpu ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    prometheus: {
                      ...current.services.prometheus,
                      [serviceMode]: {
                        ...current.services.prometheus[serviceMode],
                        resources: {
                          ...current.services.prometheus[serviceMode].resources,
                          requests: {
                            ...current.services.prometheus[serviceMode].resources?.requests,
                            cpu: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          <label>
            <span>Requests.Memory</span>
            <input
              value={state.services.prometheus[serviceMode].resources?.requests?.memory ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    prometheus: {
                      ...current.services.prometheus,
                      [serviceMode]: {
                        ...current.services.prometheus[serviceMode],
                        resources: {
                          ...current.services.prometheus[serviceMode].resources,
                          requests: {
                            ...current.services.prometheus[serviceMode].resources?.requests,
                            memory: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          <label>
            <span>Limits.CPU</span>
            <input
              value={state.services.prometheus[serviceMode].resources?.limits?.cpu ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    prometheus: {
                      ...current.services.prometheus,
                      [serviceMode]: {
                        ...current.services.prometheus[serviceMode],
                        resources: {
                          ...current.services.prometheus[serviceMode].resources,
                          limits: {
                            ...current.services.prometheus[serviceMode].resources?.limits,
                            cpu: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          <label>
            <span>Limits.Memory</span>
            <input
              value={state.services.prometheus[serviceMode].resources?.limits?.memory ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    prometheus: {
                      ...current.services.prometheus,
                      [serviceMode]: {
                        ...current.services.prometheus[serviceMode],
                        resources: {
                          ...current.services.prometheus[serviceMode].resources,
                          limits: {
                            ...current.services.prometheus[serviceMode].resources?.limits,
                            memory: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          {isLocal ? (
            <label>
              <span>数据路径</span>
              <input
                value={state.services.prometheus.local.data_path}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      prometheus: {
                        ...current.services.prometheus,
                        local: {
                          ...current.services.prometheus.local,
                          data_path: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          ) : (
            <label>
              <span>storageClassName</span>
              <input
                value={state.services.prometheus.managed.storageClassName}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      prometheus: {
                        ...current.services.prometheus,
                        managed: {
                          ...current.services.prometheus.managed,
                          storageClassName: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          )}
        </div>
      </div>

      <div className="legacy-service-block">
        <h4>Grafana</h4>
        <div className="legacy-form-grid">
          {isLocal ? (
            <label>
              <span>部署节点</span>
              <span className="legacy-checkbox-panel">
                {state.nodes.map((node) => {
                  const checked = state.services.grafana.local.hosts.includes(node.name)

                  return (
                    <label key={node.name} className="legacy-checkbox-panel__item">
                      <input
                        type="checkbox"
                        aria-label={node.name}
                        checked={checked}
                        onChange={(event) =>
                          updateState(setState, (current) => {
                            const nextHosts = event.target.checked
                              ? [...current.services.grafana.local.hosts, node.name]
                              : current.services.grafana.local.hosts.filter((item) => item !== node.name)

                            return {
                              ...current,
                              services: {
                                ...current.services,
                                grafana: {
                                  ...current.services.grafana,
                                  local: {
                                    ...current.services.grafana.local,
                                    hosts: Array.from(new Set(nextHosts)),
                                  },
                                },
                              },
                            }
                          })
                        }
                      />
                      {node.name}
                    </label>
                  )
                })}
              </span>
            </label>
          ) : (
            <label>
              <span>副本数</span>
              <input
                type="number"
                min={1}
                value={state.services.grafana.managed.replica_count}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      grafana: {
                        ...current.services.grafana,
                        managed: {
                          ...current.services.grafana.managed,
                          replica_count: Number(event.target.value),
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          )}
          <label>
            <span>Requests.CPU</span>
            <input
              value={state.services.grafana[serviceMode].resources?.requests?.cpu ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    grafana: {
                      ...current.services.grafana,
                      [serviceMode]: {
                        ...current.services.grafana[serviceMode],
                        resources: {
                          ...current.services.grafana[serviceMode].resources,
                          requests: {
                            ...current.services.grafana[serviceMode].resources?.requests,
                            cpu: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          <label>
            <span>Requests.Memory</span>
            <input
              value={state.services.grafana[serviceMode].resources?.requests?.memory ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    grafana: {
                      ...current.services.grafana,
                      [serviceMode]: {
                        ...current.services.grafana[serviceMode],
                        resources: {
                          ...current.services.grafana[serviceMode].resources,
                          requests: {
                            ...current.services.grafana[serviceMode].resources?.requests,
                            memory: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          <label>
            <span>Limits.CPU</span>
            <input
              value={state.services.grafana[serviceMode].resources?.limits?.cpu ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    grafana: {
                      ...current.services.grafana,
                      [serviceMode]: {
                        ...current.services.grafana[serviceMode],
                        resources: {
                          ...current.services.grafana[serviceMode].resources,
                          limits: {
                            ...current.services.grafana[serviceMode].resources?.limits,
                            cpu: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          <label>
            <span>Limits.Memory</span>
            <input
              value={state.services.grafana[serviceMode].resources?.limits?.memory ?? ''}
              onChange={(event) =>
                updateState(setState, (current) => ({
                  ...current,
                  services: {
                    ...current.services,
                    grafana: {
                      ...current.services.grafana,
                      [serviceMode]: {
                        ...current.services.grafana[serviceMode],
                        resources: {
                          ...current.services.grafana[serviceMode].resources,
                          limits: {
                            ...current.services.grafana[serviceMode].resources?.limits,
                            memory: event.target.value,
                          },
                        },
                      },
                    },
                  },
                }))
              }
            />
          </label>
          {isLocal ? (
            <label>
              <span>数据路径</span>
              <input
                value={state.services.grafana.local.data_path}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      grafana: {
                        ...current.services.grafana,
                        local: {
                          ...current.services.grafana.local,
                          data_path: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          ) : (
            <label>
              <span>storageClassName</span>
              <input
                value={state.services.grafana.managed.storageClassName}
                onChange={(event) =>
                  updateState(setState, (current) => ({
                    ...current,
                    services: {
                      ...current.services,
                      grafana: {
                        ...current.services.grafana,
                        managed: {
                          ...current.services.grafana.managed,
                          storageClassName: event.target.value,
                        },
                      },
                    },
                  }))
                }
              />
            </label>
          )}
        </div>
      </div>
    </section>
  )
}

function ConnectStep({
  state,
  setState,
  previewFormat,
  setPreviewFormat,
}: {
  state: WizardState
  setState: Dispatch<SetStateAction<WizardState>>
  previewFormat: 'json' | 'yaml'
  setPreviewFormat: Dispatch<SetStateAction<'json' | 'yaml'>>
}) {
  const rdsExternal = !isInternalSource(state.resource_connect_info.rds.source_type)
  const redisExternal = !isInternalSource(state.resource_connect_info.redis.source_type)
  const searchExternal = !isInternalSource(state.resource_connect_info.opensearch.source_type)
  const mqExternal = !isInternalSource(state.resource_connect_info.mq.source_type)

  function updateRdsField<K extends keyof WizardState['resource_connect_info']['rds']>(
    field: K,
    value: WizardState['resource_connect_info']['rds'][K],
  ) {
    updateState(setState, (current) => ({
      ...current,
      resource_connect_info: {
        ...current.resource_connect_info,
        rds: {
          ...current.resource_connect_info.rds,
          [field]: value,
        },
      },
    }))
  }

  function updateRedisField<K extends keyof WizardState['resource_connect_info']['redis']>(
    field: K,
    value: WizardState['resource_connect_info']['redis'][K],
  ) {
    updateState(setState, (current) => ({
      ...current,
      resource_connect_info: {
        ...current.resource_connect_info,
        redis: {
          ...current.resource_connect_info.redis,
          [field]: value,
        },
      },
    }))
  }

  function updateOpenSearchField<K extends keyof WizardState['resource_connect_info']['opensearch']>(
    field: K,
    value: WizardState['resource_connect_info']['opensearch'][K],
  ) {
    updateState(setState, (current) => ({
      ...current,
      resource_connect_info: {
        ...current.resource_connect_info,
        opensearch: {
          ...current.resource_connect_info.opensearch,
          [field]: value,
        },
      },
    }))
  }

  function updateMqField<K extends keyof WizardState['resource_connect_info']['mq']>(
    field: K,
    value: WizardState['resource_connect_info']['mq'][K],
  ) {
    updateState(setState, (current) => ({
      ...current,
      resource_connect_info: {
        ...current.resource_connect_info,
        mq: {
          ...current.resource_connect_info.mq,
          [field]: value,
        },
      },
    }))
  }

  function updateMqAuth<K extends keyof WizardState['resource_connect_info']['mq']['auth']>(
    field: K,
    value: WizardState['resource_connect_info']['mq']['auth'][K],
  ) {
    updateState(setState, (current) => ({
      ...current,
      resource_connect_info: {
        ...current.resource_connect_info,
        mq: {
          ...current.resource_connect_info.mq,
          auth: {
            ...current.resource_connect_info.mq.auth,
            [field]: value,
          },
        },
      },
    }))
  }

  return (
    <section className="legacy-stack">
      <section className="legacy-card">
        <h3>连接配置表单</h3>

        <div className="legacy-service-block">
          <h4>MariaDB连接信息</h4>
          <div className="legacy-form-grid">
            <label>
              <span>资源类型</span>
              <select
                aria-label="RDS source"
                value={state.resource_connect_info.rds.source_type}
                onChange={(event) =>
                  updateState(setState, (current) => setResourceSource(current, 'rds', event.target.value as 'internal' | 'external'))
                }
              >
                <option value="internal">internal</option>
                <option value="external">external</option>
              </select>
            </label>
          </div>
          {rdsExternal ? (
            <>
              <div className="legacy-service-block">
                <h5>RDS 类型</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>类型选择</span>
                    <select
                      aria-label="RDS type"
                      value={state.resource_connect_info.rds.rds_type}
                      onChange={(event) => updateRdsField('rds_type', event.target.value)}
                    >
                      <option value="">请选择</option>
                      <option value="MySQL">MySQL</option>
                      <option value="PostgreSQL">PostgreSQL</option>
                      <option value="KingbaseES">KingbaseES</option>
                      <option value="Oracle">Oracle</option>
                    </select>
                  </label>
                </div>
              </div>
              <div className="legacy-service-block">
                <h5>账户信息</h5>
                <div className="legacy-form-grid">
                  <label className="legacy-inline-checkbox">
                    <input
                      aria-label="自动化创建数据库"
                      type="checkbox"
                      checked={state.resource_connect_info.rds.auto_create_database}
                      onChange={(event) => updateRdsField('auto_create_database', event.target.checked)}
                    />
                    自动化创建数据库
                  </label>
                </div>
                {state.resource_connect_info.rds.auto_create_database ? (
                  <div className="legacy-form-grid">
                    <label>
                      <span>用户名（管理权限）</span>
                      <input
                        aria-label="RDS admin username"
                        value={state.resource_connect_info.rds.admin_user}
                        onChange={(event) => updateRdsField('admin_user', event.target.value)}
                      />
                    </label>
                    <label>
                      <span>密码</span>
                      <input
                        aria-label="RDS admin password"
                        type="password"
                        value={state.resource_connect_info.rds.admin_passwd}
                        onChange={(event) => updateRdsField('admin_passwd', event.target.value)}
                      />
                    </label>
                  </div>
                ) : null}
                <div className="legacy-form-grid">
                  <label>
                    <span>用户名</span>
                    <input
                      aria-label="RDS username"
                      value={state.resource_connect_info.rds.username}
                      onChange={(event) => updateRdsField('username', event.target.value)}
                    />
                  </label>
                  <label>
                    <span>密码</span>
                    <input
                      aria-label="RDS password"
                      type="password"
                      value={state.resource_connect_info.rds.password}
                      onChange={(event) => updateRdsField('password', event.target.value)}
                    />
                  </label>
                </div>
              </div>
              <div className="legacy-service-block">
                <h5>连接信息</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>地址</span>
                    <input
                      aria-label="RDS hosts"
                      value={state.resource_connect_info.rds.hosts}
                      onChange={(event) => updateRdsField('hosts', event.target.value)}
                    />
                  </label>
                  <label>
                    <span>端口</span>
                    <input
                      aria-label="RDS port"
                      type="number"
                      value={state.resource_connect_info.rds.port ?? ''}
                      onChange={(event) =>
                        updateRdsField('port', event.target.value ? Number(event.target.value) : null)
                      }
                    />
                  </label>
                </div>
              </div>
            </>
          ) : (
            <div className="legacy-inline-note">
              <strong>使用内置 MariaDB</strong>
              <p>连接信息由初始化流程自动生成，这里无需额外填写。</p>
            </div>
          )}
        </div>

        <div className="legacy-service-block">
          <h4>Redis连接信息</h4>
          <div className="legacy-form-grid">
            <label>
              <span>资源类型</span>
              <select
                aria-label="Redis source"
                value={state.resource_connect_info.redis.source_type}
                onChange={(event) =>
                  updateState(setState, (current) =>
                    setResourceSource(current, 'redis', event.target.value as 'internal' | 'external'),
                  )
                }
              >
                <option value="internal">internal</option>
                <option value="external">external</option>
              </select>
            </label>
          </div>
          {redisExternal ? (
            <>
              <div className="legacy-service-block">
                <h5>Redis 连接模式</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>模式</span>
                    <select
                      aria-label="Redis connect type"
                      value={state.resource_connect_info.redis.connect_type}
                      onChange={(event) => updateRedisField('connect_type', event.target.value as WizardState['resource_connect_info']['redis']['connect_type'])}
                    >
                      <option value="">请选择</option>
                      <option value="standalone">standalone</option>
                      <option value="cluster">cluster</option>
                      <option value="master-slave">master-slave</option>
                      <option value="sentinel">sentinel</option>
                    </select>
                  </label>
                </div>
              </div>
              <div className="legacy-service-block">
                <h5>账户配置</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>用户名</span>
                    <input
                      aria-label="Redis username"
                      value={state.resource_connect_info.redis.username}
                      onChange={(event) => updateRedisField('username', event.target.value)}
                    />
                  </label>
                  <label>
                    <span>密码</span>
                    <input
                      aria-label="Redis password"
                      type="password"
                      value={state.resource_connect_info.redis.password}
                      onChange={(event) => updateRedisField('password', event.target.value)}
                    />
                  </label>
                </div>
              </div>
              {state.resource_connect_info.redis.connect_type === 'standalone' ||
              state.resource_connect_info.redis.connect_type === 'cluster' ? (
                <div className="legacy-service-block">
                  <h5>连接信息</h5>
                  <div className="legacy-form-grid">
                    <label>
                      <span>地址</span>
                      <input
                        aria-label="Redis hosts"
                        value={state.resource_connect_info.redis.hosts}
                        onChange={(event) => updateRedisField('hosts', event.target.value)}
                      />
                    </label>
                    <label>
                      <span>端口</span>
                      <input
                        aria-label="Redis port"
                        type="number"
                        value={state.resource_connect_info.redis.port ?? ''}
                        onChange={(event) =>
                          updateRedisField('port', event.target.value ? Number(event.target.value) : null)
                        }
                      />
                    </label>
                  </div>
                </div>
              ) : null}
              {state.resource_connect_info.redis.connect_type === 'master-slave' ? (
                <>
                  <div className="legacy-service-block">
                    <h5>master连接信息</h5>
                    <div className="legacy-form-grid">
                      <label>
                        <span>地址</span>
                        <input
                          aria-label="Master hosts"
                          value={state.resource_connect_info.redis.master_hosts}
                          onChange={(event) => updateRedisField('master_hosts', event.target.value)}
                        />
                      </label>
                      <label>
                        <span>端口</span>
                        <input
                          aria-label="Master port"
                          type="number"
                          value={state.resource_connect_info.redis.master_port ?? ''}
                          onChange={(event) =>
                            updateRedisField('master_port', event.target.value ? Number(event.target.value) : null)
                          }
                        />
                      </label>
                    </div>
                  </div>
                  <div className="legacy-service-block">
                    <h5>slave连接信息</h5>
                    <div className="legacy-form-grid">
                      <label>
                        <span>地址</span>
                        <input
                          aria-label="Slave hosts"
                          value={state.resource_connect_info.redis.slave_hosts}
                          onChange={(event) => updateRedisField('slave_hosts', event.target.value)}
                        />
                      </label>
                      <label>
                        <span>端口</span>
                        <input
                          aria-label="Slave port"
                          type="number"
                          value={state.resource_connect_info.redis.slave_port ?? ''}
                          onChange={(event) =>
                            updateRedisField('slave_port', event.target.value ? Number(event.target.value) : null)
                          }
                        />
                      </label>
                    </div>
                  </div>
                </>
              ) : null}
              {state.resource_connect_info.redis.connect_type === 'sentinel' ? (
                <div className="legacy-service-block">
                  <h5>哨兵连接信息</h5>
                  <div className="legacy-form-grid">
                    <label>
                      <span>Sentinel hosts</span>
                      <input
                        aria-label="Sentinel hosts"
                        value={state.resource_connect_info.redis.sentinel_hosts}
                        onChange={(event) => updateRedisField('sentinel_hosts', event.target.value)}
                      />
                    </label>
                    <label>
                      <span>Sentinel port</span>
                      <input
                        aria-label="Sentinel port"
                        type="number"
                        value={state.resource_connect_info.redis.sentinel_port ?? ''}
                        onChange={(event) =>
                          updateRedisField('sentinel_port', event.target.value ? Number(event.target.value) : null)
                        }
                      />
                    </label>
                    <label>
                      <span>Master group name</span>
                      <input
                        aria-label="Master group name"
                        value={state.resource_connect_info.redis.master_group_name}
                        onChange={(event) => updateRedisField('master_group_name', event.target.value)}
                      />
                    </label>
                  </div>
                </div>
              ) : null}
            </>
          ) : (
            <div className="legacy-inline-note">
              <strong>使用内置 Redis</strong>
              <p>连接配置由初始化过程自动下发，这里无需单独填写。</p>
            </div>
          )}
        </div>

        <div className="legacy-service-block">
          <h4>SearchEngine连接信息</h4>
          <div className="legacy-form-grid">
            <label>
              <span>资源类型</span>
              <select
                aria-label="OpenSearch source"
                value={state.resource_connect_info.opensearch.source_type}
                onChange={(event) =>
                  updateState(setState, (current) =>
                    setResourceSource(current, 'opensearch', event.target.value as 'internal' | 'external'),
                  )
                }
              >
                <option value="internal">internal</option>
                <option value="external">external</option>
              </select>
            </label>
          </div>
          {searchExternal ? (
            <>
              <div className="legacy-service-block">
                <h5>账户信息</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>用户名</span>
                    <input
                      aria-label="OpenSearch username"
                      value={state.resource_connect_info.opensearch.username}
                      onChange={(event) => updateOpenSearchField('username', event.target.value)}
                    />
                  </label>
                  <label>
                    <span>密码</span>
                    <input
                      aria-label="OpenSearch password"
                      type="password"
                      value={state.resource_connect_info.opensearch.password}
                      onChange={(event) => updateOpenSearchField('password', event.target.value)}
                    />
                  </label>
                </div>
              </div>
              <div className="legacy-service-block">
                <h5>连接信息</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>地址</span>
                    <input
                      aria-label="OpenSearch hosts"
                      value={state.resource_connect_info.opensearch.hosts}
                      onChange={(event) => updateOpenSearchField('hosts', event.target.value)}
                    />
                  </label>
                  <label>
                    <span>端口</span>
                    <input
                      aria-label="OpenSearch port"
                      type="number"
                      value={state.resource_connect_info.opensearch.port ?? ''}
                      onChange={(event) =>
                        updateOpenSearchField('port', event.target.value ? Number(event.target.value) : null)
                      }
                    />
                  </label>
                </div>
              </div>
            </>
          ) : (
            <div className="legacy-inline-note">
              <strong>使用内置 SearchEngine</strong>
              <p>SearchEngine 连接信息会在初始化时自动配置。</p>
            </div>
          )}
          <div className="legacy-service-block">
            <h5>SearchEngine版本</h5>
            <div className="legacy-form-grid">
              <label>
                <span>发行版</span>
                <select
                  aria-label="OpenSearch distribution"
                  value={state.resource_connect_info.opensearch.distribution}
                  disabled={!searchExternal}
                  onChange={(event) => updateOpenSearchField('distribution', event.target.value)}
                >
                  <option value="elasticsearch">elasticsearch</option>
                  <option value="opensearch">opensearch</option>
                </select>
              </label>
              <label>
                <span>版本</span>
                <select
                  aria-label="OpenSearch version"
                  value={state.resource_connect_info.opensearch.version}
                  disabled={!searchExternal}
                  onChange={(event) => updateOpenSearchField('version', event.target.value)}
                >
                  <option value="5.6.4">5.6.4</option>
                  <option value="7.10.0">7.10.0</option>
                </select>
              </label>
            </div>
          </div>
        </div>

        <div className="legacy-service-block">
          <h4>MQ连接信息</h4>
          <div className="legacy-form-grid">
            <label>
              <span>资源类型</span>
              <select
                aria-label="MQ source"
                value={state.resource_connect_info.mq.source_type}
                onChange={(event) =>
                  updateState(setState, (current) => setResourceSource(current, 'mq', event.target.value as 'internal' | 'external'))
                }
              >
                <option value="internal">internal</option>
                <option value="external">external</option>
              </select>
            </label>
            <label>
              <span>MQ 类型</span>
              <select
                aria-label="MQ type"
                value={state.resource_connect_info.mq.mq_type}
                onChange={(event) =>
                  updateMqField('mq_type', event.target.value as WizardState['resource_connect_info']['mq']['mq_type'])
                }
              >
                {mqExternal ? <option value="">请选择</option> : null}
                <option value="kafka">kafka</option>
                {mqExternal ? <option value="nsq">nsq</option> : null}
                {mqExternal ? <option value="tonglink">tonglink</option> : null}
                {mqExternal ? <option value="htp20">htp20</option> : null}
                {mqExternal ? <option value="htp202">htp202</option> : null}
                {mqExternal ? <option value="bmq">bmq</option> : null}
              </select>
            </label>
          </div>
          {mqExternal ? (
            <>
              <div className="legacy-service-block">
                <h5>连接信息</h5>
                <div className="legacy-form-grid">
                  <label>
                    <span>地址</span>
                    <input
                      aria-label="MQ hosts"
                      value={state.resource_connect_info.mq.mq_hosts}
                      onChange={(event) => updateMqField('mq_hosts', event.target.value)}
                    />
                  </label>
                  <label>
                    <span>端口</span>
                    <input
                      aria-label="MQ port"
                      type="number"
                      value={state.resource_connect_info.mq.mq_port ?? ''}
                      onChange={(event) =>
                        updateMqField('mq_port', event.target.value ? Number(event.target.value) : null)
                      }
                    />
                  </label>
                </div>
              </div>
              {state.resource_connect_info.mq.mq_type === 'nsq' ? (
                <div className="legacy-service-block">
                  <h5>lookupd 连接信息</h5>
                  <div className="legacy-form-grid">
                    <label>
                      <span>NSQ lookupd hosts</span>
                      <input
                        aria-label="NSQ lookupd hosts"
                        value={state.resource_connect_info.mq.mq_lookupd_hosts}
                        onChange={(event) => updateMqField('mq_lookupd_hosts', event.target.value)}
                      />
                    </label>
                    <label>
                      <span>NSQ lookupd port</span>
                      <input
                        aria-label="NSQ lookupd port"
                        type="number"
                        value={state.resource_connect_info.mq.mq_lookupd_port ?? ''}
                        onChange={(event) =>
                          updateMqField('mq_lookupd_port', event.target.value ? Number(event.target.value) : null)
                        }
                      />
                    </label>
                  </div>
                </div>
              ) : null}
              {state.resource_connect_info.mq.mq_type === 'kafka' ? (
                <div className="legacy-service-block">
                  <h5>认证账户信息</h5>
                  <div className="legacy-form-grid">
                    <label>
                      <span>用户名</span>
                      <input
                        aria-label="MQ username"
                        value={state.resource_connect_info.mq.auth.username}
                        onChange={(event) => updateMqAuth('username', event.target.value)}
                      />
                    </label>
                    <label>
                      <span>密码</span>
                      <input
                        aria-label="MQ password"
                        type="password"
                        value={state.resource_connect_info.mq.auth.password}
                        onChange={(event) => updateMqAuth('password', event.target.value)}
                      />
                    </label>
                    <label>
                      <span>认证机制</span>
                      <select
                        aria-label="MQ mechanism"
                        value={state.resource_connect_info.mq.auth.mechanism}
                        onChange={(event) => updateMqAuth('mechanism', event.target.value)}
                      >
                        <option value="PLAIN">PLAIN</option>
                        <option value="SCRAM-SHA-512">SCRAM-SHA-512</option>
                        <option value="SCRAM-SHA-256">SCRAM-SHA-256</option>
                      </select>
                    </label>
                  </div>
                </div>
              ) : null}
            </>
          ) : (
            <div className="legacy-inline-note">
              <strong>使用内置 Kafka / ZooKeeper</strong>
              <p>MQ 连接配置由初始化脚本自动生成，外置时再填写这里的连接参数。</p>
            </div>
          )}
        </div>
      </section>

      <PreviewPanel state={state} format={previewFormat} onFormatChange={setPreviewFormat} />
    </section>
  )
}

export function Wizard() {
  const [screen, setScreen] = useState<WizardScreen>('template')
  const [storageMode, setStorageMode] = useState<StorageMode | null>(null)
  const [state, setState] = useState(defaultWizardState)
  const [currentStepIndex, setCurrentStepIndex] = useState(0)
  const [previewFormat, setPreviewFormat] = useState<'json' | 'yaml'>('json')
  const [servicePasswordVisibility, setServicePasswordVisibility] = useState({
    mariadb: false,
    redis: false,
  })
  const mariadbPasswordRef = useRef<HTMLInputElement>(null)
  const redisPasswordRef = useRef<HTMLInputElement>(null)
  const errorCardRef = useRef<HTMLElement>(null)
  const validation = useMemo(() => validateWizardState(state), [state])
  const { status, message, submit, reset } = useSubmit()

  const steps = storageMode === 'deposit-kubernetes' ? managedSteps : localSteps
  const currentStep = steps[currentStepIndex]?.key
  const currentIssues = currentStep ? getCurrentStepIssues(currentStep, validation.issues) : []

  function chooseTemplate(mode: StorageMode) {
    setStorageMode(mode)
    setCurrentStepIndex(0)
    setScreen('steps')
    setServicePasswordVisibility({
      mariadb: false,
      redis: false,
    })
    setState((current) => ({
      ...current,
      deploymentKind: mode === 'standard' ? 'local' : 'managed',
      resource_connect_info: {
        ...current.resource_connect_info,
        rds: {
          ...current.resource_connect_info.rds,
          source_type: mode === 'standard' ? 'internal' : 'external',
        },
        redis: {
          ...current.resource_connect_info.redis,
          source_type: mode === 'standard' ? 'internal' : 'external',
        },
        opensearch: {
          ...current.resource_connect_info.opensearch,
          source_type: mode === 'standard' ? 'internal' : 'external',
        },
        mq: {
          ...current.resource_connect_info.mq,
          source_type: mode === 'standard' ? 'internal' : 'external',
        },
      },
    }))
  }

  function goNext() {
    if (currentIssues.length > 0) {
      focusIssueTarget()
      return
    }
    setCurrentStepIndex((index) => Math.min(index + 1, steps.length - 1))
  }

  function goPrev() {
    setCurrentStepIndex((index) => Math.max(index - 1, 0))
  }

  function toggleServicePasswordVisibility(field: 'mariadb' | 'redis') {
    setServicePasswordVisibility((current) => ({
      ...current,
      [field]: !current[field],
    }))
  }

  function focusIssueTarget() {
    const firstIssue = currentIssues[0]

    const target =
      firstIssue?.field === 'services.mariadb.local.admin_passwd' || firstIssue?.field === 'services.mariadb.managed.admin_passwd'
        ? mariadbPasswordRef.current
        : firstIssue?.field === 'services.redis.local.admin_passwd' || firstIssue?.field === 'services.redis.managed.admin_passwd'
          ? redisPasswordRef.current
          : errorCardRef.current

    if (!target) {
      return
    }

    if ('scrollIntoView' in target && typeof target.scrollIntoView === 'function') {
      target.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }

    if ('focus' in target) {
      target.focus()
    }
  }

  async function finish() {
    if (currentIssues.length > 0) {
      focusIssueTarget()
      return
    }
    if (status === 'error') {
      reset()
    }
    await submit(state)
  }

  if (status === 'submitting') {
    return (
      <div className="legacy-page">
        <header className="legacy-header">
          <h1>部署工具</h1>
        </header>
        <main className="legacy-main">
          <section className="legacy-card legacy-success-card" aria-live="polite">
            <h2>正在提交初始化请求</h2>
            <p className="legacy-muted">正在向 proton-cli server 提交配置，请稍候。</p>
          </section>
        </main>
      </div>
    )
  }

  if (status === 'success' && message) {
    return (
      <div className="legacy-page">
        <header className="legacy-header">
          <h1>部署工具</h1>
        </header>
        <main className="legacy-main">
          <section className="legacy-card legacy-success-card">
            <h2>集群初始化成功</h2>
            <p className="legacy-muted">{message}</p>
          </section>
        </main>
      </div>
    )
  }

  return (
    <div className="legacy-page">
      <header className="legacy-header">
        <h1>部署工具</h1>
      </header>

      <main className="legacy-main">
        {screen === 'template' ? (
          <TemplateChooser onChoose={chooseTemplate} />
        ) : (
          <>
            <section className="legacy-steps" aria-label="初始化步骤">
              {steps.map((step, index) => (
                <div
                  key={step.key}
                  className={`legacy-step ${index === currentStepIndex ? 'is-active' : ''}`}
                  aria-current={index === currentStepIndex ? 'step' : undefined}
                >
                  <strong>{step.title}</strong>
                  <span>{step.description}</span>
                </div>
              ))}
            </section>

            <section className="legacy-content">
              {currentStep === 'node' ? <NodeStep state={state} setState={setState} /> : null}
              {currentStep === 'network' ? <NetworkStep state={state} setState={setState} /> : null}
              {currentStep === 'repository' ? <RepositoryStep state={state} setState={setState} /> : null}
              {currentStep === 'service' ? (
                <ServiceStep
                  state={state}
                  setState={setState}
                  passwordVisibility={servicePasswordVisibility}
                  onTogglePasswordVisibility={toggleServicePasswordVisibility}
                  mariadbPasswordRef={mariadbPasswordRef}
                  redisPasswordRef={redisPasswordRef}
                />
              ) : null}
              {currentStep === 'connect' ? (
                <ConnectStep
                  state={state}
                  setState={setState}
                  previewFormat={previewFormat}
                  setPreviewFormat={setPreviewFormat}
                />
              ) : null}

              {currentIssues.length > 0 ? (
                <section ref={errorCardRef} className="legacy-card legacy-error-card">
                  <h4>当前步骤存在校验问题</h4>
                  <ul>
                    {currentIssues.map((issue) => (
                      <li key={`${issue.field}-${issue.message}`}>{issue.message}</li>
                    ))}
                  </ul>
                </section>
              ) : null}
            </section>
          </>
        )}
      </main>

      {screen === 'steps' ? (
        <footer className="legacy-footer">
          {currentStepIndex > 0 ? (
            <button type="button" className="legacy-button" onClick={goPrev}>
              上一步
            </button>
          ) : (
            <span />
          )}

          {currentStepIndex === steps.length - 1 ? (
            <button type="button" className="legacy-button is-primary" onClick={finish}>
              完成
            </button>
          ) : (
            <button type="button" className="legacy-button is-primary" onClick={goNext}>
              下一步
            </button>
          )}
        </footer>
      ) : null}

      {screen === 'steps' && status === 'error' && message ? (
        <div className="legacy-modal-backdrop" role="presentation">
          <section
            className="legacy-modal legacy-card legacy-error-card"
            role="dialog"
            aria-modal="true"
            aria-labelledby="submit-error-title"
            aria-live="assertive"
          >
            <h2 id="submit-error-title">初始化失败</h2>
            <p className="legacy-muted">{message}</p>
            <div className="legacy-modal-actions">
              <button type="button" className="legacy-button" onClick={reset}>
                确认
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </div>
  )
}
