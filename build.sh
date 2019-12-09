#!/usr/bin/env bash
mkdir -p output/bin output/conf
RUN_NAME="proxy_server"
find conf/ -type f ! -name "*_local.*" | xargs -I{} cp {} output/conf/
export GO15VENDOREXPERIMENT="1"
go build -o output/bin/${RUN_NAME}
