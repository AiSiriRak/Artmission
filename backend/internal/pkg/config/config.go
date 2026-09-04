// Package config loads application configuration from an embedded YAML file
// (schema + safe defaults, committed to the repo) overlaid with environment
// variables (secrets and per-environment values, never committed). Locally,
// environment variables are populated from a .env file; in Docker/production
// they come from the real environment.
package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config interface {
	App() App
	Database() Database
	Auth() Auth

	String() string
}

type App struct {
	Address        string   `mapstructure:"address"         validate:"required"`
	BasePath       string   `mapstructure:"base_path"       validate:"required"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	IsProduction   bool     `mapstructure:"is_production"`
}

type Database struct {
	DSN string `mapstructure:"dsn" validate:"required"`
}

type Auth struct {
	JWTSecret           string        `mapstructure:"jwt_secret"            validate:"required"`
	AccessTokenTTL      time.Duration `mapstructure:"access_token_ttl"      validate:"required"`
	RefreshTokenTTL     time.Duration `mapstructure:"refresh_token_ttl"     validate:"required"`
	RefreshCookieDomain string        `mapstructure:"refresh_cookie_domain"`
}

// -------------------------------------------------------------------------- //

type config struct {
	AppCfg      App      `mapstructure:"app"`
	DatabaseCfg Database `mapstructure:"database"`
	AuthCfg     Auth     `mapstructure:"auth"`
}

func (c *config) App() App           { return c.AppCfg }
func (c *config) Database() Database { return c.DatabaseCfg }
func (c *config) Auth() Auth         { return c.AuthCfg }

func (c *config) String() string {
	redacted := *c
	redacted.DatabaseCfg.DSN = "***"
	redacted.AuthCfg.JWTSecret = "***"

	jsonBytes, err := json.MarshalIndent(&redacted, "", "  ")
	if err != nil {
		return fmt.Sprintf("<failed to marshal config: %v>", err)
	}
	return string(jsonBytes)
}

//go:embed config.yaml
var defaultConfigYAML []byte

// InitConfig loads the embedded YAML schema/defaults, then overlays
// environment variables (loaded from envFile when non-empty). Nested keys
// such as `app.address` are overridden by `APP_ADDRESS`.
func InitConfig(envFile string) (Config, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %q: %w", envFile, err)
		}
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadConfig(bytes.NewReader(defaultConfigYAML)); err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	var cfg config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validator.New().Struct(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}
