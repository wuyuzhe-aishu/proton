#!/usr/bin/env bash
# -*- codinng: utf-8 -*-
TZ="Asia/Shanghai"
PROMETHEUS_NAMESPACE=""
PROTON_ETCD_NAMESPACE=""
SCRIPT_PATH=$(dirname "$(realpath "$0")")
# shellcheck disable=SC1091
LOG_DIR="/tmp/prometheus.log"

logger_output_private()
{
    if [ $# -le 0 ]; then
        return;
    fi
    info_level=$1
    msg_id=$2
    shift
    millisecond=$(expr $(date +%N) / 1000)
    debug_msg_head="$(TZ=$TZ date '+%Y-%m-%d %H:%M:%S').$(printf '%06d' ${millisecond}) [$$] - ${info_level} "
    debug_msg="${debug_msg_head}$@"
	  echo "${debug_msg}" >> ${LOG_DIR}
    echo "${debug_msg}"
}
logger_info()
{
    logger_output_private "INFO" "$@"
}
logger_error()
{
    logger_output_private "ERROR" "$@"
}

if [[ $# -le 0 ]];then
    echo "Usage: $0 PROMETHEUS_NAMESPACE PROTON_ETCD_NAMESPACE"
    exit 0
elif [[ $# = 1 ]];then
    PROMETHEUS_NAMESPACE=$1
else
    PROMETHEUS_NAMESPACE=$1
    PROTON_ETCD_NAMESPACE=$2
fi


updateEtcdCerts(){
    logger_info "更新Prometheus中Etcd证书"

    ETCD_CERTS_DIR=/etc/kubernetes/pki/etcd
    if [[ ! -d ${ETCD_CERTS_DIR} ]];then
      logger_error "本地证书目录【${ETCD_CERTS_DIR}】不存在!!!"
    fi

    CA_CRT=$(base64 < "${ETCD_CERTS_DIR}/ca.crt" | tr -d '\n')
    HC_KEY=$(base64 < "${ETCD_CERTS_DIR}/healthcheck-client.key" | tr -d '\n')
    HC_CRT=$(base64 < "${ETCD_CERTS_DIR}/healthcheck-client.crt" | tr -d '\n')

    kubectl get secrets etcd-certs -n "${PROMETHEUS_NAMESPACE}" >/dev/null
    if [[ $? -eq 1 ]];then
      logger_error "缺少etcd-certs!"
      return 0
    fi

    kubectl get secrets etcd-certs -n "${PROMETHEUS_NAMESPACE}" -o json \
            | sed 's/"ca.crt": ".*"/"ca.crt": "'$CA_CRT'"/g'\
            | sed 's/"healthcheck-client.crt": ".*"/"healthcheck-client.crt": "'$HC_CRT'"/g'\
            | sed 's/"healthcheck-client.key": ".*"/"healthcheck-client.key": "'$HC_KEY'"/g'\
            | kubectl replace -f - >/dev/null

    if [[ $? -eq 1 ]];then
      logger_error "更新etcd证书失败, 这将会导致Etcd监控无法正常使用!!!"
    else
      logger_info "更新Prometheus中Etcd证书成功."
    fi
}

updateProtonEtcdCerts(){
    logger_info "更新Prometheus中Proton-Etcd证书"

    kubectl get secrets etcdssl-secret -n "${PROTON_ETCD_NAMESPACE}" >/dev/null
    if [[ $? -eq 1 ]];then
      logger_error "当前命名空间下不存在etcdssl-sercet!"
      return 0
    fi

    CA_CRT=$(kubectl get secrets etcdssl-secret -n "${PROTON_ETCD_NAMESPACE}" -ojsonpath='{$.data.ca\.crt}')
    PEER_KEY=$(kubectl get secrets etcdssl-secret -n "${PROTON_ETCD_NAMESPACE}" -ojsonpath='{$.data.peer\.key}')
    PEER_CRT=$(kubectl get secrets etcdssl-secret -n "${PROTON_ETCD_NAMESPACE}" -ojsonpath='{$.data.peer\.crt}')

    kubectl get secrets proton-etcd-certs -n "${PROMETHEUS_NAMESPACE}" >/dev/null
    if [[ $? -eq 1 ]];then
      logger_error "缺少proton-etcd-certs!"
      return 0
    fi

    kubectl get secrets proton-etcd-certs -n "${PROMETHEUS_NAMESPACE}" -o json \
            | sed 's/"ca.crt": ".*"/"ca.crt": "'$CA_CRT'"/g'\
            | sed 's/"peer.crt": ".*"/"peer.crt": "'$PEER_CRT'"/g'\
            | sed 's/"peer.key": ".*"/"peer.key": "'$PEER_KEY'"/g'\
            | kubectl replace -f - >/dev/null

    if [[ $? -eq 1 ]];then
      logger_error "更新proton-etcd证书失败, 这将会导致Etcd监控无法正常使用!!!"
    else
      logger_info "更新Prometheus中Proton-Etcd证书成功."
    fi
}

if [ $PROMETHEUS_NAMESPACE ];then
  updateEtcdCerts
  if [ $PROTON_ETCD_NAMESPACE ];then
    updateProtonEtcdCerts
  else
    logger_info "PROTON_ETCD_NAMESPACE 为空"
  fi
fi