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
    helm pull [chart-name] --destination ./charts
    ```
    Replace `[chart-name]` with the name of your chart. This command saves the chart as a `.tgz` file in the `charts` folder.

### 2. Updating the Component Matrix in the Code
You need to update the `componentMap` entry for the component in the code. Follow these instructions:

- **Locate the Component File**: Open the `component.go` file located at `src/k8s/pkg/component`.

- **Edit the Chart Path**: Find the entry for the component you are updating (e.g., `cilium`). If the new chart version is `1.14.2`, update the `ChartPath` to point to the new chart version. For example:
    ```go
    var componentMap = map[string]ChartInfo{
	"cni": {ReleaseName: "ck-cni", ChartPath: path.Join(os.Getenv("SNAP"), "cilium-1.14.2.tgz")},
	"dns": {ReleaseName: "ck-dns", ChartPath: path.Join(os.Getenv("SNAP"), "coredns-1.28.2.tgz")},
    }
    ```
