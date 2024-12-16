default: help

COREDNS_FILE_NAME := coredns-1.36.0
CK_LOADBALANCER_FILE_NAME = ck-loadbalancer
CILIUM_FILE_NAME := cilium-1.16.3
METALLB_FILE_NAME := metallb-0.14.8
RAWFILE_CSI_FILE_NAME := rawfile-csi-0.9.0
METRICS_SERVER_FILE_NAME := metrics-server-3.12.2
TARGET_DIR := ../../../src/k8s/pkg/k8sd/features/values

.PHONY: help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

.PHONY: extract-values
extract-values: ## Extract values.yaml files from Helm charts.
	@echo "Extracting CoreDNS values..."
	@tar --transform='s|coredns/values.yaml|${COREDNS_FILE_NAME}_values.yaml|' -zxf ${COREDNS_FILE_NAME}.tgz coredns/values.yaml

	@echo "Extracting Cilium values..."
	@tar --transform='s|cilium/values.yaml|${CILIUM_FILE_NAME}_values.yaml|' -zxf ${CILIUM_FILE_NAME}.tgz cilium/values.yaml

	@echo "Copying CK Loadbalancer values..."
	@cp ck-loadbalancer/values.yaml ${CK_LOADBALANCER_FILE_NAME}_values.yaml

	@echo "Extracting MetalLB values..."
	@tar --transform='s|metallb/values.yaml|${METALLB_FILE_NAME}_values.yaml|' -zxf ${METALLB_FILE_NAME}.tgz metallb/values.yaml

	@echo "Extracting Rawfile CSI values..."
	@tar --transform='s|rawfile-csi/values.yaml|${RAWFILE_CSI_FILE_NAME}_values.yaml|' -zxf ${RAWFILE_CSI_FILE_NAME}.tgz rawfile-csi/values.yaml

	@echo "Extracting Metrics Server values..."
	@tar --transform='s|metrics-server/values.yaml|${METRICS_SERVER_FILE_NAME}_values.yaml|' -zxf ${METRICS_SERVER_FILE_NAME}.tgz metrics-server/values.yaml

.PHONY: gen
gen: extract-values ## Generate Go code from values.yaml files.
	@go run generator.go -files=${COREDNS_FILE_NAME}_values.yaml,${CILIUM_FILE_NAME}_values.yaml,${CK_LOADBALANCER_FILE_NAME}_values.yaml,${METALLB_FILE_NAME}_values.yaml,${RAWFILE_CSI_FILE_NAME}_values.yaml,${METRICS_SERVER_FILE_NAME}_values.yaml -pkg=values -out-dir=${TARGET_DIR} -advanced-types=true -unsafe-field=true

.PHONY: clean
clean: clean-yaml-values clean-gen-values ## Clean up.

.PHONY: clean-yaml-values
clean-yaml-values: ## Clean up extracted values.yaml files.
	rm -f ${COREDNS_FILE_NAME}_values.yaml
	rm -f ${CILIUM_FILE_NAME}_values.yaml
	rm -f ${CK_LOADBALANCER_FILE_NAME}_values.yaml
	rm -f ${METALLB_FILE_NAME}_values.yaml
	rm -f ${RAWFILE_CSI_FILE_NAME}_values.yaml
	rm -f ${METRICS_SERVER_FILE_NAME}_values.yaml

.PHONY: clean-gen-values
clean-gen-values: ## Clean up generated Go code.
	rm -f ${TARGET_DIR}/${COREDNS_FILE_NAME}_values.go
	rm -f ${TARGET_DIR}/${CILIUM_FILE_NAME}_values.go
	rm -f ${TARGET_DIR}/${CK_LOADBALANCER_FILE_NAME}_values.go
	rm -f ${TARGET_DIR}/${METALLB_FILE_NAME}_values.go
	rm -f ${TARGET_DIR}/${RAWFILE_CSI_FILE_NAME}_values.go
	rm -f ${TARGET_DIR}/${METRICS_SERVER_FILE_NAME}_values.go
