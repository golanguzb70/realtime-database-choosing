FROM redis/redis-stack:latest

ARG DEBIAN_FRONTEND=noninteractive

RUN apt update &&\
    apt install -yqq curl dnsutils net-tools iputils-ping mc procps screen telnet tshark wget

ENTRYPOINT ["/entrypoint.sh"]