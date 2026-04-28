#!/usr/bin/env bash
# -*- codinng: utf-8 -*-
NAMESPACE=""
SCRIPT_PATH=$(dirname "$(realpath "$0")")
# shellcheck disable=SC1091
TZ="Asia/Shanghai"
LOG_DIR="/tmp/prometheus.log"
SLB_MANAGER_URL="http://localhost:9547"
SERVER_NAME="monitor"
LISTEN_PORT=8004
HC_PATH="/healthcheck_status"
STUB_STATUS_PATH="/stub_status"
NODE_LIST=($(kubectl get node -owide|awk 'NR>1{print $6}'))

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






initNginxMonitor(){
  local conf_name=$1
  local listen_port=$2
  local hc_path=$3
  local stub_status_path=$4
  local slb_manager_url=$5
  local request_body='''
    {
        "name": "'${conf_name}'",
        "conf": {
            "server": {
            "listen": "'${listen_port}'",
            "location": {
                "'${hc_path}'": {
                "healthcheck_status": "prometheus"
                },
                "'${stub_status_path}'": {
                "stub_status": "on"
                }
            }
            }
        }
    }'''

  local get_request=$(curl -s -w '|%{http_code}' -XGET "${slb_manager_url}/api/slb/v1/nginx/http/${conf_name}")
  if [[ "${get_request##*|}" -eq 200 ]];then
    request_body="$(echo "${request_body}"|sed 's/name:.*,//')"
    local response=$(echo "${request_body}"|curl -s -w '|%{http_code}' -XPUT "${slb_manager_url}/api/slb/v1/nginx/http/${conf_name}" -H "Content-Type: application/json" -d @-)
  else
    local response=$(echo "${request_body}"|curl -s -w '|%{http_code}' -XPOST "${slb_manager_url}/api/slb/v1/nginx/http" -H "Content-Type: application/json" -d @-)
  fi

  local res_code=${response##*|}
  local res_msg=${response%%|*}
  if [[ ${res_code} -ne 200 && ${res_code} -ne 204 ]];then
    echo "初始化Nginx监控接口失败，response:{code:${res_code}, msg:${res_msg}}"
    exit 1
  fi
}

Usage(){
    echo "Usage:"
    echo "$0 [-n SERVER_NAME] [-p PORT] [-c HC_PATH] [-s STUB_STATUS_PATH] "
    echo "Description:"
    echo "    -n, server_name for slb-nginx http server"
    echo "    -p, listen port for slb-nginx http server"
    echo "    -c, healthcheck_status context path for slb-nginx ngx-healthcheck-module"
    echo "    -s, stub_status context path for slb-nginx ngx_http_stub_status_module"
    echo "    -h, help message"
}


while getopts 'hn:p:c:s:' OPT; do
    case ${OPT} in
        n)
            SERVER_NAME="$OPTARG";;
        p)
            LISTEN_PORT="$OPTARG";;
        c)
            HC_PATH="$OPTARG";;
        s)
            STUB_STATUS_PATH="$OPTARG";;
        h)
            Usage
            exit 0 ;;
        ?)
            Usage 
            exit 0 ;;
     esac
done


logger_info "开始初始化Nginx监控接口"
logger_info """配置信息:
    SCRIPT_PATH=${SCRIPT_PATH}
    LOG_DIR=${LOG_DIR}
    SLB_MANAGER_URL=${SLB_MANAGER_URL}
    SERVER_NAME=${SERVER_NAME}
    LISTEN_PORT=${LISTEN_PORT}
    HC_PATH=${HC_PATH}
    STUB_STATUS_PATH=${STUB_STATUS_PATH}
    NODE_LIST=${NODE_LIST[@]}"""


for node in ${NODE_LIST[@]};do
    result=$(initNginxMonitor "${SERVER_NAME}" "${LISTEN_PORT}" "${HC_PATH}" "${STUB_STATUS_PATH}" "http://${node}:9547")
    if [[ $? -eq 0 ]];then
        logger_info """Nginx监控接口初始化成功: 
        healthcheck接口：http://${node}:${LISTEN_PORT}${HC_PATH}
        stub_status接口：http://${node}:${LISTEN_PORT}${STUB_STATUS_PATH}"""
    else
        logger_error "${result}"
    fi
done