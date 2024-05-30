################################################################################
# Variables                                                                    #
################################################################################

PROTOC ?= protoc
PROTOC_GEN_GO_VERSION = v1.32.0
PROTOC_GEN_GO_GRPC_VERSION = 1.3.0
PROTO_PREFIX := github.com/dennishilgert/apollo
GRPC_PROTOS := $(shell find apollo/proto -mindepth 1 -maxdepth 1 -type d)

BIN_DIR := ./bin
BUILD_DIR := ./build
DOCKER_DIR := ./docker

GOARCH ?= amd64
GOOS ?= linux
DOCKER_IMAGE_PREFIX ?= apollo
DOCKER_TAG ?= latest

################################################################################
# Proto Targets                                                               #
################################################################################

.PHONY: init-proto
init-proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)

define genProtoc
.PHONY: gen-proto-$(notdir $(1))
gen-proto-$(notdir $(1)):
	$(PROTOC) --go_out=. --go_opt=module=$(PROTO_PREFIX) \
		--go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,module=$(PROTO_PREFIX) \
		$(1)/v1/*.proto
endef

$(foreach ITEM,$(GRPC_PROTOS),$(eval $(call genProtoc,$(ITEM))))

GEN_PROTOS:=$(foreach ITEM,$(GRPC_PROTOS),gen-proto-$(notdir $(ITEM)))

.PHONY: gen-proto
gen-proto: $(GEN_PROTOS)

################################################################################
# Build Targets                                                                #
################################################################################

CMD_DIRS := $(wildcard cmd/*)

define buildCmd
.PHONY: build-$(notdir $(1))
build-$(notdir $(1)):
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(GOOS)_$(GOARCH)/$(notdir $(1)) ./cmd/$(notdir $(1))
endef

$(foreach ITEM,$(CMD_DIRS),$(eval $(call buildCmd,$(ITEM))))

.PHONY: build
build: $(foreach ITEM,$(CMD_DIRS),build-$(notdir $(ITEM)))

################################################################################
# Docker Targets                                                               #
################################################################################

define dockerBuild
.PHONY: docker-$(notdir $(1))
docker-$(notdir $(1)):
	docker build -f $(DOCKER_DIR)/$(notdir $(1))/Dockerfile -t $(DOCKER_IMAGE_PREFIX)/$(notdir $(1)):$(DOCKER_TAG) .
endef

$(foreach ITEM,$(CMD_DIRS),$(eval $(call dockerBuild,$(ITEM))))

.PHONY: docker-build
docker-build: $(foreach ITEM,$(CMD_DIRS),docker-$(notdir $(ITEM)))

################################################################################
# Build resources                                                              #
################################################################################

.PHONY: resources-build
resources-build:
	mkdir -p build
	cp $(BIN_DIR)/$(GOOS)_$(GOARCH)/cli $(BUILD_DIR)/cli
	cp $(BIN_DIR)/$(GOOS)_$(GOARCH)/agent $(BUILD_DIR)/agent
	cp ./init/init $(BUILD_DIR)/init

	cp ./resources/examples/node/index.mjs $(BUILD_DIR)/index.mjs
	cp ./resources/examples/node/package.json $(BUILD_DIR)/package.json

	cp ./resources/package/node/Dockerfile.initial-node $(BUILD_DIR)/Dockerfile.initial-node
	cp ./resources/runtimes/node/Dockerfile.node $(BUILD_DIR)/Dockerfile.node
	cp ./resources/wrappers/node/wrapper.js $(BUILD_DIR)/wrapper.js
	cp ./resources/rootfs/Dockerfile.baseos $(BUILD_DIR)/Dockerfile.baseos

	./scripts/build-base-images.sh $(BUILD_DIR)

################################################################################
# Dependency Management                                                        #
################################################################################

.PHONY: deps
deps:
	go mod tidy
	go mod verify

################################################################################
# Cross-Compilation Targets                                                    #
################################################################################

.PHONY: cross-compile
cross-compile:
	@$(MAKE) GOOS=linux GOARCH=amd64 build
	@$(MAKE) GOOS=linux GOARCH=arm64 build
	@$(MAKE) GOOS=darwin GOARCH=amd64 build
	@$(MAKE) GOOS=darwin GOARCH=arm64 build
	@$(MAKE) GOOS=windows GOARCH=amd64 build

################################################################################
# Clean                                                                        #
################################################################################

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(BUILD_DIR)