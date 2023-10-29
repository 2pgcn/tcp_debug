GOHOSTOS:=$(shell go env GOHOSTOS)
BASEPATH=$(shell pwd)
TCP_DEBUG_VERSION=$(shell cat version)
ifeq ($(GOHOSTOS), windows)
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	CONF_PROTO_FILES=$(shell $(Git_Bash) -c "find conf -name *.proto")
    API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	CONF_PROTO_FILES=$(shell find conf -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: conf
# generate conf proto
conf:
	protoc --proto_path=./conf \
		   --go_out=paths=source_relative:./conf \
		   $(CONF_PROTO_FILES)
.PHONY: api
# generate conf proto
api:
	protoc --proto_path=./api \
		   --go_out=paths=source_relative:./api \
		   $(API_PROTO_FILES)

.PHONY: srv
srv:
	go run ./cmd/main.go srv --conf=$(BASEPATH)/conf/
.PHONY: cli
cli:
	go run ./cmd/main.go cli --conf=$(BASEPATH)/conf/ --dail 127.0.0.1:30001 --startNum 1
.PHONY: push
push:
	docker buildx build --platform linux/amd64 -f ./Dockerfile --push  -t registry.cn-shenzhen.aliyuncs.com/pg/tcpdebug:$(TCP_DEBUG_VERSION) ./
	#&& docker push  registry.cn-shenzhen.aliyuncs.com/pg/tcp_debug:$(TCP_DEBUG_VERSION)