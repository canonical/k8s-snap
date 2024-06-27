package apiv1

// GenerateKubernetesAuthTokenRequest is used to request a new Kubernetes auth token.
type GenerateKubernetesAuthTokenRequest struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

// CreateKubernetesAuthTokenResponse is used to return the Kubernetes auth token.
type CreateKubernetesAuthTokenResponse struct {
	Token string `json:"token"`
}

// CreateKubernetesAuthTokenRequest is the request for "DELETE 1.0/kubernetes/auth/tokens".
type RevokeKubernetesAuthTokenRequest struct {
	Token string `json:"token"`
}

// CheckKubernetesAuthTokenResponse is the response for "GET 1.0/kubernetes/auth/tokens".
type CheckKubernetesAuthTokenResponse struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

// TokenReviewRequest is the request for "POST 1.0/kubernetes/auth/webhook".
// This mirrors the definition of the Kubernetes API group="authentication.k8s.io/v1" kind="TokenReview"
// https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/token-review-v1/
type TokenReview struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Spec       TokenReviewSpec   `json:"spec"`
	Status     TokenReviewStatus `json:"status"`
}

// TokenReviewSpec is set by kube-apiserver in TokenReview.
// This mirrors the definition of the Kubernetes API group="authentication.k8s.io/v1" kind="TokenReview"
// https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/token-review-v1/#TokenReviewSpec
type TokenReviewSpec struct {
	Audiences []string `json:"audiences,omitempty"`
	Token     string   `json:"token"`
}

// TokenReviewStatus is set by the webhook server in TokenReview.
// This mirrors the definition of the Kubernetes API group="authentication.k8s.io/v1" kind="TokenReview"
// https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/token-review-v1/#TokenReviewStatus
type TokenReviewStatus struct {
	Audiences     []string                  `json:"audiences,omitempty"`
	Authenticated bool                      `json:"authenticated"`
	Error         string                    `json:"error,omitempty"`
	User          TokenReviewStatusUserInfo `json:"user,omitempty"`
}

// TokenReviewStatusUserInfo is set by the webhook server in TokenReview.
// This mirrors the definition of the Kubernetes API group="authentication.k8s.io/v1" kind="TokenReview"
// https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/token-review-v1/#TokenReviewStatus
type TokenReviewStatusUserInfo struct {
	Extra    map[string][]string `json:"extra,omitempty"`
	Groups   []string            `json:"groups,omitempty"`
	Username string              `json:"username,omitempty"`
	UID      string              `json:"uid,omitempty"`
}
