#!/bin/bash
# tips:
# 1,protoc:下载linux版本的protoc编译器,拷贝到../bin目录下
# 2,linux下go get github.com/gogo/protobuf/protoc-gen-gofast,拷贝protoc-gen-gofast插件到本地(windows)GOPATH/bin目录下
# 3,本地(windows)下执行改脚本：Git Bash 客户端执行该脚本
if [[ -z `command -v protoc` ]];then
  alias protoc=../bin/protoc
fi

export PATH=$PWD/../bin/:$PATH

export GO_OUT_DIR='../proto'

for file in `ls *.proto`
do
    if test -f ${file}
    then
        protoc -I=. --gofast_out=plugins=grpc:${GO_OUT_DIR} ${file}
    fi
done




# python -m grpc_tools.protoc -I=C:\\Code\\GO\\videoDynamicSpider\\webSiteGRPC --python_out=../bilibili_grpc_server --pyi_out=../bilibili_grpc_server --grpc_python_out=../bilibili_grpc_server C:\\Code\\GO\\videoDynamicSpider\\webSiteGRPC\\server.proto
# protoc --go_out=../proto --go_opt=paths=source_relative  --go-grpc_out=../proto --go-grpc_opt=paths=source_relative  server.proto