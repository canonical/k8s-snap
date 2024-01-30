# Updating Helm Charts

## Overview
This section provides a guide on how to update the Helm charts for k8s components. Currently, the
process is manual, but there are plans to automate it in the near future.

## Steps for Updating

### 1. Pulling the New Charts
To pull the new charts, follow these steps:

- **Add the Helm Repository**: For the component you need to update (e.g., Cilium, CoreDNS), add the relevant Helm repository. 

- **Pull the Chart**: Use the following command to pull the chart:
    ```bash
    helm pull [chart-name] --destination ./k8s/components/charts
    ```
    Replace `[chart-name]` with the name of your chart. This command saves the chart as a `.tgz` file in the `charts` folder.

### 2. Updating the Component Matrix in the File
You need to update the component entry in the `components.yaml` file. Follow these instructions:

- **Locate the Component File**: Open the `components.yaml` file located at `k8s/components/components.yaml`.

- **Edit the Chart Name**: Find the entry for the component you are updating (e.g., `network`). For example, if updating Cilium to version `1.14.2`, modify as follows:
    ```diff
    network:
      release: "ck-network"
    -  chart: "cilium-1.14.1.tgz"
    +  chart: "cilium-1.14.2.tgz"
      namespace: "kube-system"
    dns:
      release: "ck-dns"
      chart: "coredns-1.28.2.tgz"
      namespace: "kube-system"
    ```
