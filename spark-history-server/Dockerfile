FROM ghcr.io/datapunchorg/spark:spark-3.2.1-1643336295

USER root

RUN groupadd -g 185 spark && \
    useradd -u 185 -g 185 spark

USER 185
ENTRYPOINT /opt/spark/sbin/start-history-server.sh
