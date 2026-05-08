# proton-cli

`proton-cli` 是 Proton 的命令行工具，用于集群部署、配置管理、备份恢复、Kubernetes 维护以及离线包操作。

本 README 基于当前目录中已经提交到仓库的源代码编写，重点说明 CLI 的使用方式，不假设仓库之外还存在其他未纳入版本控制的打包产物。

## 💬 交流社区

<div align="center">
<img src="./docs/qrcode.png" alt="KWeaver 交流群二维码" width="30%"/>

扫码加入 KWeaver 交流群，联系群主获取 Kweaver 离线安装包
</div>

## 快速开始

### 操作系统兼容性
OpenEuler 24 x86/arm、银河麒麟V10 SP3 x86/arm、 RHEL8.10 X86、OpenEuler 23 x86/arm

### 离线安装快速开始

以下步骤适用于通过预构建二进制和离线依赖包快速安装 Proton。

1. 下载 `proton-cli`

请从 GitHub Releases 下载最新版本，并根据机器架构选择对应的压缩包。以下以 `v0.1.0` 为例：

```bash
# amd64
wget https://github.com/kweaver-ai/proton/releases/download/v0.1.0/proton-cli-linux-amd64.tar.gz

# arm64
wget https://github.com/kweaver-ai/proton/releases/download/v0.1.0/proton-cli-linux-arm64.tar.gz

tar xf proton-cli-linux-<arch>.tar.gz
```

2. 下载 Proton 离线依赖包

同样需要根据机器架构选择对应的离线包：

```bash
# amd64
proton-cli oras pull swr.cn-east-3.myhuaweicloud.com/kweaver-ai/offline-pkg/proton-offline-package:24653585065-amd64

# arm64
proton-cli oras pull swr.cn-east-3.myhuaweicloud.com/kweaver-ai/offline-pkg/proton-offline-package:24653585065-arm64
```

3. 安装离线包

```bash
tar xf proton-offline-package.tar
bash install.sh
```

4. 启动初始化服务

关闭防火墙

```bash
proton-cli server -ldebug -s proton-offline/service-package
```

然后通过浏览器访问 `http://<ip>:8888` 进入初始化页面。

5. 完成初始化

在初始化页面中自定义 `MariaDB` 和 `Redis` 密码，最后点击“完成”，等待初始化成功。

6. 导出应用离线安装包

可以通过以下命令下载指定 `yaml` 文件中定义的 chart 和 images，并打包为离线安装包：

```bash
proton-cli app export \
  -f release-manifest/x.y.z/kweaver-dip.yaml \
  --cache-dir /tmp/cache \
  --override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

7. 导入应用离线安装包

通过以下命令将离线应用包中的 chart 和 image 导入到内置仓库：

```bash
proton-cli app import \
  -i kweaver-dip-x.y.z-offline-package.tar \
  --auto --force
```

8. 安装指定产品版本

通过以下命令将指定产品的指定版本安装到 `kweaver` 命名空间：

```bash
proton-cli app install \
  -f release-manifest/x.y.z/kweaver-dip.yaml \
  -n kweaver
```

如果需要在安装时覆盖 values，例如关闭可选能力：

```bash
proton-cli app install \
  -f release-manifest/x.y.z/kweaver-dip.yaml \
  -n kweaver \
  --set auth.enabled=false \
  --set businessDomain.enabled=false
```

`--set` 用于安装时覆盖 values；详细的优先级规则以及与 manifest YAML 的关系，见后文 `app install` 说明。

### 环境要求

- Linux 环境
- 从源码构建时需要 Go 1.24 或更高版本
- 可用的 Kubernetes 客户端配置文件，路径为 `~/.kube/config`
- 使用 `apply` 时，需要一个至少包含以下内容的 `service-package` 目录：
  - `charts/`
  - `images/`
- 如果要使用 `edit conf`，请设置 `EDITOR`；否则默认使用 `vi`

### 构建

```bash
cd proton-cli
go build -o bin/proton-cli ./cmd/proton-cli
./bin/proton-cli --help
```

### 全局参数

大多数命令都支持以下参数：

- `-l, --log-level`：日志级别，例如 `info`、`debug`、`error`
- `-s, --service-package`：`service-package` 目录路径，默认值为 `service-package`
- `--service-package-eceph`：`service-package-eceph` 目录路径，默认值为 `service-package-eceph`
- `--cm-direct`：启用 component-manage 直连模式

## 常见工作流

### 1. 生成集群配置模板

显示内置模板之一：

```bash
proton-cli get template --type internal
proton-cli get template --type external
proton-cli get template --type perfrec
```

将模板输出到文件：

```bash
proton-cli get template --type internal > cluster.yaml
```

代码中内置的模板类型如下：

- `internal`：部署本地集群
- `external`：部署到已有 Kubernetes 集群
- `perfrec`：推荐配置参考

### 2. 应用集群配置

应用一个配置文件：

```bash
proton-cli apply -f cluster.yaml
```

显式指定部署命名空间：

```bash
proton-cli apply -f cluster.yaml -n proton
```

自定义 `ingress-nginx` addon 的 HTTP/HTTPS 端口：

```yaml
cs:
  addons:
    - ingress-nginx
  addonsConfig:
    ingress-nginx:
      httpPort: 8080
      httpsPort: 8443
```

这里会映射到 `ingress-nginx` chart 的 `controller.containerPort.http/https`。

显式指定 service package 路径：

```bash
proton-cli apply \
  -f cluster.yaml \
  -s /path/to/service-package \
  --service-package-eceph /path/to/service-package-eceph
```

根据代码确认的行为：

- `apply` 会从 `-f` 读取 YAML 文件
- `apply` 会从 `service-package` 加载包内容
- 如果传入 `-n`，会覆盖配置文件中的命名空间
- 配置文件中的 `deploy` 目前仅兼容 `namespace` 和 `serviceaccount` 字段；旧的 `mode` 与 `devicespec` 字段已移除，若仍存在会被忽略
- 如果最终选择了某个命名空间，`apply` 会更新本地文件 `~/.proton-cli.yaml`
- `apply` 成功后，会把集群配置上传到 Kubernetes

### 3. 查看当前已保存的集群配置

从 Kubernetes 读取当前配置：

```bash
proton-cli get conf
```

从指定命名空间读取：

```bash
proton-cli get conf -n proton
```

`get conf` 会从 Kubernetes Secret `proton-cli-config` 中读取配置。

### 4. 编辑 Kubernetes 中保存的配置

在编辑器中打开当前保存的配置：

```bash
proton-cli edit conf
```

编辑指定命名空间中的配置：

```bash
proton-cli edit conf -n proton
```

注意：

- 该命令会直接修改 Secret 中保存的内容
- 它会使用 `$EDITOR`，如果未设置则使用 `vi`
- 命令本身会输出：`ONLY change secrets, not apply!`

### 5. 备份与恢复

创建备份：

```bash
proton-cli backup create --resources all
```

指定备份名称和保留数量：

```bash
proton-cli backup create \
  --backupname nightly-001 \
  --resources all \
  --ttl 3
```

查看备份列表和日志：

```bash
proton-cli backup list
proton-cli backup log
proton-cli backup directory
proton-cli backup schedule
```

基于备份创建恢复任务：

```bash
proton-cli recover create \
  --from-backup nightly-001 \
  --resources all
```

查看恢复列表和日志：

```bash
proton-cli recover list
proton-cli recover log
```

### 6. 构建并安装离线包

输出内置 manifest 模板：

```bash
proton-cli offline-package plan > manifest.yaml
```

指定 CPU 架构输出 manifest 模板：

```bash
proton-cli offline-package plan --architecture arm64 > manifest.yaml
```

基于 manifest 构建离线包：

```bash
proton-cli offline-package build --manifest manifest.yaml
```

构建命令会在当前目录生成 `proton-offline-package.tar`。

安装离线包：

```bash
proton-cli offline-package install proton-offline-package.tar
```

保留解压后的工作目录：

```bash
proton-cli offline-package install proton-offline-package.tar --remain
```

根据代码确认的行为：

- `build` 默认读取 `manifest.yaml`
- `plan` 会把生成的 manifest 模板输出到标准输出，并支持用 `--architecture` 切换 `amd64` 或 `arm64`
- `install` 会先解压到 `.proton-offline-package/`，然后执行 `install.sh`

### 7. 导出、导入并安装应用离线包

`app` 子命令用于围绕 `VersionSet` manifest 执行应用离线包导出、导入、安装和卸载。

导出应用离线包：

```bash
proton-cli app export \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  --cache-dir /tmp/cache \
  --override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

如果需要显式指定输出文件名：

```bash
proton-cli app export \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -o kweaver-dip-0.5.0-offline-package.tar \
  --cache-dir /tmp/cache \
  --override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

导入应用离线包时，建议启用 `--auto`，自动探测当前 Proton 集群的内置镜像仓库和 ChartMuseum：

```bash
proton-cli app import -i kweaver-dip-0.5.0-offline-package.tar --auto
```

如需覆盖已存在的 chart：

```bash
proton-cli app import -i kweaver-dip-0.5.0-offline-package.tar --auto --force
```

导入完成后安装应用：

```bash
proton-cli app install \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver
```

仅查看安装计划而不真正执行：

```bash
proton-cli app install \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver \
  --dry-run
```

指定完整访问地址进行安装：

```bash
proton-cli app install \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver \
  --access-address=https://1.1.1.1:8443/
```

安装时向所有 chart 传入统一 values，并按 values 控制可选依赖：

```bash
proton-cli app install \
  -f release-manifests/0.6.0/kweaver-core.yaml \
  -n kweaver \
  --set auth.enabled=false \
  --set businessDomain.enabled=false
```

卸载应用：

```bash
proton-cli app uninstall \
  -f kelu/deploy/release-manifests/0.5.0/kweaver-dip.yaml \
  -n kweaver
```

`app` 相关常用参数：

- `app export`
- `-f, --manifest`：VersionSet manifest 路径
- `-o, --output`：输出 tar 包文件名；不指定时会根据 manifest 自动生成
- `--cache-dir`：导出镜像缓存目录，适合重复导出同一批镜像时复用
- `--override-registry`：导出时覆盖 chart values 中的 `.image.registry`
- `--platform`：导出目标平台，默认自动检测当前主机架构
- `--disable-dependencies`：只处理根 manifest，不递归处理 dependencies
- `--ignore-missing-images`：遇到部分镜像拉取失败时继续导出
- `app import`
- `--auto`：自动探测当前 Proton 集群配置中的镜像仓库和 ChartMuseum，推荐默认带上
- `--force`：覆盖 ChartMuseum 中已存在的 chart
- `--registry` / `--chartmuseum-url`：手工指定导入目标；未使用 `--auto` 时需要显式提供
- `app install`
- `-f, --file`：安装所用的 VersionSet manifest 路径
- `-n, --namespace`：目标 Kubernetes namespace，默认值为 `kweaver`
- `--timeout`：单个 release 的安装超时时间，默认 `30m`
- `--dry-run`：只打印安装计划，不执行安装
- `--config`：无法从集群读取 `proton-cli-config` 时，改为从本地配置文件加载 values
- `--image-registry`：安装时覆盖 values 中的镜像仓库地址
- `--set`：以 `key=value` 形式传入安装时 values，可重复传递，支持点路径，如 `auth.enabled=false`
- `--access-address`：以完整 URL 形式覆盖访问地址，例如 `https://1.1.1.1:8443/`；未显式指定端口时会按 `https=443`、`http=80` 补默认值
- `--access-host` / `--access-port` / `--access-scheme`：兼容旧用法，按字段覆盖访问地址；未显式传参时会保留现有 values 中的对应字段
- `app uninstall`
- `-f, --file`：卸载所用的 VersionSet manifest 路径
- `-n, --namespace`：目标 Kubernetes namespace，默认值为 `kweaver`
- `--timeout`：单个 release 的卸载超时时间，默认 `1m`
- `--dry-run`：只打印卸载计划，不执行卸载

根据代码确认的行为：

- `app export` 复用了 `offline-package app export` 的实现，支持 `--cache-dir` 和 `--override-registry`
- `app import` 复用了 `offline-package app import` 的实现，推荐始终带 `--auto`
- `app install` 成功后会打印安装完成提示，并显示所使用的 manifest 和 namespace
- `app install` 会先展开 manifest dependencies，再根据 `stage: pre` 和 `dependsOn` 生成安装顺序
- `app install` 会把 cluster config、`--set` 和 release 自身的 `values` 合并后传给 chart，优先级为 `release.values > --set > cluster config`
- `--set` 与 YAML 的关系如下：

```yaml
dependencies:
  - product: isf
    version: 0.6.0
    manifest: ./isf.yaml
    optional: true
    defaultEnabled: true
    enabledIf: auth.enabled

releases:
  mf-model-manager:
    chart: mf-model-manager
    version: 0.6.0
    values:
      auth:
        enabled: true
```

  对应命令：

```bash
proton-cli app install \
  -f release-manifests/0.6.0/kweaver-core.yaml \
  --set auth.enabled=false \
  --set businessDomain.enabled=false
```

  语义：
  `--set auth.enabled=false` 会作为全局 values 传给所有 chart；如果某个 release 在 manifest 的 `values` 里再次定义 `auth.enabled`，则以该 release 自己的 `values` 为准。
  `dependencies[].enabledIf: auth.enabled` 会读取同一份合并后的 install-time values；当值为 `false` 时跳过该 dependency，当未传值时回退到 `defaultEnabled`。
- `dependsOn` 表示同一 manifest 内 release 之间的 ready 依赖；被依赖的 release 会先安装并等待就绪
- optional dependency 可通过 manifest 中的 `enabledIf` 结合 `--set` 控制，例如：

```yaml
dependencies:
  - product: isf
    version: 0.6.0
    manifest: ./isf.yaml
    defaultEnabled: true
    enabledIf: auth.enabled
```
- `app uninstall` 会按安装计划的逆序执行卸载

### 8. Kubernetes 工具

查看当前 Kubernetes 相关状态：

```bash
proton-cli kubernetes show
```

升级 Calico：

```bash
proton-cli kubernetes calico upgrade <version>
```

该命令会先根据内置支持版本列表校验目标版本，再开始升级。

### 9. Shell 补全与版本信息

输出版本信息：

```bash
proton-cli version
```

启用 shell 补全：

```bash
proton-cli completion bash
proton-cli completion zsh
proton-cli completion fish
proton-cli completion powershell
```

## 命令概览

当前已提交代码暴露的根命令包括：

- `apply`：应用集群配置文件
- `get`：查看当前配置或内置模板
- `edit`：编辑 Kubernetes 中保存的配置
- `backup`：创建并查看备份
- `recover`：创建并查看恢复任务
- `offline-package`：输出 manifest 模板、构建离线包、安装离线包
- `app`：导出、导入、安装和卸载基于 VersionSet 的应用包
- `kubernetes`：查看 Kubernetes 状态并管理 Calico
- `completion`：生成 shell 补全脚本
- `version`：输出版本信息
- `precheck`：安装前检查节点环境
- `reset`：重置 Proton 集群
- `migrate`：迁移由其他程序部署的组件
- `check`：检查 Proton 运行时健康状态
- `images`：管理镜像
- `push-images`：把镜像推送到仓库
- `push-charts`：把 chart 推送到仓库
- `package`：与打包相关的命令
- `component`：数据组件管理
- `delete-images`：删除镜像以回收磁盘空间
- `server`：以 CLI Server 模式运行
- `alpha`：实验性命令

按需查看帮助：

```bash
proton-cli --help
proton-cli <command> --help
```

## 运行时状态与默认路径

当前已提交代码使用以下默认路径和名称：

- Kubernetes 客户端配置：`~/.kube/config`
- 本地 Proton CLI 环境文件：`~/.proton-cli.yaml`
- Proton CLI 默认配置命名空间：`proton`
- Proton 资源默认命名空间：`resource`
- 保存集群配置的 Secret：`proton-cli-config`
- Secret 中的集群配置字段名：`ClusterConfiguration`

实际行为上：

- `get conf` 和 `edit conf` 都会从 `proton-cli-config` 读取配置
- `apply` 可能会更新 `~/.proton-cli.yaml`
- 对命名空间敏感的命令通常会默认使用 `~/.proton-cli.yaml` 中记录的命名空间；如果该文件不存在，则默认使用 `proton`

## 说明

- 本 README 有意不假设仓库中一定已经包含可直接使用的 `service-package` 预构建产物。
- 如果你不确定某个命令怎么用，优先查看命令自身的帮助输出，而不是依赖旧脚本或占位文档。
