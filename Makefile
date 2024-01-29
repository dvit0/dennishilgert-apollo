################################################################################
# Variables                                                                    #
################################################################################

# Variables needed for Proto Initialization and Generation
PROTOC ?= protoc
PROTOC_GEN_GO_VERSION = v1.32.0
PROTOC_GEN_GO_GRPC_VERSION = 1.3.0
PROTO_PREFIX := github.com/dennishilgert/apollo
GRPC_PROTOS := $(shell ls apollo/proto)

################################################################################
# Target: init-proto                                                           #
################################################################################
.PHONY: init-proto
init-proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)

################################################################################
# Target: gen-proto                                                            #
################################################################################
# Generate proto files for each service
define genProtoc
.PHONY: gen-proto-$(1)
gen-proto-$(1):
	$(PROTOC) --go_out=. --go_opt=module=$(PROTO_PREFIX) --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,module=$(PROTO_PREFIX) ./apollo/proto/$(1)/v1/*.proto
endef

$(foreach ITEM,$(GRPC_PROTOS),$(eval $(call genProtoc,$(ITEM))))

GEN_PROTOS:=$(foreach ITEM,$(GRPC_PROTOS),gen-proto-$(ITEM))

.PHONY: gen-proto
gen-proto: $(GEN_PROTOS)