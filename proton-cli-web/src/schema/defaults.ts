import type { WizardState } from './config'

export const defaultWizardState: WizardState = {
  deploymentKind: 'local',
  deviceSpec: 'AS10000',
  service_package_dir: './service-package',
  nodes: [],
  chrony: {
    mode: 'usermanaged',
    server: [],
  },
  firewall: {
    mode: 'usermanaged',
  },
  cs: {
    local: {
      master: ['node1'],
      addons: ['ingress-nginx'],
      ingressNginx: {
        httpPort: 80,
        httpsPort: 443,
      },
      ipFamilies: ['IPv4'],
      enableDualStack: false,
      ha_port: 8443,
      host_network: {
        bip: '172.33.0.1/16',
        pod_network_cidr: '192.169.0.0/16',
        service_cidr: '10.96.0.0/12',
        ipv4_interface: '',
        ipv6_interface: '',
      },
      etcd_data_dir: '/sysvol/proton_data/cs_etcd_data',
      containerd_data_dir: '/sysvol/proton_data/cs_containerd_data',
    },
    managed: {
      namespace: '',
      serviceaccount: '',
      addons: ['ingress-nginx'],
      ingressNginx: {
        httpPort: 80,
        httpsPort: 443,
      },
    },
  },
  cr: {
    local: {
      hosts: ['node1'],
      ports: {
        chartmuseum: 5001,
        registry: 5000,
        rpm: 5003,
        cr_manager: 5002,
      },
      ha_ports: {
        chartmuseum: 15001,
        registry: 15000,
        rpm: 15003,
        cr_manager: 15002,
      },
      storage: '/sysvol/proton_data/cr_data',
    },
    external: {
      chart_repository: 'chartmuseum',
      image_repository: 'registry',
      registry: {
        host: '',
        username: '',
        password: '',
      },
      chartmuseum: {
        host: '',
        username: '',
        password: '',
      },
      oci: {
        registry: '',
        username: '',
        password: '',
        plain_http: false,
      },
    },
  },
  services: {
    mariadb: {
      local: {
        enabled: true,
        hosts: ['node1'],
        data_path: '/sysvol/components/mariadb',
        storage_capacity: '',
        admin_user: 'root',
        admin_passwd: '',
        config: {
          innodb_buffer_pool_size: '1G',
          resource_requests_memory: '2G',
          resource_limits_memory: '2G',
        },
      },
      managed: {
        enabled: true,
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        admin_user: 'root',
        admin_passwd: '',
        config: {
          innodb_buffer_pool_size: '1G',
          resource_requests_memory: '2G',
          resource_limits_memory: '2G',
        },
      },
    },
    redis: {
      local: {
        enabled: true,
        hosts: ['node1'],
        data_path: '/sysvol/components/redis',
        storage_capacity: '',
        admin_user: 'root',
        admin_passwd: '',
        resources: {
          limits: {
            cpu: '500m',
            memory: '512Mi',
          },
          requests: {
            cpu: '100m',
            memory: '30Mi',
          },
        },
      },
      managed: {
        enabled: true,
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        admin_user: 'root',
        admin_passwd: '',
        resources: {
          limits: {
            cpu: '500m',
            memory: '512Mi',
          },
          requests: {
            cpu: '100m',
            memory: '30Mi',
          },
        },
      },
    },
    opensearch: {
      local: {
        enabled: true,
        hosts: ['node1'],
        data_path: '/sysvol/components/opensearch',
        storage_capacity: '',
        mode: 'master',
        config: {
          jvmOptions: '-Xmx1g -Xms1g',
          hanlpRemoteextDict: '',
          hanlpRemoteextStopwords: '',
        },
        settings: {
          'cluster.routing.allocation.disk.watermark.flood_stage': '95%',
          'cluster.routing.allocation.disk.watermark.high': '90%',
          'cluster.routing.allocation.disk.watermark.low': '85%',
        },
      },
      managed: {
        enabled: true,
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        mode: 'master',
        config: {
          jvmOptions: '-Xmx1g -Xms1g',
          hanlpRemoteextDict: '',
          hanlpRemoteextStopwords: '',
        },
        settings: {
          'cluster.routing.allocation.disk.watermark.flood_stage': '95%',
          'cluster.routing.allocation.disk.watermark.high': '90%',
          'cluster.routing.allocation.disk.watermark.low': '85%',
        },
      },
    },
    kafka: {
      local: {
        enabled: true,
        hosts: ['node1'],
        data_path: '/sysvol/components/kafka',
        storage_capacity: '',
        env: {
          KAFKA_HEAP_OPTS: '-Xms1g -Xmx1g',
          KAFKA_LOG_RETENTION_BYTES: '',
          KAFKA_LOG_RETENTION_HOURS: '',
          KAFKA_LOG_ROLL_HOURS: '',
        },
        disable_external_service: false,
        external_service_list: [{ name: 'kafka-external', ip: '', port: 30092, enableSSL: false }],
      },
      managed: {
        enabled: true,
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        env: {
          KAFKA_HEAP_OPTS: '-Xms1g -Xmx1g',
          KAFKA_LOG_RETENTION_BYTES: '',
          KAFKA_LOG_RETENTION_HOURS: '',
          KAFKA_LOG_ROLL_HOURS: '',
        },
        disable_external_service: false,
        external_service_list: [{ name: 'kafka-external', ip: '', port: 30092, enableSSL: false }],
      },
    },
    zookeeper: {
      local: {
        enabled: true,
        hosts: ['node1'],
        data_path: '/sysvol/components/zookeeper',
        storage_capacity: '',
        env: {
          JVMFLAGS: '-Xms500m -Xmx500m',
        },
        resources: {
          limits: {
            cpu: '1',
            memory: '2Gi',
          },
          requests: {
            cpu: '100m',
            memory: '270Mi',
          },
        },
      },
      managed: {
        enabled: true,
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        env: {
          JVMFLAGS: '-Xms500m -Xmx500m',
        },
        resources: {
          limits: {
            cpu: '1',
            memory: '2Gi',
          },
          requests: {
            cpu: '100m',
            memory: '270Mi',
          },
        },
      },
    },
    prometheus: {
      local: {
        hosts: ['node1'],
        data_path: '/sysvol/components/prometheus',
        storage_capacity: '',
        resources: {
          limits: {
            cpu: '500m',
            memory: '1Gi',
          },
          requests: {
            cpu: '100m',
            memory: '256Mi',
          },
        },
        storageClassName: '',
      },
      managed: {
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        resources: {
          limits: {
            cpu: '500m',
            memory: '1Gi',
          },
          requests: {
            cpu: '100m',
            memory: '256Mi',
          },
        },
      },
    },
    grafana: {
      local: {
        hosts: ['node1'],
        data_path: '/sysvol/components/grafana',
        storage_capacity: '',
        resources: {
          limits: {
            cpu: '200m',
            memory: '256Mi',
          },
          requests: {
            cpu: '100m',
            memory: '128Mi',
          },
        },
        storageClassName: '',
      },
      managed: {
        replica_count: 1,
        storage_capacity: '',
        storageClassName: 'csi-disk',
        resources: {
          limits: {
            cpu: '200m',
            memory: '256Mi',
          },
          requests: {
            cpu: '100m',
            memory: '128Mi',
          },
        },
      },
    },
  },
  resource_connect_info: {
    rds: {
      source_type: 'external',
      rds_type: 'MySQL',
      auto_create_database: true,
      admin_user: '',
      admin_passwd: '',
      hosts: '',
      port: null,
      username: 'root',
      password: '',
    },
    redis: {
      source_type: 'external',
      connect_type: 'standalone',
      username: '',
      password: '',
      sentinel_hosts: '',
      sentinel_port: null,
      master_group_name: '',
      master_hosts: '',
      master_port: null,
      slave_hosts: '',
      slave_port: null,
      hosts: '',
      port: null,
    },
    opensearch: {
      source_type: 'external',
      hosts: '',
      port: null,
      username: '',
      password: '',
      distribution: 'elasticsearch',
      version: '7.10.0',
    },
    mq: {
      source_type: 'external',
      mq_type: 'kafka',
      mq_hosts: '',
      mq_port: null,
      mq_lookupd_hosts: '',
      mq_lookupd_port: null,
      auth: {
        username: '',
        password: '',
        mechanism: 'PLAIN',
      },
    },
  },
}
