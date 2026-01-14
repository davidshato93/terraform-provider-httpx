package config

// ProviderConfig holds the provider configuration
type ProviderConfig struct {
	DefaultHeaders       map[string]string
	BasicAuth            *BasicAuthModel
	BearerToken          *string
	TimeoutMs            int64
	InsecureSkipVerify   bool
	ProxyUrl             *string
	CaCertPem            *string
	ClientCertPem        *string
	ClientKeyPem         *string
	RedactHeaders        []string
	MaxResponseBodyBytes int64
	Debug                bool
}

// BasicAuthModel represents basic auth credentials
type BasicAuthModel struct {
	Username string
	Password string
}

