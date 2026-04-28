ARG BASE_IMAGE=acr.aishu.cn/public/grafana:10.4.19-security-01-ubuntu

FROM ${BASE_IMAGE} AS prod

ADD dashboards.yaml /etc/grafana/provisioning/dashboards/dashboards.yaml
ADD dashboards /etc/grafana/provisioning-dashboards/

ADD plugins/grafana-anyrobot-datasource.tar.gz /etc/grafana/provisioning-plugins/
ADD plugins/grafana-anyrobot-metric-model-datasource.tar.gz /etc/grafana/provisioning-plugins/
ADD plugins/grafana-polystat-panel.any.tar.gz /etc/grafana/provisioning-plugins/

ENV GF_PATHS_PLUGINS=/etc/grafana/provisioning-plugins/

USER root
RUN set -ex; \
    chmod 777 /etc/grafana/provisioning-dashboards/ -R
USER grafana

COPY entrypoint.sh /etc/grafana/entrypoint.sh

ENTRYPOINT ["/etc/grafana/entrypoint.sh"]
