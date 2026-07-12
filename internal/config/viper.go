package config

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// LoadConfig reads configuration from file and environment variables, validates it, and returns a Config pointer.
func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	setDefaults(v)

	bindEnvs(v, &Config{})

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8000)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.env", "develop")
	v.SetDefault("mongo.uri", "mongodb://localhost:27017")
	v.SetDefault("mongo.database", "business_chat")
}

func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) {
	ifaceVal := reflect.ValueOf(iface)
	if ifaceVal.Kind() == reflect.Ptr {
		ifaceVal = ifaceVal.Elem()
	}
	if ifaceVal.Kind() != reflect.Struct {
		return
	}
	ifaceType := ifaceVal.Type()
	for i := 0; i < ifaceVal.NumField(); i++ {
		field := ifaceType.Field(i)
		tagVal := field.Tag.Get("mapstructure")
		if tagVal == "" || tagVal == "," {
			continue
		}
		newParts := append(parts, tagVal)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			bindEnvs(v, reflect.New(fieldType).Interface(), newParts...)
		} else {
			key := strings.Join(newParts, ".")
			_ = v.BindEnv(key)
		}
	}
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			var errMsg []string
			for _, valErr := range valErrs {
				errMsg = append(errMsg, fmt.Sprintf("field '%s' failed validation on tag '%s'", valErr.Namespace(), valErr.Tag()))
			}
			return fmt.Errorf("configuration validation error: %s", strings.Join(errMsg, "; "))
		}
		return fmt.Errorf("failed to validate config: %w", err)
	}
	return nil
}

func (c *Config) GetMaskedMongoURI() string {
	if !strings.HasPrefix(c.Mongo.URI, "mongodb://") && !strings.HasPrefix(c.Mongo.URI, "mongodb+srv://") {
		return "***"
	}
	u, err := url.Parse(c.Mongo.URI)
	if err != nil {
		return "***"
	}
	if u.User != nil {
		u.User = url.UserPassword("username_masked", "password_masked")
	}
	return u.String()
}
