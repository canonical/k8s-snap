package csrsigning

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/utils"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	certv1 "k8s.io/api/certificates/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *csrSigningReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("csr", req.Name)

	obj := &certv1.CertificateSigningRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get CSR")
		return ctrl.Result{}, err
	}

	log = log.WithValues("signerName", obj.Spec.SignerName)

	// skip CSRs that already have a signed certificate.
	if len(obj.Status.Certificate) > 0 {
		log.V(1).Info("CSR already has a signed certificate")
		return ctrl.Result{}, nil
	}

	// skip CSRs with an unknown signerName.
	if _, ok := r.managedSignerNames[obj.Spec.SignerName]; !ok {
		return ctrl.Result{}, nil
	}

	var approved bool
	for _, condition := range obj.Status.Conditions {
		switch condition.Type {
		case certv1.CertificateDenied:
			log.WithValues("condition", condition).Info("CSR is denied")
			return ctrl.Result{}, nil
		case certv1.CertificateFailed:
			log.WithValues("condition", condition).Info("CSR is failed")
			return ctrl.Result{}, nil
		case certv1.CertificateApproved:
			approved = true
		}
	}

	config, err := r.getClusterConfig(ctx)
	if err != nil {
		log.Error(err, "Failed to retrieve k8sd cluster configuration")
		return ctrl.Result{}, err
	}
	internal := internalConfigFromAnnotations(config.Annotations)

	if !approved {
		log.Info("CSR is not approved")
		if internal.autoApprove {
			log.V(1).Info("CSR auto-approval is enabled")
			keyPEM := config.Certificates.GetK8sdPrivateKey()

			if keyPEM == "" {
				return ctrl.Result{}, fmt.Errorf("cluster RSA key not set")
			}

			priv, err := pkiutil.LoadRSAPrivateKey(keyPEM)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to load cluster RSA key: %w", err)
			}
			return r.reconcileAutoApprove(ctx, log, obj, priv, r.Client)
		}

		log.Info("Requeue while waiting for CSR to be approved")
		return ctrl.Result{RequeueAfter: requeueAfterWaitingForApproved}, nil
	}
	log.Info("CSR is approved")

	certRequest, err := pkiutil.LoadCertificateRequest(string(obj.Spec.Request))
	if err != nil {
		log.Error(err, "Failed to parse CSR from object")
		return ctrl.Result{}, err
	}

	serialNumber, err := pkiutil.GenerateSerialNumber()
	if err != nil {
		log.Error(err, "Failed to generate certificate serial number")
		return ctrl.Result{}, err
	}

	notBefore := time.Now()
	var notAfter time.Time

	if obj.Spec.ExpirationSeconds != nil {
		notAfter = utils.SecondsToExpirationDate(notBefore, int(*obj.Spec.ExpirationSeconds))
	} else {
		notAfter = time.Now().AddDate(10, 0, 0)
	}

	var crtPEM []byte
	switch obj.Spec.SignerName {
	case "k8sd.io/kubelet-serving":
		caCert, caKey, err := pkiutil.LoadCertificate(config.Certificates.GetCACert(), config.Certificates.GetCAKey())
		if err != nil {
			log.Error(err, "Failed to load CA certificate and key")
			return ctrl.Result{}, err
		}
		if caKey == nil {
			log.Error(err, "Cannot sign certificate as CA private key is not available")
			return ctrl.Result{}, nil
		}
		cert := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName:   obj.Spec.Username,
				Organization: obj.Spec.Groups,
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			IPAddresses:           certRequest.IPAddresses,
			DNSNames:              certRequest.DNSNames,
			BasicConstraintsValid: true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, certRequest.PublicKey, caKey)
		if err != nil {
			log.Error(err, "Failed to sign certificate")
			return ctrl.Result{}, err
		}
		crtPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		if crtPEM == nil {
			log.Info("Failed to encode signed certificate to memory")
			return ctrl.Result{RequeueAfter: requeueAfterSigningFailure}, nil
		}
	case "k8sd.io/kubelet-client":
		caCert, caKey, err := pkiutil.LoadCertificate(config.Certificates.GetClientCACert(), config.Certificates.GetClientCAKey())
		if err != nil {
			log.Error(err, "Failed to load client CA certificate and key")
			return ctrl.Result{}, err
		}
		if caKey == nil {
			log.Error(err, "Cannot sign certificate as client CA private key is not available")
			return ctrl.Result{}, nil
		}
		cert := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName:   obj.Spec.Username,
				Organization: obj.Spec.Groups,
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			BasicConstraintsValid: true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, certRequest.PublicKey, caKey)
		if err != nil {
			log.Error(err, "Failed to sign certificate")
			return ctrl.Result{}, err
		}
		crtPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		if crtPEM == nil {
			log.Info("Failed to encode signed certificate to memory")
			return ctrl.Result{RequeueAfter: requeueAfterSigningFailure}, nil
		}
	case "k8sd.io/kube-proxy-client":
		caCert, caKey, err := pkiutil.LoadCertificate(config.Certificates.GetClientCACert(), config.Certificates.GetClientCAKey())
		if err != nil {
			log.Error(err, "Failed to load client CA certificate and key")
			return ctrl.Result{}, err
		}
		if caKey == nil {
			log.Error(err, "Cannot sign certificate as client CA private key is not available")
			return ctrl.Result{}, nil
		}
		cert := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName: "system:kube-proxy",
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			BasicConstraintsValid: true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, certRequest.PublicKey, caKey)
		if err != nil {
			log.Error(err, "Failed to sign certificate")
			return ctrl.Result{}, err
		}
		crtPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		if crtPEM == nil {
			log.Info("Failed to encode signed certificate to memory")
			return ctrl.Result{RequeueAfter: requeueAfterSigningFailure}, nil
		}
	default:
		// NOTE(neoaggelos): this should never happen
		log.V(1).Info("Ignoring CSR with invalid signerName")
		return ctrl.Result{}, nil
	}

	obj.Status.Certificate = crtPEM
	if err := r.Client.Status().Update(ctx, obj); err != nil {
		log.Error(err, "Failed to update CSR with signed certificate")
		return ctrl.Result{}, err
	}

	log.Info("CSR signed")
	return ctrl.Result{}, nil
}
