// Package cfg A simple and idiomatic Go library for loading configuration with environment variable override support.
package cfg

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// Action implements func for main parameters.
type Action func(*parameters)

type parameters struct {
	paths     []string
	name      string
	envPrefix string
}

// WithPaths set path for find config files.
func WithPaths(paths ...string) Action {
	return func(o *parameters) {
		o.paths = paths
	}
}

// WithName set config name.
func WithName(name string) Action {
	return func(o *parameters) {
		o.name = name
	}
}

// WithEnvPrefix set prefix for environment variables.
func WithEnvPrefix(prefix string) Action {
	return func(o *parameters) {
		o.envPrefix = strings.TrimSuffix(strings.ToUpper(prefix), "_")
	}
}

// MustLoad downloads the configuration or panics.
func MustLoad(cfg any, paramsAction ...Action) {
	if err := Load(cfg, paramsAction...); err != nil {
		panic("cfg: failed to load config: " + err.Error())
	}
}

// Load downloads the configuration
func Load(cfg any, paramsActions ...Action) error {
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	p := defaultParameters()
	for _, paramAction := range paramsActions {
		paramAction(p)
	}

	// first load from YAML
	if err := loadFromYaml(cfg, p); err != nil {
		return fmt.Errorf("unload config file: %w", err)
	}

	// then override with environment variables
	if err := loadFromEnv(cfg, p); err != nil {
		return fmt.Errorf("load env: %w", err)
	}

	return nil
}

func validateConfig(cfg any) error {
	if cfg == nil {
		return fmt.Errorf("config must not be nil")
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer to struct")
	}

	if v.IsNil() {
		return fmt.Errorf("config pointer must not be nil")
	}

	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to struct")
	}

	return nil
}

func defaultParameters() *parameters {
	return &parameters{
		paths:     []string{".", "./config"},
		name:      "config",
		envPrefix: "APP",
	}
}

func loadFromYaml(cfg any, parameters *parameters) error {
	for _, path := range parameters.paths {
		fullName := filepath.Join(path, parameters.name+".yaml")
		data, err := os.ReadFile(fullName)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("unread file %s: %w", fullName, err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("unparse yaml %s: %w", fullName, err)
		}

		return nil
	}

	return nil
}

func loadFromEnv(cfg any, params *parameters) error {
	v := reflect.ValueOf(cfg).Elem()
	return loadStructFromEnv(v, params.envPrefix)
}

func loadStructFromEnv(v reflect.Value, envPrefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		structField := t.Field(i)

		// Рекурсивно обрабатываем вложенные структуры
		if field.Kind() == reflect.Struct {
			if err := loadStructFromEnv(field, envPrefix); err != nil {
				return err
			}
			continue
		}

		envVar := getEnvVarName(structField, envPrefix)
		if envValue, exists := os.LookupEnv(envVar); exists {
			if err := setFieldFromEnv(field, envValue); err != nil {
				return fmt.Errorf("set field %s from env %s: %w",
					structField.Name, envVar, err)
			}
		}
	}

	return nil
}

func getEnvVarName(field reflect.StructField, envPrefix string) string {
	// Используем тег env, если указан
	if envTag := field.Tag.Get("env"); envTag != "" {
		envName := strings.ToUpper(envTag)
		if envPrefix != "" {
			return envPrefix + "_" + envName
		}
		return envName
	}

	// Если тег env не указан, НЕ создаем автоматическое имя
	// Пользователь должен явно указать тег env для полей, которые хочет переопределять
	return ""
}

func setFieldFromEnv(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}
	return nil
}
