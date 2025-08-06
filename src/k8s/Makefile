DQLITE_BUILD_SCRIPTS_DIR ?= $(shell pwd)/hack
TOOLS_DIR ?= $(shell pwd)/tools
GO_SOURCES = $(shell find . -name '*.go')
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.17.2
CRDS_DIR ?= $(shell pwd)/pkg/k8sd/crds
CRDS_OUTPUT_DIR ?= ../../k8s/crds

## Development
go.fmt:
	go mod tidy
	go fmt ./...

go.lint:
ifeq (, $(shell which golangci-lint))
	echo "golangci-lint not found, installing it"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.0
endif
	golangci-lint run

go.vet:
	$(DQLITE_BUILD_SCRIPTS_DIR)/static-go-vet.sh ./...

go.unit:
	$(DQLITE_BUILD_SCRIPTS_DIR)/static-go-test.sh -v ./pkg/... ./cmd/... -coverprofile=coverage.txt --cover

go.doc: bin/static/k8s
	bin/static/k8s generate-docs --output-dir ../../docs/canonicalk8s/_parts/ --project-dir .

### CRDs
.PHONY: crds
crds: generate manifests ## Build all targets.

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="$(CRDS_DIR)/..." output:crd:artifacts:config=$(CRDS_OUTPUT_DIR)

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object paths="$(CRDS_DIR)/..."

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

## Static Builds
static: bin/static/k8s bin/static/k8sd bin/static/k8s-apiserver-proxy

$(DQLITE_BUILD_SCRIPTS_DIR)/.deps/static/lib/libdqlite.a:
	$(DQLITE_BUILD_SCRIPTS_DIR)/static-dqlite.sh

bin/static/k8s: $(DQLITE_BUILD_SCRIPTS_DIR)/.deps/static/lib/libdqlite.a $(GO_SOURCES)
	mkdir -p bin/static
	$(DQLITE_BUILD_SCRIPTS_DIR)/static-go-build.sh -o bin/static/k8s ./cmd/main.go

bin/static/k8s-apiserver-proxy: bin/static/k8s
	cd bin/static && ln -sf k8s k8s-apiserver-proxy

bin/static/k8sd: bin/static/k8s
	cd bin/static && ln -sf k8s k8sd

## Dynamic Builds
dynamic: bin/dynamic/k8s bin/dynamic/k8sd bin/dynamic/k8s-apiserver-proxy

bin/dynamic/lib/libdqlite.so:
	mkdir -p bin/dynamic/lib
	$(DQLITE_BUILD_SCRIPTS_DIR)/dynamic-dqlite.sh
	cp $(DQLITE_BUILD_SCRIPTS_DIR)/.deps/dynamic/lib/*.so* bin/dynamic/lib

bin/dynamic/k8s: bin/dynamic/lib/libdqlite.so
	mkdir -p bin/dynamic
	$(DQLITE_BUILD_SCRIPTS_DIR)/dynamic-go-build.sh -o bin/dynamic/k8s ./cmd/main.go

bin/dynamic/k8s-apiserver-proxy: bin/dynamic/k8s
	cd bin/dynamic && ln -s k8s k8s-apiserver-proxy

bin/dynamic/k8sd: bin/dynamic/k8s
	cd bin/dynamic && ln -s k8s k8sd

## Cleanup
clean:
	rm -rf bin hack/.build hack/.deps
