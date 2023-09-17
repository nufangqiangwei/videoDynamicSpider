FROM ubuntu:20.04


RUN rm -rf /etc/apt/sources.list
RUN touch /etc/apt/sources.list


RUN echo 'deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse'>/etc/apt/sources.list
RUN echo 'deb-src http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb-src http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb-src http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb-src http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse'>>/etc/apt/sources.list
RUN echo 'deb-src http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse'>>/etc/apt/sources.list


RUN apt-get update --fix-missing
RUN apt-get install -y wget git gcc


ENV GO_VERSION 1.20
RUN wget -P /tmp "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz"

RUN tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
RUN rm "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR $GOPATH
