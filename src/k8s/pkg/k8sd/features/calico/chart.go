package calico

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartCalico represents manifests to deploy Calico.
	chartCalico = helm.InstallableChart{
		Name:         "ck-network",
		Namespace:    "tigera-operator",
		ManifestPath: path.Join("charts", "tigera-operator-v3.28.0.tgz"),
	}

	// Note: Tigera is the company behind Calico and the tigera-operator is the operator for Calico.
	// TODO: use ROCKs instead of upstream
	// imageRepo represents the repo to fetch the Calico CNI images.
	imageRepo = "ghcr.io/canonical/k8s-snap"

	// calicoImageRepo represents the repo to fetch the calico images.
	calicoImageRepo = "ghcr.io/canonical/k8s-snap/calico"
	// calicoTag represents the tag to use for the calico images.
	calicoTag = "v3.28.0"

	// tigeraOperatorImage represents the image to fetch for calico.
	tigeraOperatorImage = "tigera/operator"

	// tigeraOperatorVersion is the version to use for the tigera-operator image.
	tigeraOperatorVersion = "v1.34.0"

	// calicoCtlImage represents the image to fetch for calicoctl.
	// TODO: use ROCKs instead of upstream
	calicoCtlImage = "ghcr.io/canonical/k8s-snap/calico/ctl"
	// calicoCtlTag represents the tag to use for the calicoctl image.
	calicoCtlTag = "v3.28.0"
)
