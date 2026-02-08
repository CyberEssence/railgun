package config

type JWTConfig struct {
	EncryptionKey string `yaml:"encryption_key"`
	Issuer        string `yaml:"issuer" default:"Railgun SIEM"`
	TOTPIssuer    string `yaml:"totp_issuer" default:"Railgun SIEM"`
	TOTPPeriod    int    `yaml:"totp_period" default:"30"`
	TOTPDigits    int    `yaml:"totp_digits" default:"6"`
}
