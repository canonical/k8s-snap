package csrsigning

import (
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	certv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func validateCSR(obj *certv1.CertificateSigningRequest) error {
	csr, err := pkiutil.LoadCertificateRequest(string(obj.Spec.Request))
	if err != nil {
		return fmt.Errorf("failed to parse x509 certificate request: %w", err)
	}

	_ = csr

	// TODO(neoaggelos): validate requests have been encrypted using the k8sd public key:
	// encryptedSignature = obj.Annotations["k8sd.io/signature"]
	// signature = RSA_DECRYPT(k8sdPrivateKey, encryptedSignature)
	// hash = SHA256(obj.Spec.Request)
	// assert hash == signature

	// COMMON ASSERTIONS
	hostname := obj.GetAnnotations()["k8sd.io/node"]
	if len(hostname) == 0 {
		return fmt.Errorf("k8sd.io/node annotation missing from CSR object")
	}
	if clean, err := utils.CleanHostname(hostname); err != nil {
		return fmt.Errorf("CSR has invalid node name %q: %w", hostname, err)
	} else if clean != hostname {
		return fmt.Errorf("CSR has invalid node name %q, should be %q", hostname, clean)
	}
	if obj.Spec.Username != fmt.Sprintf("system:node:%s", hostname) {
		return fmt.Errorf("CSR requestor must be system:node:%s", hostname)
	}
	// NOTE(neoaggelos): .spec.groups might contain more groups, e.g. `system:authenticated`
	if !sets.New(obj.Spec.Groups...).Has("system:nodes") {
		return fmt.Errorf("CSR missing required group system:nodes")
	}

	switch obj.Spec.SignerName {
	case "k8sd.io/kubelet-serving":
		expectUsages := sets.New(certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment)
		if !sets.New(obj.Spec.Usages...).Equal(expectUsages) {
			return fmt.Errorf("CSR usages %v must match %v", obj.Spec.Usages, expectUsages)
		}
		if csr.Subject.CommonName != obj.Spec.Username {
			return fmt.Errorf("CSR commonName %v must match %v", csr.Subject.CommonName, obj.Spec.Username)
		}
		if !sets.New(csr.Subject.Organization...).Equal(sets.New("system:nodes")) {
			return fmt.Errorf("CSR organization %v must match %v", csr.Subject.Organization, []string{"system:nodes"})
		}
		// csr.DNSNames == [...]
		// csr.IPAddresses == [...]
	case "k8sd.io/kubelet-client":
		expectUsages := sets.New(certv1.UsageClientAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment)
		if !sets.New(obj.Spec.Usages...).Equal(expectUsages) {
			return fmt.Errorf("CSR usages %v must match %v", obj.Spec.Usages, expectUsages)
		}
		if csr.Subject.CommonName != obj.Spec.Username {
			return fmt.Errorf("CSR commonName %v must match %v", csr.Subject.CommonName, obj.Spec.Username)
		}
		if !sets.New(csr.Subject.Organization...).Equal(sets.New("system:nodes")) {
			return fmt.Errorf("CSR organization %v must match %v", csr.Subject.Organization, []string{"system:nodes"})
		}
	case "k8sd.io/kube-proxy-client":
		expectUsages := sets.New(certv1.UsageClientAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment)
		if !sets.New(obj.Spec.Usages...).Equal(expectUsages) {
			return fmt.Errorf("CSR usages %v must match %v", obj.Spec.Usages, expectUsages)
		}
		if csr.Subject.CommonName != "system:kube-proxy" {
			return fmt.Errorf("CSR commonName %v must match %v", csr.Subject.CommonName, "system:kube-proxy")
		}
		if len(csr.Subject.Organization) > 0 {
			return fmt.Errorf("CSR organization %v must be empty", csr.Subject.Organization)
		}
	default:
		return fmt.Errorf("CSR has unknown signerName=%v", obj.Spec.SignerName)
	}
	return nil
}
