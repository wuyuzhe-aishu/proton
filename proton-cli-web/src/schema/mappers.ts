import type {
  KafkaLocalValue,
  KafkaManagedValue,
  OpenSearchLocalValue,
  OpenSearchManagedValue,
  PrometheusLocalValue,
  PrometheusManagedValue,
  GrafanaLocalValue,
  GrafanaManagedValue,
  SubmitConfig,
  WizardState,
} from './config'

function cleanNode(node: WizardState['nodes'][number]) {
  return {
    name: node.name,
    ...(node.ip4 ? { ip4: node.ip4 } : {}),
    ...(node.ip6 ? { ip6: node.ip6 } : {}),
  }
}

function isBlankString(value: unknown): value is string {
  return typeof value === 'string' && value.trim() === ''
}

function cleanStringMap<T extends Record<string, unknown>>(value: T | undefined) {
  if (!value) {
    return undefined
  }

  const entries = Object.entries(value).filter(([, item]) => !isBlankString(item))
  if (!entries.length) {
    return undefined
  }

  return Object.fromEntries(entries)
}

function cleanExternalServiceList<T extends { name?: string; ip?: string; port?: number | null }>(value: T[] | undefined) {
  if (!value?.length) {
    return undefined
  }

  const items = value.filter((item) => {
    const hasName = typeof item.name === 'string' && item.name.trim() !== ''
    const hasIP = typeof item.ip === 'string' && item.ip.trim() !== ''
    const hasPort = typeof item.port === 'number' && item.port > 0
    return hasName && hasIP && hasPort
  })

  return items.length ? items : undefined
}

function cleanKafkaCommon(value: Pick<KafkaLocalValue, 'env' | 'disable_external_service' | 'external_service_list'>) {
  const env = cleanStringMap(value.env)
  const externalServiceList = cleanExternalServiceList(
    value.external_service_list,
  )

  return {
    ...(env ? { env } : {}),
    ...(externalServiceList ? { external_service_list: externalServiceList } : {}),
    ...(value.disable_external_service ? { disable_external_service: true } : {}),
  }
}

function cleanKafkaLocalConfig(value: KafkaLocalValue) {
  return {
    enabled: value.enabled,
    hosts: value.hosts,
    data_path: value.data_path,
    ...(value.storage_capacity ? { storage_capacity: value.storage_capacity } : {}),
    ...cleanKafkaCommon(value),
  }
}

function cleanPrometheusLocalConfig(value: PrometheusLocalValue) {
  return {
    hosts: value.hosts,
    data_path: value.data_path,
    ...(value.storage_capacity ? { storage_capacity: value.storage_capacity } : {}),
    ...(value.storageClassName ? { storageClassName: value.storageClassName } : {}),
    ...(value.resources ? { resources: value.resources } : {}),
  }
}

function cleanPrometheusManagedConfig(value: PrometheusManagedValue) {
  return {
    replica_count: value.replica_count,
    ...(value.storage_capacity ? { storage_capacity: value.storage_capacity } : {}),
    ...(value.storageClassName ? { storageClassName: value.storageClassName } : {}),
    ...(value.resources ? { resources: value.resources } : {}),
  }
}

function cleanGrafanaLocalConfig(value: GrafanaLocalValue) {
  return {
    hosts: value.hosts,
    data_path: value.data_path,
    ...(value.storage_capacity ? { storage_capacity: value.storage_capacity } : {}),
    ...(value.storageClassName ? { storageClassName: value.storageClassName } : {}),
    ...(value.resources ? { resources: value.resources } : {}),
  }
}

function cleanGrafanaManagedConfig(value: GrafanaManagedValue) {
  return {
    replica_count: value.replica_count,
    ...(value.storage_capacity ? { storage_capacity: value.storage_capacity } : {}),
    ...(value.storageClassName ? { storageClassName: value.storageClassName } : {}),
    ...(value.resources ? { resources: value.resources } : {}),
  }
}

function cleanKafkaManagedConfig(value: KafkaManagedValue) {
  return {
    enabled: value.enabled,
    replica_count: value.replica_count,
    ...(value.storage_capacity ? { storage_capacity: value.storage_capacity } : {}),
    ...(value.storageClassName ? { storageClassName: value.storageClassName } : {}),
    ...cleanKafkaCommon(value),
  }
}

function cleanOpenSearchCommon(value: Pick<OpenSearchLocalValue, 'config' | 'mode' | 'settings'>) {
  const config = cleanStringMap(value.config)

  return {
    mode: value.mode,
    settings: value.settings,
    ...(config ? { config } : {}),
  }
}

function cleanOpenSearchLocalConfig(value: OpenSearchLocalValue) {
  return {
    enabled: value.enabled,
    hosts: value.hosts,
    data_path: value.data_path,
    ...cleanOpenSearchCommon(value),
  }
}

function cleanOpenSearchManagedConfig(value: OpenSearchManagedValue) {
  return {
    enabled: value.enabled,
    replica_count: value.replica_count,
    ...(value.storageClassName ? { storageClassName: value.storageClassName } : {}),
    ...cleanOpenSearchCommon(value),
  }
}

function cleanCSConfig<T extends { addons: string[]; ingressNginx: { httpPort: number; httpsPort: number } }>(value: T) {
  const csConfig: Record<string, unknown> = {
    ...value,
  }

  if (value.addons.includes('ingress-nginx')) {
    csConfig.addonsConfig = {
      'ingress-nginx': {
        httpPort: value.ingressNginx.httpPort,
        httpsPort: value.ingressNginx.httpsPort,
      },
    }
  }

  delete csConfig.ingressNginx
  return csConfig
}

export function toSubmitConfig(state: WizardState): SubmitConfig {
  const isLocal = state.deploymentKind === 'local'
  const useInternalRds = state.resource_connect_info.rds.source_type === 'internal'
  const useInternalRedis = state.resource_connect_info.redis.source_type === 'internal'
  const useInternalSearch = state.resource_connect_info.opensearch.source_type === 'internal'
  const useInternalMq = state.resource_connect_info.mq.source_type === 'internal'
  const internalMariaDB = useInternalRds
    ? isLocal
      ? state.services.mariadb.local
      : state.services.mariadb.managed
    : null
  const rdsConfig = useInternalRds
    ? {
        ...state.resource_connect_info.rds,
        username: state.resource_connect_info.rds.username || internalMariaDB?.admin_user || 'root',
        password: state.resource_connect_info.rds.password || internalMariaDB?.admin_passwd || '',
      }
    : state.resource_connect_info.rds

  const config: SubmitConfig = {
    apiVersion: 'v1',
    nodes: state.nodes.map(cleanNode),
    chrony: {
      mode: state.chrony.mode,
      ...(state.chrony.server.length ? { server: state.chrony.server } : {}),
    },
    firewall: {
      mode: state.firewall.mode,
    },
    cs: isLocal
      ? {
          provisioner: 'local',
          ...cleanCSConfig(state.cs.local),
        }
      : {
          provisioner: 'external',
          ...cleanCSConfig(state.cs.managed),
        },
    cr: isLocal
      ? {
          local: state.cr.local,
        }
      : {
          external: state.cr.external,
        },
    component_management: {},
    resource_connect_info: {
      ...state.resource_connect_info,
      rds: rdsConfig,
    },
  }

  if (useInternalRds && isLocal && state.services.mariadb.local.enabled) {
    config.proton_mariadb = state.services.mariadb.local
  } else if (useInternalRds && !isLocal && state.services.mariadb.managed.enabled) {
    config.proton_mariadb = state.services.mariadb.managed
  }

  if (useInternalRedis && isLocal && state.services.redis.local.enabled) {
    config.proton_redis = state.services.redis.local
  } else if (useInternalRedis && !isLocal && state.services.redis.managed.enabled) {
    config.proton_redis = state.services.redis.managed
  }

  if (useInternalSearch && isLocal && state.services.opensearch.local.enabled) {
    config.opensearch = cleanOpenSearchLocalConfig(state.services.opensearch.local)
  } else if (useInternalSearch && !isLocal && state.services.opensearch.managed.enabled) {
    config.opensearch = cleanOpenSearchManagedConfig(state.services.opensearch.managed)
  }

  if (useInternalMq && isLocal) {
    config.kafka = {
      ...cleanKafkaLocalConfig(state.services.kafka.local),
      enabled: true,
    }
  } else if (useInternalMq && !isLocal) {
    config.kafka = {
      ...cleanKafkaManagedConfig(state.services.kafka.managed),
      enabled: true,
    }
  }

  if (useInternalMq && isLocal) {
    config.zookeeper = {
      ...state.services.zookeeper.local,
      enabled: true,
    }
  } else if (useInternalMq && !isLocal) {
    config.zookeeper = {
      ...state.services.zookeeper.managed,
      enabled: true,
    }
  }

  if (isLocal) {
    config.prometheus = cleanPrometheusLocalConfig(state.services.prometheus.local)
  } else {
    config.prometheus = cleanPrometheusManagedConfig(state.services.prometheus.managed)
  }

  if (isLocal) {
    config.grafana = cleanGrafanaLocalConfig(state.services.grafana.local)
  } else {
    config.grafana = cleanGrafanaManagedConfig(state.services.grafana.managed)
  }

  return config
}
