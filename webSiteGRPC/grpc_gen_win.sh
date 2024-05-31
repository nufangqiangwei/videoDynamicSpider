#!/bin/bash

python -m grpc_tools.protoc -I=E:\\GoCode\\videoDynamicAcquisition\\webSiteGRPC --python_out=../bilibili_grpc_server --pyi_out=../bilibili_grpc_server --grpc_python_out=../bilibili_grpc_server E:\\GoCode\\videoDynamicAcquisition\\webSiteGRPC\\server.proto
protoc --go_out=../proto --go_opt=paths=source_relative  --go-grpc_out=../proto --go-grpc_opt=paths=source_relative  server.proto
