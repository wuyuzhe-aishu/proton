#!/usr/bin/env bash
# -*- codinng: utf-8 -*-
#set -ex

CONFIG_FILE=/etc/grafana/provisioning-monitor/monitor.yaml
DASHBOARDS_DIR=/etc/grafana/provisioning-dashboards


OPENSEARCH_VARS_ROOT=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.opensearch.vars.root="|awk -F'=' '{print $2}')
OPENSEARCH_VARS_HOT=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.opensearch.vars.hot="|awk -F'=' '{print $2}')
OPENSEARCH_VARS_WARM=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.opensearch.vars.warm="|awk -F'=' '{print $2}')
OPENSEARCH_VARS_ROOTDEV=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.opensearch.vars.rootDev="|awk -F'=' '{print $2}')
OPENSEARCH_VARS_HOTDEV=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.opensearch.vars.hotDev="|awk -F'=' '{print $2}')
OPENSEARCH_VARS_WARMDEV=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.opensearch.vars.warmDev="|awk -F'=' '{print $2}')
STORAGE_WARNING=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.storage.warning.default="|awk -F'=' '{print $2}')
STORAGE_CRITICAL=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.storage.critical.default="|awk -F'=' '{print $2}')
CPU_WARNING=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.cpu.warning.default="|awk -F'=' '{print $2}')
CPU_CRITICAL=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.cpu.critical.default="|awk -F'=' '{print $2}')
MEM_WARNING=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.memory.warning.default="|awk -F'=' '{print $2}')
MEM_CRITICAL=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.memory.critical.default="|awk -F'=' '{print $2}')
LOAD1_WARNING=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.load1.warning.default="|awk -F'=' '{print $2}')
LOAD1_CRITICAL=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.load1.critical.default="|awk -F'=' '{print $2}')
LOAD5_WARNING=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.load5.warning.default="|awk -F'=' '{print $2}')
LOAD5_CRITICAL=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.load5.critical.default="|awk -F'=' '{print $2}')
LOAD15_WARNING=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.load15.warning.default="|awk -F'=' '{print $2}')
LOAD15_CRITICAL=$(cat ${CONFIG_FILE}|grep -v "^#"|grep "monitor.threshold.load15.critical.default="|awk -F'=' '{print $2}')


#### init AnyRobot_Monitor_Summary_Dashboard ####

sed -i "s@#OPENSEARCH_VARS_ROOT#@${OPENSEARCH_VARS_ROOT}@g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s@#OPENSEARCH_VARS_HOT#@${OPENSEARCH_VARS_HOT}@g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s@#OPENSEARCH_VARS_WARM#@${OPENSEARCH_VARS_WARM}@g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#OPENSEARCH_VARS_ROOTDEV#/${OPENSEARCH_VARS_ROOTDEV}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#OPENSEARCH_VARS_HOTDEV#/${OPENSEARCH_VARS_HOTDEV}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#OPENSEARCH_VARS_WARMDEV#/${OPENSEARCH_VARS_WARMDEV}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#STORAGE_CRITICAL#/${STORAGE_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#STORAGE_WARNING#/${STORAGE_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#CPU_WARNING#/${CPU_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#CPU_CRITICAL#/${CPU_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#MEM_WARNING#/${MEM_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#MEM_CRITICAL#/${MEM_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#LOAD1_WARNING#/${LOAD1_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#LOAD1_CRITICAL#/${LOAD1_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#LOAD5_WARNING#/${LOAD5_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#LOAD5_CRITICAL#/${LOAD5_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#LOAD15_WARNING#/${LOAD15_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json
sed -i "s/#LOAD15_CRITICAL#/${LOAD15_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/A-01_巡检仪表盘.json

#### init ES_Storage_Monitor_Details dashboard ####

sed -i "s@#OPENSEARCH_VARS_ROOT#@${OPENSEARCH_VARS_ROOT}@g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s@#OPENSEARCH_VARS_HOT#@${OPENSEARCH_VARS_HOT}@g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s@#OPENSEARCH_VARS_WARM#@${OPENSEARCH_VARS_WARM}@g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s/#OPENSEARCH_VARS_ROOTDEV#/${OPENSEARCH_VARS_ROOTDEV}/g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s/#OPENSEARCH_VARS_HOTDEV#/${OPENSEARCH_VARS_HOTDEV}/g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s/#OPENSEARCH_VARS_WARMDEV#/${OPENSEARCH_VARS_WARMDEV}/g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s/#STORAGE_CRITICAL#/${STORAGE_CRITICAL}/g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json
sed -i "s/#STORAGE_WARNING#/${STORAGE_WARNING}/g" ${DASHBOARDS_DIR}/anyrobot/C-01_OpenSearch存储监控.json


exec /run.sh $@
