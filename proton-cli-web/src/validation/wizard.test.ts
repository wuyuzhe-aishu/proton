import { describe, expect, it } from 'vitest'

import { defaultWizardState } from '../schema/defaults'
import { validateWizardState } from './wizard'

describe('validateWizardState', () => {
  it('accepts the default local state when resource connections are switched to internal', () => {
    const result = validateWizardState({
      ...defaultWizardState,
       nodes: [{ id: 'test-node-1', name: 'node1', ip4: '192.168.40.11', ip6: '' }],
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
      services: {
        ...defaultWizardState.services,
        mariadb: {
          ...defaultWizardState.services.mariadb,
          local: {
            ...defaultWizardState.services.mariadb.local,
            admin_passwd: 'mariadb-pass',
          },
        },
        redis: {
          ...defaultWizardState.services.redis,
          local: {
            ...defaultWizardState.services.redis.local,
            admin_passwd: 'redis-pass',
          },
        },
      },
    })

    expect(result.valid).toBe(true)
    expect(result.issues).toEqual([])
  })

  it('requires zookeeper when kafka is enabled', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      services: {
        ...defaultWizardState.services,
        zookeeper: {
          ...defaultWizardState.services.zookeeper,
          local: {
            ...defaultWizardState.services.zookeeper.local,
            enabled: false,
          },
        },
      },
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
          hosts: '127.0.0.1',
          port: 3306,
          username: 'user',
          password: 'pass',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
          mq_type: 'kafka',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.message.includes('Kafka requires ZooKeeper'))).toBe(true)
  })

  it('requires replica count for managed mariadb', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      deploymentKind: 'managed',
      services: {
        ...defaultWizardState.services,
        mariadb: {
          ...defaultWizardState.services.mariadb,
          managed: {
            ...defaultWizardState.services.mariadb.managed,
            replica_count: 0,
          },
        },
      },
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
          hosts: '127.0.0.1',
          port: 3306,
          username: 'user',
          password: 'pass',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
          mq_type: 'kafka',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'services.mariadb.managed.replica_count')).toBe(true)
  })

  it('requires an ipv6 address when dual stack includes IPv6', () => {
    const result = validateWizardState({
      ...defaultWizardState,
       nodes: [{ id: 'test-node-1', name: 'node1', ip4: '192.168.40.11', ip6: '' }],
      cs: {
        ...defaultWizardState.cs,
        local: {
          ...defaultWizardState.cs.local,
          ipFamilies: ['IPv4', 'IPv6'],
          enableDualStack: true,
        },
      },
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'nodes[0].ip6')).toBe(true)
  })

  it('requires external chartmuseum credentials for managed repository mode', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      deploymentKind: 'managed',
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          hosts: '127.0.0.1',
          port: 3306,
          username: 'user',
          password: 'pass',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          hosts: '127.0.0.1',
          port: 6379,
          source_type: 'external',
          connect_type: 'standalone',
          username: 'user',
          password: 'pass',
          sentinel_hosts: '',
          sentinel_port: null,
          master_group_name: '',
          master_hosts: '',
          master_port: null,
          slave_hosts: '',
          slave_port: null,
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'external',
          hosts: '127.0.0.1',
          port: 9200,
          username: 'user',
          password: 'pass',
          version: '7.10.0',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'external',
          mq_type: 'kafka',
          mq_hosts: '127.0.0.1',
          mq_port: 9092,
          auth: {
            username: 'user',
            password: 'pass',
            mechanism: 'PLAIN',
          },
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'cr.external.chartmuseum.host')).toBe(true)
  })

  it('rejects invalid ingress-nginx ports', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      cs: {
        ...defaultWizardState.cs,
        local: {
          ...defaultWizardState.cs.local,
          ingressNginx: {
            httpPort: 0,
            httpsPort: 70000,
          },
        },
      },
       nodes: [{ id: 'test-node-1', name: 'node1', ip4: '192.168.40.11', ip6: '' }],
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'cs.local.ingressNginx.httpPort')).toBe(true)
    expect(result.issues.some((issue) => issue.field === 'cs.local.ingressNginx.httpsPort')).toBe(true)
  })

  it('requires sentinel fields when redis external mode uses sentinel', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'external',
          connect_type: 'sentinel',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'resource_connect_info.redis.sentinel_hosts')).toBe(true)
    expect(result.issues.some((issue) => issue.field === 'resource_connect_info.redis.master_group_name')).toBe(true)
  })

  it('requires namespace and serviceaccount for managed kubernetes mode', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      deploymentKind: 'managed',
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
          hosts: '127.0.0.1',
          port: 3306,
          username: 'user',
          password: 'pass',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
          mq_type: 'kafka',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'cs.managed.namespace')).toBe(true)
    expect(result.issues.some((issue) => issue.field === 'cs.managed.serviceaccount')).toBe(true)
  })

  it('requires admin credentials when external rds enables auto database creation', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'external',
          rds_type: 'MySQL',
          hosts: '127.0.0.1',
          port: 3306,
          username: 'app_user',
          password: 'app_pass',
          auto_create_database: true,
          admin_user: '',
          admin_passwd: '',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'resource_connect_info.rds.admin_user')).toBe(true)
    expect(result.issues.some((issue) => issue.field === 'resource_connect_info.rds.admin_passwd')).toBe(true)
  })

  it('requires distribution when opensearch is external', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      resource_connect_info: {
        ...defaultWizardState.resource_connect_info,
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'external',
          hosts: '127.0.0.1',
          port: 9200,
          username: 'user',
          password: 'pass',
          distribution: '',
          version: '7.10.0',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'resource_connect_info.opensearch.distribution')).toBe(true)
  })

  it('does not require local deployment fields when base services are marked external', () => {
    const result = validateWizardState({
      ...defaultWizardState,
      services: {
        ...defaultWizardState.services,
        mariadb: {
          ...defaultWizardState.services.mariadb,
          local: {
            ...defaultWizardState.services.mariadb.local,
            hosts: [],
            data_path: '',
          },
        },
        redis: {
          ...defaultWizardState.services.redis,
          local: {
            ...defaultWizardState.services.redis.local,
            hosts: [],
            data_path: '',
          },
        },
        opensearch: {
          ...defaultWizardState.services.opensearch,
          local: {
            ...defaultWizardState.services.opensearch.local,
            hosts: [],
            data_path: '',
          },
        },
        kafka: {
          ...defaultWizardState.services.kafka,
          local: {
            ...defaultWizardState.services.kafka.local,
            hosts: [],
            data_path: '',
          },
        },
        zookeeper: {
          ...defaultWizardState.services.zookeeper,
          local: {
            ...defaultWizardState.services.zookeeper.local,
            hosts: [],
            data_path: '',
          },
        },
      },
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'external',
          rds_type: 'MySQL',
          hosts: '192.168.10.10',
          port: 3306,
          username: 'root',
          password: 'secret',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'external',
          connect_type: 'standalone',
          hosts: '192.168.10.11',
          port: 6379,
          username: 'default',
          password: 'secret',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'external',
          hosts: '192.168.10.12',
          port: 9200,
          username: 'admin',
          password: 'secret',
          version: '7.10.0',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'external',
          mq_type: 'kafka',
          mq_hosts: '192.168.10.13',
          mq_port: 9092,
          auth: {
            username: 'user',
            password: 'secret',
            mechanism: 'PLAIN',
          },
        },
      },
    })

    expect(result.issues.some((issue) => issue.field.startsWith('services.mariadb.local'))).toBe(false)
    expect(result.issues.some((issue) => issue.field.startsWith('services.redis.local'))).toBe(false)
    expect(result.issues.some((issue) => issue.field.startsWith('services.opensearch.local'))).toBe(false)
    expect(result.issues.some((issue) => issue.field.startsWith('services.kafka.local'))).toBe(false)
    expect(result.issues.some((issue) => issue.field.startsWith('services.zookeeper.local'))).toBe(false)
  })

  it('requires internal mariadb and redis passwords in the service step', () => {
    const result = validateWizardState({
      ...defaultWizardState,
       nodes: [{ id: 'test-node-1', name: 'node1', ip4: '192.168.40.11', ip6: '' }],
      resource_connect_info: {
        rds: {
          ...defaultWizardState.resource_connect_info.rds,
          source_type: 'internal',
        },
        redis: {
          ...defaultWizardState.resource_connect_info.redis,
          source_type: 'internal',
        },
        opensearch: {
          ...defaultWizardState.resource_connect_info.opensearch,
          source_type: 'internal',
        },
        mq: {
          ...defaultWizardState.resource_connect_info.mq,
          source_type: 'internal',
        },
      },
      services: {
        ...defaultWizardState.services,
        mariadb: {
          ...defaultWizardState.services.mariadb,
          local: {
            ...defaultWizardState.services.mariadb.local,
            admin_passwd: '',
          },
        },
        redis: {
          ...defaultWizardState.services.redis,
          local: {
            ...defaultWizardState.services.redis.local,
            admin_passwd: '',
          },
        },
      },
    })

    expect(result.valid).toBe(false)
    expect(result.issues.some((issue) => issue.field === 'services.mariadb.local.admin_passwd')).toBe(true)
    expect(result.issues.some((issue) => issue.field === 'services.redis.local.admin_passwd')).toBe(true)
  })
})
