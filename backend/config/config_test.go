package config

import "testing"

func validServerConfig() *Config {
	return &Config{
		DBPassword:           "0123456789abcdef0123456789abcdef",
		MinioAccessKey:       "audit-service",
		MinioSecretKey:       "abcdef0123456789abcdef0123456789",
		SettingEncryptionKey: "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		JWTSecret:            "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
}

func TestValidateServerAcceptsStrongConfiguration(t *testing.T) {
	if err := validServerConfig().ValidateServer(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateServerRejectsMissingAndWeakSecrets(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"missing jwt", func(c *Config) { c.JWTSecret = "" }},
		{"short jwt", func(c *Config) { c.JWTSecret = "short" }},
		{"default database password", func(c *Config) { c.DBPassword = "secret" }},
		{"default minio password", func(c *Config) { c.MinioSecretKey = "minioadmin" }},
		{"invalid aes", func(c *Config) { c.SettingEncryptionKey = "not-base64" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validServerConfig()
			tt.mutate(cfg)
			if err := cfg.ValidateServer(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
