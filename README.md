# cfg

[![Go Reference](https://pkg.go.dev/badge/github.com/ev-kotov/cfg.svg)](https://pkg.go.dev/github.com/ev-kotov/cfg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ev-kotov/cfg)](https://goreportcard.com/report/github.com/ev-kotov/cfg)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A simple and idiomatic Go library for loading configuration with environment variable override support.

## Features

- **YAML configuration files** with sensible defaults
- **Environment variable override** for flexible deployment
- **Clean and simple API** following Go idioms
- **Type-safe configuration** with struct tags
- **Zero dependencies** beyond standard library and YAML parser

## Installation

```bash
go get github.com/ev-kotov/cfg
```

## Quick Start

### Step 1: Define your configuration structure

Create a type with your configuration structure:
```go
type Config struct {
    App struct {
        Name    string `yaml:"name" env:"APP_NAME"`
        Version string `yaml:"version" env:"APP_VERSION"`
    } `yaml:"app"`
    
    Server struct {
        Host  string `yaml:"host" env:"SERVER_HOST"`
        Port  int    `yaml:"port" env:"SERVER_PORT"`
        Debug bool   `yaml:"debug" env:"SERVER_DEBUG"`
    } `yaml:"server"`
    
    Database struct {
        Host string `yaml:"host" env:"DB_HOST"`
        Port int    `yaml:"port" env:"DB_PORT"`
        Name string `yaml:"name" env:"DB_NAME"`
    } `yaml:"database"`
}
```

### Step 2: Create a configuration file

Create `config.yaml` in your project:
```yaml
app:
  name: "my-app"
  version: "1.0.0"

server:
  host: "localhost"
  port: 3000
  debug: true

database:
  host: "localhost"
  port: 5432
  name: "myapp"
```

### Step 3: Load configuration in your application

In your `main.go`:
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/ev-kotov/cfg"
)

func main() {
    var config Config
    
    err := cfg.Load(&config,
        cfg.WithPaths(".", "./config"),
        cfg.WithName("config"),
        cfg.WithEnvPrefix("MYAPP"),
    )
    
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    fmt.Printf("Starting %s v%s\n", config.App.Name, config.App.Version)
    fmt.Printf("Server: %s:%d (debug: %t)\n", 
        config.Server.Host, config.Server.Port, config.Server.Debug)
    fmt.Printf("Database: %s:%d/%s\n",
        config.Database.Host, config.Database.Port, config.Database.Name)
}
```

### Step 4: Override with environment variables (optional)

```bash
# Override specific values
export MYAPP_SERVER_PORT=8080
export MYAPP_SERVER_DEBUG=false
export MYAPP_DATABASE_HOST=production-db.example.com

# Run your application
go run main.go
```

## Configuration Priority
The library follows a clear priority order:
1. **Environment Variables** (highest priority) - override everything
2. **YAML File** (base configuration) - provides defaults

## API Reference
### Core Functions
#### `Load(cfg interface{}, opts ...Option) error`
Loads configuration into the provided struct. The struct must be a pointer to a struct.
```go
var cfg Config
err := cfg.Load(&cfg, cfg.WithName("app"))
```

#### `MustLoad(cfg interface{}, opts ...Option)`
Panics if configuration cannot be loaded. Ideal for package-level initialization.
```go
var cfg Config

func init() {
    cfg.MustLoad(&cfg) // Panics on error
}
```

### Configuration Options
#### `WithPaths(paths ...string) Option`
Sets search paths for configuration files. Default: `[]string{".", "./config"}`
```go
cfg.Load(&cfg, 
    cfg.WithPaths(".", "/etc/myapp", "./configs", "/opt/app/config")
)
```

#### `WithName(name string) Option`
Sets the configuration file name (without extension). Default: `"config"`
```go
cfg.Load(&cfg, cfg.WithName("app")) // Looks for app.yaml
```

#### `WithEnvPrefix(prefix string) Option`
Sets the prefix for environment variables. Default: `"APP"`
```go
cfg.Load(&cfg, cfg.WithEnvPrefix("MYAPP")) // MYAPP_* variables
```

## Struct Tags
### `yaml` tag
Maps struct fields to YAML configuration keys.

### `env` tag
Specifies the environment variable name (without prefix) that can override the field.

```go
type Config struct {
    Port    int    `yaml:"port" env:"PORT"`      // Can be overridden
    Timeout int    `yaml:"timeout"`              // Cannot be overridden
    Host    string `yaml:"host" env:"HOST"`      // Can be overridden
}
```

**Important:** Only fields with `env` tags can be overridden by environment variables.

## Environment Variable Names

Environment variables follow this pattern:
```
<ENV_PREFIX>_<ENV_TAG>
```

### Examples with `WithEnvPrefix("MYAPP")`:

| Struct Field | env Tag | Environment Variable |
|-------------|---------|---------------------|
| `Port int` | `env:"PORT"` | `MYAPP_PORT` |
| `Host string` | `env:"DB_HOST"` | `MYAPP_DB_HOST` |
| `Debug bool` | `env:"DEBUG"` | `MYAPP_DEBUG` |

## File Search Behavior
- Searches paths in the order they are provided
- Uses the **first found** configuration file
- Stops searching after finding a valid file
- Returns no error if no file is found (continues with env vars only)

## Examples

### Basic Example
```go
package main

import "github.com/ev-kotov/cfg"

func main() {
    var config struct {
        Host string `yaml:"host" env:"HOST"`
        Port int    `yaml:"port" env:"PORT"`
    }
    
    cfg.Load(&config) // Uses defaults: ./, ./config, config.yaml, APP_ prefix
}
```

### Different Environments
**Development (uses YAML defaults):**
```bash
go run main.go
```

**Staging (partial overrides):**
```bash
export MYAPP_PORT=8080
export MYAPP_DEBUG=true
go run main.go
```

## Contributing
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `go test -v`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments
- Built with the Go standard library principles in mind
- YAML parsing powered by [go-yaml/yaml](https://github.com/go-yaml/yaml)