default: help

SOURCES_FILE_NAME = sources.txt
TMP_SOURCES_FILE_NAME := tmp_${SOURCES_FILE_NAME}
COREDNS_FILE_NAME := coredns-1.36.0
CK_LOADBALANCER_FILE_NAME = ck-loadbalancer
CILIUM_FILE_NAME := cilium-1.16.3
METALLB_FILE_NAME := metallb-0.14.8
RAWFILE_CSI_FILE_NAME := rawfile-csi-0.9.0
METRICS_SERVER_FILE_NAME := metrics-server-3.12.2

.PHONY: help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'



.PHONY: extract-values
extract-values: ## Extract values.yaml files and update sources.txt
	> ${TMP_SOURCES_FILE_NAME}

	@echo "Extracting CoreDNS values..."
	tar --transform='s|coredns/values.yaml|${COREDNS_FILE_NAME}_values.yaml|' -zxvf ${COREDNS_FILE_NAME}.tgz coredns/values.yaml
	@echo ${COREDNS_FILE_NAME}_values.yaml >> ${TMP_SOURCES_FILE_NAME}

	@echo "Extracting Cilium values..."
	tar --transform='s|cilium/values.yaml|${CILIUM_FILE_NAME}_values.yaml|' -zxvf ${CILIUM_FILE_NAME}.tgz cilium/values.yaml
	@echo ${CILIUM_FILE_NAME}_values.yaml >> ${TMP_SOURCES_FILE_NAME}

	@echo "Copying CK Loadbalancer values..."
	cp ck-loadbalancer/values.yaml ${CK_LOADBALANCER_FILE_NAME}_values.yaml
	@echo "${CK_LOADBALANCER_FILE_NAME}_values.yaml" >> ${TMP_SOURCES_FILE_NAME}

	@echo "Extracting MetalLB values..."
	tar --transform='s|metallb/values.yaml|${METALLB_FILE_NAME}_values.yaml|' -zxvf ${METALLB_FILE_NAME}.tgz metallb/values.yaml
	@echo ${METALLB_FILE_NAME}_values.yaml >> ${TMP_SOURCES_FILE_NAME}

	@echo "Extracting Rawfile CSI values..."
	tar --transform='s|rawfile-csi/values.yaml|${RAWFILE_CSI_FILE_NAME}_values.yaml|' -zxvf ${RAWFILE_CSI_FILE_NAME}.tgz rawfile-csi/values.yaml
	@echo ${RAWFILE_CSI_FILE_NAME}_values.yaml >> ${TMP_SOURCES_FILE_NAME}

	@echo "Extracting Metrics Server values..."
	tar --transform='s|metrics-server/values.yaml|${METRICS_SERVER_FILE_NAME}_values.yaml|' -zxvf ${METRICS_SERVER_FILE_NAME}.tgz metrics-server/values.yaml
	@echo ${METRICS_SERVER_FILE_NAME}_values.yaml >> ${TMP_SOURCES_FILE_NAME}
	
	mv ${TMP_SOURCES_FILE_NAME} ${SOURCES_FILE_NAME}

.PHONY: clean
clean: ## Clean up.
	rm -f ${SOURCES_FILE_NAME}
	rm -f ${COREDNS_FILE_NAME}_values.yaml
	rm -f ${CILIUM_FILE_NAME}_values.yaml
	rm -f ${CK_LOADBALANCER_FILE_NAME}_values.yaml
	rm -f ${METALLB_FILE_NAME}_values.yaml
	rm -f ${RAWFILE_CSI_FILE_NAME}_values.yaml
	rm -f ${METRICS_SERVER_FILE_NAME}_values.yaml

.PHONY: gen
gen: ${SOURCES_FILE_NAME}
	go run generator.go

$(SOURCES_FILE_NAME): extract-values
