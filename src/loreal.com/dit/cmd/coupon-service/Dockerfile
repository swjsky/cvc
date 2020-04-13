FROM centos:7

LABEL MAINTAINER="yhl10000@gmail.com" 

RUN yum -y update && yum clean all

RUN mkdir -p /go && chmod -R 777 /go && \
    yum install -y centos-release-scl && \
    yum -y install git go-toolset-1.12 && yum clean all

ENV GOPATH=/go \
    BASH_ENV=/opt/rh/go-toolset-1.12/enable \
    ENV=/opt/rh/go-toolset-1.12/enable
    # PROMPT_COMMAND=". /opt/rh/go-toolset-1.12/enable"

WORKDIR /go