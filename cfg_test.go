package cfg

import (
	"os"
	"testing"
)

type TestConfig struct {
	App      TestApp      `yaml:"app"`
	Server   TestServer   `yaml:"server"`
	Database TestDatabase `yaml:"database"`
	Features TestFeatures `yaml:"features"`
}

type TestApp struct {
	Name    string `yaml:"name" env:"APP_NAME"`
	Version string `yaml:"version" env:"APP_VERSION"`
}

type TestServer struct {
	Host  string `yaml:"host" env:"SERVER_HOST"`
	Port  int    `yaml:"port" env:"SERVER_PORT"`
	Debug bool   `yaml:"debug" env:"SERVER_DEBUG"`
}

type TestDatabase struct {
	Host string `yaml:"host" env:"DB_HOST"`
	Port int    `yaml:"port" env:"DB_PORT"`
	Name string `yaml:"name" env:"DB_NAME"`
}

type TestFeatures struct {
	Enabled bool `yaml:"enabled" env:"FEATURES_ENABLED"`
	Timeout int  `yaml:"timeout" env:"FEATURES_TIMEOUT"`
}

func TestMustPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustLoad to panic on invalid config")
		}
	}()

	var cfg TestConfig
	MustLoad(cfg) // Должен быть &cfg - вызовет panic
}

func TestNilConfig(t *testing.T) {
	err := Load(nil)

	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestConfigFileNotFound(t *testing.T) {
	var cfg TestConfig

	err := Load(&cfg,
		WithPaths("./whereAreYou"),
		WithName("missing"),
		WithEnvPrefix("TEST"),
	)

	if err != nil {
		t.Fatalf("Load should not fail when file not found, got: %v", err)
	}
}

func TestConfigFilesNotFound(t *testing.T) {
	var cfg TestConfig

	err := Load(&cfg,
		WithPaths("./whereAreYou", "./test", "/test/whereIsMyMind"),
		WithName("config"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
}

func TestNonPointerConfig(t *testing.T) {
	var cfg TestConfig

	err := Load(cfg) // Должен быть &cfg

	if err == nil {
		t.Error("Expected error for non-pointer config")
	}
}

func TestEnvPrefix(t *testing.T) {
	// Правильные имена: MYAPP_ + env тег
	err := os.Setenv("MYAPP_SERVER_PORT", "7070")
	if err != nil {
		t.Fatalf("Failed to set env var MYAPP_SERVER_PORT: %v", err)
	}
	err = os.Setenv("SERVER_PORT", "8080") // Без префикса - не должно использоваться
	if err != nil {
		t.Fatalf("Failed to set env var SERVER_PORT: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("MYAPP_SERVER_PORT")
		_ = os.Unsetenv("SERVER_PORT")
	}()

	var cfg TestConfig

	err = Load(&cfg,
		WithPaths("./test"),
		WithName("config"),
		WithEnvPrefix("MYAPP"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Должен использовать переменную с префиксом
	if cfg.Server.Port != 7070 {
		t.Errorf("Expected server.port 7070 from MYAPP_SERVER_PORT, got %d", cfg.Server.Port)
	}
}

func TestLoadFromYamlFile(t *testing.T) {
	var cfg TestConfig

	err := Load(&cfg,
		WithPaths("./test"),
		WithName("config"),
		WithEnvPrefix("TEST"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.App.Name != "test-app" {
		t.Errorf("Expected app.name 'test-app', got '%s'", cfg.App.Name)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Expected server.port 3000, got %d", cfg.Server.Port)
	}

	if cfg.Database.Host != "db.localhost" {
		t.Errorf("Expected database.host 'db.localhost', got '%s'", cfg.Database.Host)
	}
}

func TestLoadEnvAndOverrideYaml(t *testing.T) {
	// Правильные имена: TEST_ + env тег
	err := os.Setenv("TEST_APP_NAME", "env-override-app") // для App.Name с env:"APP_NAME"
	if err != nil {
		t.Fatalf("Failed to set env TEST_APP_NAME: %v", err)
	}
	err = os.Setenv("TEST_SERVER_PORT", "9090") // для Server.Port с env:"SERVER_PORT"
	if err != nil {
		t.Fatalf("Failed to set env TEST_SERVER_PORT: %v", err)
	}
	err = os.Setenv("TEST_SERVER_DEBUG", "false") // для Server.Debug с env:"SERVER_DEBUG"
	if err != nil {
		t.Fatalf("Failed to set env TEST_SERVER_DEBUG: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_APP_NAME")
		_ = os.Unsetenv("TEST_SERVER_PORT")
		_ = os.Unsetenv("TEST_SERVER_DEBUG")
	}()

	var cfg TestConfig

	err = Load(&cfg,
		WithPaths("./test"),
		WithName("config"),
		WithEnvPrefix("TEST"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.App.Name != "env-override-app" {
		t.Errorf("Expected app.name 'env-override-app' from env, got '%s'", cfg.App.Name)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected server.port 9090 from env, got %d", cfg.Server.Port)
	}

	if cfg.Server.Debug != false {
		t.Errorf("Expected server.debug false from env, got %t", cfg.Server.Debug)
	}

	// Проверяем, что непереопределенные поля остались из YAML
	if cfg.Database.Host != "db.localhost" {
		t.Errorf("Expected database.host 'db.localhost' from file, got '%s'", cfg.Database.Host)
	}
}

func TestNestedStructs(t *testing.T) {
	// Правильные имена: TEST_ + env тег
	err := os.Setenv("TEST_FEATURES_ENABLED", "true")
	if err != nil {
		t.Fatalf("Failed to set env TEST_FEATURES_ENABLED: %v", err)
	}
	err = os.Setenv("TEST_FEATURES_TIMEOUT", "45")
	if err != nil {
		t.Fatalf("Failed to set env TEST_FEATURES_TIMEOUT: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_FEATURES_ENABLED")
		_ = os.Unsetenv("TEST_FEATURES_TIMEOUT")
	}()

	var cfg TestConfig

	err = Load(&cfg,
		WithPaths("./test"),
		WithName("config_override"), // Используем другой конфиг
		WithEnvPrefix("TEST"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Проверяем значения из config_override.yaml
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server.port 8080 from config_override.yaml, got %d", cfg.Server.Port)
	}

	if cfg.Server.Debug != false {
		t.Errorf("Expected server.debug false from config_override.yaml, got %t", cfg.Server.Debug)
	}

	// Проверяем, что env переопределил файл для features
	if cfg.Features.Enabled != true {
		t.Errorf("Expected features.enabled true from env, got %t", cfg.Features.Enabled)
	}

	if cfg.Features.Timeout != 45 {
		t.Errorf("Expected features.timeout 45 from env, got %d", cfg.Features.Timeout)
	}
}

func TestDeeplyNestedStructs(t *testing.T) {
	type DeepConfig struct {
		Level1 struct {
			Level2 struct {
				Level3 struct {
					Value string `yaml:"value" env:"DEEP_VALUE"`
				} `yaml:"level3"`
			} `yaml:"level2"`
		} `yaml:"level1"`
	}

	// Правильное имя: TEST_ + env тег
	err := os.Setenv("TEST_DEEP_VALUE", "env-override-value")
	if err != nil {
		t.Fatalf("Failed to set env TEST_DEEP_VALUE: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_DEEP_VALUE")
	}()

	var cfg DeepConfig

	err = Load(&cfg,
		WithPaths("./test"),
		WithName("deep_config"),
		WithEnvPrefix("TEST"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Проверяем, что env переопределил глубоко вложенное поле
	if cfg.Level1.Level2.Level3.Value != "env-override-value" {
		t.Errorf("Expected deep nested value 'env-override-value' from env, got '%s'",
			cfg.Level1.Level2.Level3.Value)
	}
}

func TestNoEnvOverrideWithoutTag(t *testing.T) {
	type SimpleConfig struct {
		Name    string `yaml:"name"`                  // НЕТ env тега
		Version string `yaml:"version" env:"VERSION"` // ЕСТЬ env тег
	}

	// Пытаемся установить переменные
	err := os.Setenv("TEST_NAME", "should-not-work") // Не должно работать - нет тега
	if err != nil {
		t.Fatalf("Failed to set env TEST_NAME: %v", err)
	}
	err = os.Setenv("TEST_VERSION", "2.0.0") // Должно работать - есть тег
	if err != nil {
		t.Fatalf("Failed to set env TEST_VERSION: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_NAME")
		_ = os.Unsetenv("TEST_VERSION")
	}()

	var cfg SimpleConfig

	err = Load(&cfg,
		WithPaths("./test"),
		WithName("simple_config"),
		WithEnvPrefix("TEST"),
	)

	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Name не должен быть переопределен (нет env тега)
	if cfg.Name == "should-not-work" {
		t.Errorf("Field without env tag should not be overridden by environment variable")
	}

	// Version должен быть переопределен (есть env тег)
	if cfg.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0' from env, got '%s'", cfg.Version)
	}
}
