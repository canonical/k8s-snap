package mock

import (
	"context"
	"testing"

	certv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type getArgs struct {
	name types.NamespacedName
	obj  client.Object
	opts []client.GetOption
}

type getRet struct {
	csr certv1.CertificateSigningRequest
	err error
}

type K8sMock struct {
	client.Client

	t    *testing.T
	srcm *subResourceClientMock

	getCalledWith getArgs
	getReturns    getRet
}

func New(
	t *testing.T,
	srcm *subResourceClientMock,
	getCSR certv1.CertificateSigningRequest,
	getErr error,
) *K8sMock {
	return &K8sMock{
		t:    t,
		srcm: srcm,
		getReturns: getRet{
			csr: getCSR,
			err: getErr,
		},
	}
}

func (m *K8sMock) Get(ctx context.Context, name types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	m.getCalledWith = getArgs{name, obj, opts}
	csr, ok := obj.(*certv1.CertificateSigningRequest)
	if !ok {
		m.t.Fatalf("unexpected object type: %T", obj)
	}
	*csr = m.getReturns.csr
	return m.getReturns.err
}

func (m *K8sMock) Status() client.SubResourceWriter {
	return m.srcm
}

func (m *K8sMock) SubResource(subResource string) client.SubResourceClient {
	switch subResource {
	case "approval":
		return m.srcm
	default:
		m.t.Fatalf("unexpected subResource: %s", subResource)
	}
	return nil
}

func (m *K8sMock) AssertUpdateCalled(t *testing.T) {
	m.srcm.assertUpdateCalled(t)
}

type updateArgs struct {
	obj  client.Object
	opts []client.SubResourceUpdateOption
}

type subResourceClientMock struct {
	client.SubResourceClient

	updateCalledWith []updateArgs
	updateErr        error
}

func NewSubResourceClientMock(
	updateErr error,
) *subResourceClientMock {
	return &subResourceClientMock{
		updateErr: updateErr,
	}
}

func (src *subResourceClientMock) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	src.updateCalledWith = append(src.updateCalledWith, updateArgs{obj, opts})
	return src.updateErr
}

func (src *subResourceClientMock) assertUpdateCalled(t *testing.T) {
	if len(src.updateCalledWith) == 0 {
		t.Error("expected update to have been called")
	}
}
