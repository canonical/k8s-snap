package types

type ClusterAPI struct {
	AuthToken *string `json:"auth-token,omitempty"`
}

func (c ClusterAPI) GetAuthToken() string { return getField(c.AuthToken) }
func (c ClusterAPI) Empty() bool          { return c == ClusterAPI{} }
