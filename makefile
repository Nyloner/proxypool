build:
	sh ./build.sh
run:
	sh build.sh && ./output/bin/proxy_server
mod:
	export GO111MODULE=on && go mod vendor