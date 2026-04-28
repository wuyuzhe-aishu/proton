import type {
  ExternalCrRepository,
  OpenSearchConnectInfoFormValue,
  RedisConnectInfoFormValue,
  RdsConnectInfoFormValue,
  WizardState,
} from '../schema/config'

export interface ValidationIssue {
  field: string
  message: string
}

export interface ValidationResult {
  valid: boolean
  issues: ValidationIssue[]
}

function push(issues: ValidationIssue[], field: string, message: string) {
  issues.push({ field, message })
}

function isBlank(value: string | null | undefined) {
  return !value || value.trim() === ''
}

function hasValidPort(value: number | null) {
  return Number.isInteger(value) && value !== null && value > 0 && value <= 65535
}

function validateIngressNginxPorts(
  value: { addons: string[]; ingressNginx: { httpPort: number; httpsPort: number } },
  field: string,
  issues: ValidationIssue[],
) {
  if (!value.addons.includes('ingress-nginx')) {
    return
  }

  if (!hasValidPort(value.ingressNginx.httpPort)) {
    push(issues, `${field}.ingressNginx.httpPort`, 'ingress-nginx HTTP port must be a valid TCP port.')
  }
  if (!hasValidPort(value.ingressNginx.httpsPort)) {
    push(issues, `${field}.ingressNginx.httpsPort`, 'ingress-nginx HTTPS port must be a valid TCP port.')
  }
}

function validateNodeSection(state: WizardState, issues: ValidationIssue[]) {
  if (!state.nodes.length) {
    push(issues, 'nodes', 'At least one node is required.')
    return
  }

  const names = new Set<string>()
  const ipv4 = new Set<string>()
  const ipv6 = new Set<string>()

  state.nodes.forEach((node, index) => {
    const name = node.name.trim()

    if (!name || !/^[a-z]([-a-z0-9]*[a-z0-9])?$/.test(name)) {
      push(issues, `nodes[${index}].name`, 'Nodes must include a valid name and at least one IP address.')
    }

    if (!node.ip4 && !node.ip6) {
      push(issues, `nodes[${index}]`, 'Nodes must include a valid name and at least one IP address.')
    }

    if (name && names.has(name)) {
      push(issues, `nodes[${index}].name`, 'Node names must be unique.')
    }
    names.add(name)

    if (node.ip4) {
      if (ipv4.has(node.ip4)) {
        push(issues, `nodes[${index}].ip4`, 'IPv4 addresses must be unique.')
      }
      ipv4.add(node.ip4)
    }

    if (node.ip6) {
      if (ipv6.has(node.ip6)) {
        push(issues, `nodes[${index}].ip6`, 'IPv6 addresses must be unique.')
      }
      ipv6.add(node.ip6)
    }
  })
}

function validateLocalCluster(state: WizardState, issues: ValidationIssue[]) {
  const nodeNames = new Set(state.nodes.map((node) => node.name))
  const { local } = state.cs

  if (!local.master.length) {
    push(issues, 'cs.local.master', 'Local Kubernetes requires at least one master node.')
  }

  local.master.forEach((node) => {
    if (!nodeNames.has(node)) {
      push(issues, 'cs.local.master', 'Master nodes must reference declared nodes.')
    }
  })

  if (!local.ipFamilies.length) {
    push(issues, 'cs.local.ipFamilies', 'At least one IP family is required.')
  }

  const includesIPv6 = local.ipFamilies.includes('IPv6')
  if (local.enableDualStack && local.ipFamilies.length !== 2) {
    push(issues, 'cs.local.ipFamilies', 'Dual stack requires IPv4 and IPv6.')
  }

  if (includesIPv6) {
    state.nodes.forEach((node, index) => {
      if (isBlank(node.ip6)) {
        push(issues, `nodes[${index}].ip6`, 'IPv6 is required when local cluster uses IPv6.')
      }
    })
  }

  if (state.chrony.mode === 'localmaster') {
    if (state.chrony.server.length !== 1 || !local.master.includes(state.chrony.server[0])) {
      push(issues, 'chrony.server', 'Chrony localmaster mode must point to exactly one master node.')
    }
  }

  if (state.chrony.mode === 'externalntp' && !state.chrony.server.length) {
    push(issues, 'chrony.server', 'External NTP mode requires at least one time server.')
  }

  if (state.chrony.mode === 'usermanaged' && state.chrony.server.length) {
    push(issues, 'chrony.server', 'User managed chrony should not include time servers.')
  }

  validateIngressNginxPorts(local, 'cs.local', issues)
}

function validateExternalRepo(repo: ExternalCrRepository, field: string, label: string, issues: ValidationIssue[]) {
  if (isBlank(repo.host)) {
    push(issues, `${field}.host`, `${label} host is required.`)
  }
  if (isBlank(repo.username)) {
    push(issues, `${field}.username`, `${label} username is required.`)
  }
  if (isBlank(repo.password)) {
    push(issues, `${field}.password`, `${label} password is required.`)
  }
}

function validateManagedCluster(state: WizardState, issues: ValidationIssue[]) {
  if (isBlank(state.cs.managed.namespace)) {
    push(issues, 'cs.managed.namespace', 'Managed Kubernetes requires a namespace.')
  }

  if (isBlank(state.cs.managed.serviceaccount)) {
    push(issues, 'cs.managed.serviceaccount', 'Managed Kubernetes requires a serviceaccount.')
  }

  const { external } = state.cr

  if (external.chart_repository === 'chartmuseum') {
    validateExternalRepo(external.chartmuseum, 'cr.external.chartmuseum', 'Chartmuseum', issues)
  }

  if (external.image_repository === 'registry') {
    validateExternalRepo(external.registry, 'cr.external.registry', 'Registry', issues)
  }

  if (external.chart_repository === 'oci' || external.image_repository === 'oci') {
    if (isBlank(external.oci.registry)) {
      push(issues, 'cr.external.oci.registry', 'OCI registry is required.')
    }
    if (isBlank(external.oci.username)) {
      push(issues, 'cr.external.oci.username', 'OCI username is required.')
    }
    if (isBlank(external.oci.password)) {
      push(issues, 'cr.external.oci.password', 'OCI password is required.')
    }
  }

  validateIngressNginxPorts(state.cs.managed, 'cs.managed', issues)
}

function validateManagedService(
  replicaCount: number,
  field: string,
  label: string,
  issues: ValidationIssue[],
) {
  if (replicaCount < 1) {
    push(issues, `${field}.replica_count`, `Managed ${label} requires replica_count >= 1.`)
  }
}

function validateLocalService(
  hosts: string[],
  dataPath: string,
  field: string,
  label: string,
  nodeNames: Set<string>,
  issues: ValidationIssue[],
) {
  if (!hosts.length) {
    push(issues, `${field}.hosts`, `Local ${label} requires at least one host.`)
  }

  hosts.forEach((host) => {
    if (!nodeNames.has(host)) {
      push(issues, `${field}.hosts`, `Local ${label} hosts must reference declared nodes.`)
    }
  })

  if (isBlank(dataPath)) {
    push(issues, `${field}.data_path`, `Local ${label} requires a data path.`)
  }
}

function validateRequiredServicePassword(password: string, field: string, label: string, issues: ValidationIssue[]) {
  if (isBlank(password)) {
    push(issues, field, `${label} 密码为必填项。`)
  }
}

function validateRds(rds: RdsConnectInfoFormValue, issues: ValidationIssue[]) {
  if (rds.source_type === 'internal') {
    return
  }

  if (isBlank(rds.rds_type)) {
    push(issues, 'resource_connect_info.rds.rds_type', 'RDS type is required.')
  }
  if (isBlank(rds.hosts)) {
    push(issues, 'resource_connect_info.rds.hosts', 'RDS hosts are required.')
  }
  if (!hasValidPort(rds.port)) {
    push(issues, 'resource_connect_info.rds.port', 'RDS port must be a valid TCP port.')
  }
  if (isBlank(rds.username)) {
    push(issues, 'resource_connect_info.rds.username', 'RDS username is required.')
  }
  if (isBlank(rds.password)) {
    push(issues, 'resource_connect_info.rds.password', 'RDS password is required.')
  }
  if (rds.auto_create_database) {
    if (isBlank(rds.admin_user)) {
      push(issues, 'resource_connect_info.rds.admin_user', 'RDS admin username is required when auto create database is enabled.')
    }
    if (isBlank(rds.admin_passwd)) {
      push(issues, 'resource_connect_info.rds.admin_passwd', 'RDS admin password is required when auto create database is enabled.')
    }
  }
}

function validateRedis(redis: RedisConnectInfoFormValue, issues: ValidationIssue[]) {
  if (redis.source_type === 'internal') {
    return
  }

  if (redis.connect_type === 'sentinel') {
    if (isBlank(redis.sentinel_hosts)) {
      push(issues, 'resource_connect_info.redis.sentinel_hosts', 'Sentinel hosts are required.')
    }
    if (!hasValidPort(redis.sentinel_port)) {
      push(issues, 'resource_connect_info.redis.sentinel_port', 'Sentinel port must be valid.')
    }
    if (isBlank(redis.master_group_name)) {
      push(issues, 'resource_connect_info.redis.master_group_name', 'Master group name is required.')
    }
    return
  }

  if (redis.connect_type === 'master-slave') {
    if (isBlank(redis.master_hosts)) {
      push(issues, 'resource_connect_info.redis.master_hosts', 'Master hosts are required.')
    }
    if (!hasValidPort(redis.master_port)) {
      push(issues, 'resource_connect_info.redis.master_port', 'Master port must be valid.')
    }
    if (isBlank(redis.slave_hosts)) {
      push(issues, 'resource_connect_info.redis.slave_hosts', 'Slave hosts are required.')
    }
    if (!hasValidPort(redis.slave_port)) {
      push(issues, 'resource_connect_info.redis.slave_port', 'Slave port must be valid.')
    }
    return
  }

  if (redis.connect_type === 'standalone' || redis.connect_type === 'cluster') {
    if (isBlank(redis.hosts)) {
      push(issues, 'resource_connect_info.redis.hosts', 'Redis hosts are required.')
    }
    if (!hasValidPort(redis.port)) {
      push(issues, 'resource_connect_info.redis.port', 'Redis port must be valid.')
    }
    return
  }

  push(issues, 'resource_connect_info.redis.connect_type', 'Redis connect type is required.')
}

function validateOpenSearch(search: OpenSearchConnectInfoFormValue, issues: ValidationIssue[]) {
  if (search.source_type === 'internal') {
    return
  }

  if (isBlank(search.hosts)) {
    push(issues, 'resource_connect_info.opensearch.hosts', 'OpenSearch hosts are required.')
  }
  if (!hasValidPort(search.port)) {
    push(issues, 'resource_connect_info.opensearch.port', 'OpenSearch port must be valid.')
  }
  if (isBlank(search.username)) {
    push(issues, 'resource_connect_info.opensearch.username', 'OpenSearch username is required.')
  }
  if (isBlank(search.password)) {
    push(issues, 'resource_connect_info.opensearch.password', 'OpenSearch password is required.')
  }
  if (isBlank(search.distribution)) {
    push(issues, 'resource_connect_info.opensearch.distribution', 'OpenSearch distribution is required.')
  }
  if (!['5.6.4', '7.10.0'].includes(search.version)) {
    push(issues, 'resource_connect_info.opensearch.version', 'OpenSearch version must be 5.6.4 or 7.10.0.')
  }
}

function validateMq(state: WizardState, issues: ValidationIssue[]) {
  const mq = state.resource_connect_info.mq

  if (mq.source_type !== 'internal' && mq.source_type !== 'external') {
    push(issues, 'resource_connect_info.mq.source_type', 'MQ source type must be internal or external.')
    return
  }

  if (mq.source_type === 'internal') {
    if (mq.mq_type !== 'kafka') {
      push(issues, 'resource_connect_info.mq.mq_type', 'Only Kafka is supported as an internal MQ source here.')
    }
    return
  }

  if (isBlank(mq.mq_hosts)) {
    push(issues, 'resource_connect_info.mq.mq_hosts', 'MQ hosts are required.')
  }
  if (!hasValidPort(mq.mq_port)) {
    push(issues, 'resource_connect_info.mq.mq_port', 'MQ port must be valid.')
  }
  if (isBlank(mq.mq_type)) {
    push(issues, 'resource_connect_info.mq.mq_type', 'MQ type is required.')
  }

  if (mq.mq_type === 'nsq') {
    if (isBlank(mq.mq_lookupd_hosts)) {
      push(issues, 'resource_connect_info.mq.mq_lookupd_hosts', 'NSQ lookupd hosts are required.')
    }
    if (!hasValidPort(mq.mq_lookupd_port)) {
      push(issues, 'resource_connect_info.mq.mq_lookupd_port', 'NSQ lookupd port must be valid.')
    }
  }

  if (mq.mq_type === 'kafka') {
    if (isBlank(mq.auth.username)) {
      push(issues, 'resource_connect_info.mq.auth.username', 'Kafka username is required.')
    }
    if (isBlank(mq.auth.password)) {
      push(issues, 'resource_connect_info.mq.auth.password', 'Kafka password is required.')
    }
    if (!['PLAIN', 'SCRAM-SHA-512', 'SCRAM-SHA-256'].includes(mq.auth.mechanism)) {
      push(issues, 'resource_connect_info.mq.auth.mechanism', 'Kafka auth mechanism is invalid.')
    }
  }
}

function isManagedInternally(state: WizardState, service: 'mariadb' | 'redis' | 'opensearch' | 'mq') {
  if (service === 'mq') {
    return state.resource_connect_info.mq.source_type === 'internal'
  }

  if (service === 'mariadb') {
    return state.resource_connect_info.rds.source_type === 'internal'
  }

  if (service === 'redis') {
    return state.resource_connect_info.redis.source_type === 'internal'
  }

  return state.resource_connect_info.opensearch.source_type === 'internal'
}

export function validateWizardState(state: WizardState): ValidationResult {
  const issues: ValidationIssue[] = []
  const isLocal = state.deploymentKind === 'local'
  const nodeNames = new Set(state.nodes.map((node) => node.name))

  validateNodeSection(state, issues)

  if (isLocal) {
    validateLocalCluster(state, issues)
  } else {
    validateManagedCluster(state, issues)
  }

  if (state.firewall.mode !== 'usermanaged' && state.firewall.mode !== 'firewalld') {
    push(issues, 'firewall.mode', 'Firewall mode must be usermanaged or firewalld.')
  }

  if (isManagedInternally(state, 'mariadb')) {
    validateLocalService(
      state.services.mariadb.local.hosts,
      state.services.mariadb.local.data_path,
      'services.mariadb.local',
      'MariaDB',
      nodeNames,
      issues,
    )
    validateManagedService(
      state.services.mariadb.managed.replica_count,
      'services.mariadb.managed',
      'MariaDB',
      issues,
    )
    validateRequiredServicePassword(
      isLocal ? state.services.mariadb.local.admin_passwd : state.services.mariadb.managed.admin_passwd,
      isLocal ? 'services.mariadb.local.admin_passwd' : 'services.mariadb.managed.admin_passwd',
      'MariaDB',
      issues,
    )
  }

  if (isManagedInternally(state, 'redis')) {
    validateLocalService(
      state.services.redis.local.hosts,
      state.services.redis.local.data_path,
      'services.redis.local',
      'Redis',
      nodeNames,
      issues,
    )
    validateManagedService(
      state.services.redis.managed.replica_count,
      'services.redis.managed',
      'Redis',
      issues,
    )
    validateRequiredServicePassword(
      isLocal ? state.services.redis.local.admin_passwd : state.services.redis.managed.admin_passwd,
      isLocal ? 'services.redis.local.admin_passwd' : 'services.redis.managed.admin_passwd',
      'Redis',
      issues,
    )
  }

  if (isManagedInternally(state, 'opensearch')) {
    validateLocalService(
      state.services.opensearch.local.hosts,
      state.services.opensearch.local.data_path,
      'services.opensearch.local',
      'OpenSearch',
      nodeNames,
      issues,
    )
    validateManagedService(
      state.services.opensearch.managed.replica_count,
      'services.opensearch.managed',
      'OpenSearch',
      issues,
    )
  }

  if (isManagedInternally(state, 'mq')) {
    validateLocalService(
      state.services.kafka.local.hosts,
      state.services.kafka.local.data_path,
      'services.kafka.local',
      'Kafka',
      nodeNames,
      issues,
    )
    validateLocalService(
      state.services.zookeeper.local.hosts,
      state.services.zookeeper.local.data_path,
      'services.zookeeper.local',
      'ZooKeeper',
      nodeNames,
      issues,
    )
    validateManagedService(
      state.services.kafka.managed.replica_count,
      'services.kafka.managed',
      'Kafka',
      issues,
    )
    validateManagedService(
      state.services.zookeeper.managed.replica_count,
      'services.zookeeper.managed',
      'ZooKeeper',
      issues,
    )
  }

  const kafkaEnabled = isLocal ? state.services.kafka.local.enabled : state.services.kafka.managed.enabled
  const zookeeperEnabled = isLocal ? state.services.zookeeper.local.enabled : state.services.zookeeper.managed.enabled
  if (isManagedInternally(state, 'mq') && kafkaEnabled && !zookeeperEnabled) {
    push(issues, 'services.zookeeper', 'Kafka requires ZooKeeper.')
  }

  if (isLocal && isManagedInternally(state, 'mq') && state.services.zookeeper.local.enabled) {
    const hostCount = new Set(state.services.zookeeper.local.hosts).size
    if (![1, 3].includes(hostCount)) {
      push(issues, 'services.zookeeper.local.hosts', 'ZooKeeper local hosts only support 1 or 3 nodes.')
    }
  }

  validateLocalService(
    state.services.prometheus.local.hosts,
    state.services.prometheus.local.data_path,
    'services.prometheus.local',
    'Prometheus',
    nodeNames,
    issues,
  )
  validateManagedService(
    state.services.prometheus.managed.replica_count,
    'services.prometheus.managed',
    'Prometheus',
    issues,
  )

  validateLocalService(
    state.services.grafana.local.hosts,
    state.services.grafana.local.data_path,
    'services.grafana.local',
    'Grafana',
    nodeNames,
    issues,
  )
  validateManagedService(
    state.services.grafana.managed.replica_count,
    'services.grafana.managed',
    'Grafana',
    issues,
  )

  validateRds(state.resource_connect_info.rds, issues)
  validateRedis(state.resource_connect_info.redis, issues)
  validateOpenSearch(state.resource_connect_info.opensearch, issues)
  validateMq(state, issues)

  return {
    valid: issues.length === 0,
    issues,
  }
}
