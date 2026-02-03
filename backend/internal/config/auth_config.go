package config

type AuthConfig struct {
	Secret    string
	TokenTTL  int
	IssuerURL string
}
