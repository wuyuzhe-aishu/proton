export type DeploymentKind = 'local' | 'managed'

export type SourceType = 'internal' | 'external'
export type RedisConnectType =
  | 'sentinel'
  | 'master-slave'
  | 'standalone'
  | 'cluster'
  | ''
export type MQType = 'kafka' | 'nsq' | 'tonglink' | 'htp20' | 'htp202' | 'bmq' | ''

export interface NodeFormValue {
  id: string
  name: string
  ip4: string
  ip6: string
}

export interface LocalCsFormValue {
  master: string[]
  addons: string[]
  ingressNginx: {
    httpPort: number
    httpsPort: number
  }
  ipFamilies: string[]
  enableDualStack: boolean
  ha_port: number
  host_network: {
    bip: string
    pod_network_cidr: string
    service_cidr: string
    ipv4_interface: string
    ipv6_interface: string
  }
  etcd_data_dir: string
  containerd_data_dir: string
}

export interface ManagedCsFormValue {
  namespace: string
  serviceaccount: string
  addons: string[]
  ingressNginx: {
    httpPort: number
    httpsPort: number
  }
}

export interface LocalCrFormValue {
  hosts: string[]
  ports: {
    chartmuseum: number
    registry: number
    rpm: number
    cr_manager: number
  }
  ha_ports: {
    chartmuseum: number
    registry: number
    rpm: number
    cr_manager: number
  }
  storage: string
}

export interface ExternalCrRepository {
  host: string
  username: string
  password: string
}

export interface ExternalCrFormValue {
  chart_repository: 'chartmuseum' | 'oci'
  image_repository: 'registry' | 'oci'
  registry: ExternalCrRepository
  chartmuseum: ExternalCrRepository
  oci: {
    registry: string
    username: string
    password: string
    plain_http: boolean
  }
}

export interface LocalServiceValue {
  enabled: boolean
  hosts: string[]
  data_path: string
  storage_capacity: string
}

export interface ManagedServiceValue {
  enabled: boolean
  replica_count: number
  storage_capacity: string
  storageClassName: string
}

export interface ExternalServiceConfig {
  name: string
  ip: string
  port: number | null
  enableSSL: boolean
}

export interface MariaDBLocalValue extends LocalServiceValue {
  admin_user: string
  admin_passwd: string
  config: {
    innodb_buffer_pool_size: string
    resource_requests_memory: string
    resource_limits_memory: string
  }
}

export interface MariaDBManagedValue extends ManagedServiceValue {
  admin_user: string
  admin_passwd: string
  config: {
    innodb_buffer_pool_size: string
    resource_requests_memory: string
    resource_limits_memory: string
  }
}

export interface RedisLocalValue extends LocalServiceValue {
  admin_user: string
  admin_passwd: string
  resources?: {
    limits?: {
      cpu: string
      memory: string
    }
    requests?: {
      cpu: string
      memory: string
    }
  }
}

export interface RedisManagedValue extends ManagedServiceValue {
  admin_user: string
  admin_passwd: string
  resources?: {
    limits?: {
      cpu: string
      memory: string
    }
    requests?: {
      cpu: string
      memory: string
    }
  }
}

export interface OpenSearchLocalValue extends LocalServiceValue {
  mode: 'master' | 'hot' | 'warm'
  config: {
    jvmOptions: string
    hanlpRemoteextDict?: string
    hanlpRemoteextStopwords?: string
  }
  settings: Record<string, string>
}

export interface OpenSearchManagedValue extends ManagedServiceValue {
  mode: 'master' | 'hot' | 'warm'
  config: {
    jvmOptions: string
    hanlpRemoteextDict?: string
    hanlpRemoteextStopwords?: string
  }
  settings: Record<string, string>
}

export interface KafkaLocalValue extends LocalServiceValue {
  env: {
    KAFKA_HEAP_OPTS: string
    KAFKA_LOG_RETENTION_BYTES?: string
    KAFKA_LOG_RETENTION_HOURS?: string
    KAFKA_LOG_ROLL_HOURS?: string
  }
  disable_external_service?: boolean
  external_service_list?: ExternalServiceConfig[]
}

export interface KafkaManagedValue extends ManagedServiceValue {
  env: {
    KAFKA_HEAP_OPTS: string
    KAFKA_LOG_RETENTION_BYTES?: string
    KAFKA_LOG_RETENTION_HOURS?: string
    KAFKA_LOG_ROLL_HOURS?: string
  }
  disable_external_service?: boolean
  external_service_list?: ExternalServiceConfig[]
}

export interface ZooKeeperLocalValue extends LocalServiceValue {
  env: {
    JVMFLAGS: string
  }
  resources: {
    limits: {
      cpu: string
      memory: string
    }
    requests: {
      cpu: string
      memory: string
    }
  }
}

export interface ZooKeeperManagedValue extends ManagedServiceValue {
  env: {
    JVMFLAGS: string
  }
  resources: {
    limits: {
      cpu: string
      memory: string
    }
    requests: {
      cpu: string
      memory: string
    }
  }
}

export interface PrometheusLocalValue {
  hosts: string[]
  data_path: string
  storage_capacity: string
  resources?: {
    limits?: {
      cpu: string
      memory: string
    }
    requests?: {
      cpu: string
      memory: string
    }
  }
  storageClassName: string
}

export interface PrometheusManagedValue {
  replica_count: number
  storage_capacity: string
  storageClassName: string
  resources?: {
    limits?: {
      cpu: string
      memory: string
    }
    requests?: {
      cpu: string
      memory: string
    }
  }
}

export interface GrafanaLocalValue {
  hosts: string[]
  data_path: string
  storage_capacity: string
  resources?: {
    limits?: {
      cpu: string
      memory: string
    }
    requests?: {
      cpu: string
      memory: string
    }
  }
  storageClassName: string
}

export interface GrafanaManagedValue {
  replica_count: number
  storage_capacity: string
  storageClassName: string
  resources?: {
    limits?: {
      cpu: string
      memory: string
    }
    requests?: {
      cpu: string
      memory: string
    }
  }
}

export interface RdsConnectInfoFormValue {
  source_type: SourceType | ''
  rds_type: string
  auto_create_database: boolean
  admin_user: string
  admin_passwd: string
  hosts: string
  port: number | null
  username: string
  password: string
}

export interface RedisConnectInfoFormValue {
  source_type: SourceType | ''
  connect_type: RedisConnectType
  username: string
  password: string
  sentinel_hosts: string
  sentinel_port: number | null
  master_group_name: string
  master_hosts: string
  master_port: number | null
  slave_hosts: string
  slave_port: number | null
  hosts: string
  port: number | null
}

export interface OpenSearchConnectInfoFormValue {
  source_type: SourceType | ''
  hosts: string
  port: number | null
  username: string
  password: string
  distribution: string
  version: string
}

export interface MQConnectInfoFormValue {
  source_type: SourceType | ''
  mq_type: MQType
  mq_hosts: string
  mq_port: number | null
  mq_lookupd_hosts: string
  mq_lookupd_port: number | null
  auth: {
    username: string
    password: string
    mechanism: string
  }
}

export interface WizardState {
  deploymentKind: DeploymentKind
  deviceSpec: string
  service_package_dir: string
  nodes: NodeFormValue[]
  chrony: {
    mode: 'usermanaged' | 'localmaster' | 'externalntp'
    server: string[]
  }
  firewall: {
    mode: 'usermanaged' | 'firewalld'
  }
  cs: {
    local: LocalCsFormValue
    managed: ManagedCsFormValue
  }
  cr: {
    local: LocalCrFormValue
    external: ExternalCrFormValue
  }
  services: {
    mariadb: {
      local: MariaDBLocalValue
      managed: MariaDBManagedValue
    }
    redis: {
      local: RedisLocalValue
      managed: RedisManagedValue
    }
    opensearch: {
      local: OpenSearchLocalValue
      managed: OpenSearchManagedValue
    }
    kafka: {
      local: KafkaLocalValue
      managed: KafkaManagedValue
    }
    zookeeper: {
      local: ZooKeeperLocalValue
      managed: ZooKeeperManagedValue
    }
    prometheus: {
      local: PrometheusLocalValue
      managed: PrometheusManagedValue
    }
    grafana: {
      local: GrafanaLocalValue
      managed: GrafanaManagedValue
    }
  }
  resource_connect_info: {
    rds: RdsConnectInfoFormValue
    redis: RedisConnectInfoFormValue
    opensearch: OpenSearchConnectInfoFormValue
    mq: MQConnectInfoFormValue
  }
}

export interface SubmitConfig {
  apiVersion: 'v1'
  nodes: Array<{
    name: string
    ip4?: string
    ip6?: string
    internal_ip?: string
  }>
  chrony: {
    mode: string
    server?: string[]
  }
  firewall: {
    mode: string
  }
  cs: Record<string, unknown>
  cr: Record<string, unknown>
  component_management: Record<string, never>
  proton_mariadb?: unknown
  proton_redis?: unknown
  opensearch?: unknown
  kafka?: unknown
  zookeeper?: unknown
  prometheus?: unknown
  grafana?: unknown
  resource_connect_info: Record<string, unknown>
}

export interface SubmitRequest {
  service_package_dir?: string
  cluster_config: SubmitConfig
}
