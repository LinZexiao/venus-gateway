package config

import (
	"io/ioutil"

	"github.com/ipfs-force-community/metrics"
	"github.com/pelletier/go-toml"
)

const (
	// Configuration file name
	ConfigFile = "config.toml"
)

type Config struct {
	API             *APIConfig
	Auth            *AuthConfig
	Trace           *metrics.TraceConfig
	RateLimit       *RateLimitCofnig
	LocalAuthConfig *LocalAuthConfig

	EnableVeirfyAddress bool
}

type APIConfig struct {
	ListenAddress string
}

type AuthConfig struct {
	URL string
}

type RateLimitCofnig struct {
	Redis string
}

type LocalAuthConfig struct {
	TokenFile string
}

func DefaultConfig() *Config {
	cfg := &Config{
		API:             &APIConfig{ListenAddress: "/ip4/127.0.0.1/tcp/45132"},
		Auth:            &AuthConfig{URL: "http://127.0.0.1:8989"},
		Trace:           metrics.DefaultTraceConfig(),
		RateLimit:       &RateLimitCofnig{Redis: ""},
		LocalAuthConfig: &LocalAuthConfig{TokenFile: "token"},
	}
	cfg.Trace.ServerName = "venus-gateway"
	cfg.Trace.JaegerEndpoint = ""
	cfg.EnableVeirfyAddress = true

	return cfg
}

func ReadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = toml.Unmarshal(data, cfg)

	return cfg, err
}

func WriteConfig(filePath string, cfg *Config) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}