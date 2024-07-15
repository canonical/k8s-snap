package csrsigning

import (
	"fmt"

	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	certv1 "k8s.io/api/certificates/v1"
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
	// obj.annotations[k8sd.io/node] == $hostname
	// obj.spec.username == system:node:$hostname
	// obj.spec.groups == [system:nodes]

	switch obj.Spec.SignerName {
	case "k8sd.io/kubelet-serving":
		// obj.spec.keyUsages == [certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment]
		// csr.Subject.CommonName == system:node:$hostname
		// csr.Subject.Organization == [system:nodes]
		// csr.DNSNames == [...]
		// csr.IPAddresses == [...]
	case "k8sd.io/kubelet-client":
		// obj.spec.keyUsages == [certv1.UsageClientAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment]
		// csr.Subject.CommonName == system:node:$hostname
		// csr.Subject.Organization == [system:nodes]
	case "k8sd.io/kube-proxy-client":
		// obj.spec.keyUsages == [certv1.UsageClientAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment]
		// csr.Subject.CommonName == system:kube-proxy
		// csr.Subject.Organization == []
	default:
		return fmt.Errorf("CSR has unknown signerName=%v", obj.Spec.SignerName)
	}
	return nil
}
