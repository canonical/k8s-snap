package v1

// CreateTokenRequest is the request for "POST 1.0/k8sd/tokens".
type CreateTokenRequest struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

// CreateTokenResponse is the response for "POST 1.0/k8sd/tokens".
type CreateTokenResponse struct {
	Token string `json:"token"`
	Error string `json:"error,omitempty"`
}

// CheckTokenResponse is the response for "GET 1.0/k8sd/tokens".
type CheckTokenResponse struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
	Error    string   `json:"error,omitempty"`
}

// TokenReviewRequest is the request for "POST 1.0/k8sd/tokens/webhook/v1".
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
