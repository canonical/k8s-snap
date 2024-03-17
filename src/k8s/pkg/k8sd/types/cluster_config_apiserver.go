package types

type APIServer struct {
	SecurePort        *int    `json:"port,omitempty"`
	AuthorizationMode *string `json:"authorization-mode,omitempty"`
}

func (c APIServer) GetSecurePort() int           { return getField(c.SecurePort) }
func (c APIServer) GetAuthorizationMode() string { return getField(c.AuthorizationMode) }
func (c APIServer) Empty() bool                  { return c.SecurePort == nil && c.AuthorizationMode == nil }
