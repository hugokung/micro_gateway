#!/bin/sh

export GO111MODULE=on
export GOPROXY=https://goproxy.io
ps aux | grep micro_gateway | grep -v 'grep' | awk '{print $2}' | xargs kill

nohup ./bin/micro_gateway -conf=./conf/dev/ -endpoint=dashboard >> logs/dashboard.log 2>&1 &

echo 'nohup ./bin/micro_gateway -conf=./conf/dev/ -endpoint=dashboard >> logs/dashboard.log 2>&1 &'

nohup ./bin/micro_gateway -conf=./conf/dev/ -endpoint=server >> logs/server.log 2>&1 &

echo 'nohup ./bin/micro_gateway -conf=./conf/dev/ -endpoint=server >> logs/server.log 2>&1 &'
