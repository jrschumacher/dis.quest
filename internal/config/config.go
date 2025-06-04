package config

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/spf13/viper"
)

const (
	EnvProd = "production"
	EnvDev  = "development"
	EnvTest = "test"
)

// Config holds application configuration loaded from environment variables or config file.
type Config struct {
	AppEnv      string `mapstructure:"app_env" default:"development" validate:"required"`
	Port        string `mapstructure:"port" default:"3000" validate:"required"`
	PDSEndpoint string `mapstructure:"pds_endpoint" default:"http://localhost:4000"`

	// Security settings
	DatabaseURL  string `secret:"true" mapstructure:"database_url"`
	JWKSPrivate  string `validate:"required" secret:"true" mapstructure:"jwks_private" validate:"required"`
	JWKSPublic   string `mapstructure:"jwks_public" validate:"required"`
	PublicDomain string `mapstructure:"public_domain" validate:"required"`
	AppName      string `mapstructure:"app_name" validate:"required"`

	// Logging
	LogLevel string `default:"INFO" validate:"oneof=DEBUG INFO WARN ERROR"`
}

// Load loads configuration from config file and environment variables using viper.
func Load() *Config {
	cfg := Config{}

	// Initialize viper
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "__"))

	// Set defaults for the config struct
	if err := defaults.Set(&cfg); err != nil {
		panic("failed to set struct defaults: " + err.Error())
	}

	// Bind env vars for each field
	typeOfCfg := reflect.TypeOf(cfg)
	for i := 0; i < typeOfCfg.NumField(); i++ {
		field := typeOfCfg.Field(i)
		key := field.Tag.Get("mapstructure")
		if key == "" {
			key = toSnakeCase(field.Name)
		}
		v.BindEnv(key)
	}

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Error("Error read config file", "error", err)
		}
		logger.Warn("No config file found, using environment variables")
	}

	if err := v.Unmarshal(&cfg); err != nil {
		logger.Warn("Could not unmarshal config", "error", err)
	}

	logger.Info("Loaded config", "config", cfg.String())

	return &cfg
}

func Validate(cfg *Config) error {
	validate := validator.New()
	return validate.Struct(cfg)
}

// String returns a string representation of the config with secret fields redacted.
func (c *Config) String() string {
	v := reflect.ValueOf(*c)
	t := reflect.TypeOf(*c)
	var sb strings.Builder
	sb.WriteString("Config{")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		value := v.Field(i).Interface()
		if field.Tag.Get("secret") == "true" {
			value = "***REDACTED***"
		}
		sb.WriteString(name + ": " + toString(value))
		if i < t.NumField()-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// toString converts interface{} to string for String
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(str string) string {
	runes := []rune(str)
	var out []rune
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if !unicode.IsUpper(prev) || nextLower {
				out = append(out, '_')
			}
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}
