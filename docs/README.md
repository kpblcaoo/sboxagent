# SboxAgent Documentation

Generated on: Пт 27 июн 2025 23:12:21 MSK
Version: 0.1.0-alpha 

## Package Documentation

### Main Package
```

```

### Internal Packages

#### agent
```
package agent // import "github.com/kpblcaoo/sboxagent/internal/agent"

type Agent struct{ ... }
    func New(cfg *config.Config) (*Agent, error)
```

#### aggregator
```
package aggregator // import "github.com/kpblcaoo/sboxagent/internal/aggregator"

type AggregatorStats struct{ ... }
type LogEntry struct{ ... }
type LogLevel string
    const LogLevelDebug LogLevel = "debug" ...
type MemoryAggregator struct{ ... }
    func NewMemoryAggregator(log *logger.Logger, maxEntries int, maxAge time.Duration) *MemoryAggregator
```

#### api
```
No documentation available
```

#### config
```
package config // import "github.com/kpblcaoo/sboxagent/internal/config"

type AgentConfig struct{ ... }
type ClashConfig struct{ ... }
type ClientsConfig struct{ ... }
type Config struct{ ... }
    func Load(configPath string) (*Config, error)
type HealthCheckConfig struct{ ... }
type HysteriaConfig struct{ ... }
type LoggingConfig struct{ ... }
type SboxctlConfig struct{ ... }
type SecurityConfig struct{ ... }
type ServerConfig struct{ ... }
type ServicesConfig struct{ ... }
type SingBoxConfig struct{ ... }
type XrayConfig struct{ ... }
```

#### dispatcher
```
package dispatcher // import "github.com/kpblcaoo/sboxagent/internal/dispatcher"

type ConfigHandler struct{ ... }
    func NewConfigHandler(log *logger.Logger) *ConfigHandler
type Dispatcher struct{ ... }
    func NewDispatcher(log *logger.Logger) *Dispatcher
type DispatcherStats struct{ ... }
type ErrorHandler struct{ ... }
    func NewErrorHandler(log *logger.Logger) *ErrorHandler
type ErrorRecord struct{ ... }
type Event struct{ ... }
    func ConvertSboxctlEvent(sboxEvent services.SboxctlEvent) Event
type EventHandler interface{ ... }
type EventType string
    const EventTypeLog EventType = "log" ...
type HealthHandler struct{ ... }
    func NewHealthHandler(log *logger.Logger) *HealthHandler
type HealthRecord struct{ ... }
type LogHandler struct{ ... }
    func NewLogHandler(log *logger.Logger) *LogHandler
type StatusHandler struct{ ... }
    func NewStatusHandler(log *logger.Logger) *StatusHandler
```

#### health
```
package health // import "github.com/kpblcaoo/sboxagent/internal/health"

type AggregatorHealthCheck struct{ ... }
    func NewAggregatorHealthCheck(log *logger.Logger, aggregator AggregatorStats) *AggregatorHealthCheck
type AggregatorStats interface{ ... }
type ComponentHealth struct{ ... }
type DispatcherHealthCheck struct{ ... }
    func NewDispatcherHealthCheck(log *logger.Logger, dispatcher DispatcherStats) *DispatcherHealthCheck
type DispatcherStats interface{ ... }
type HealthCheck interface{ ... }
type HealthChecker struct{ ... }
    func NewHealthChecker(log *logger.Logger, checkInterval, timeout time.Duration) *HealthChecker
type HealthReport struct{ ... }
type HealthStatus string
    const HealthStatusHealthy HealthStatus = "healthy" ...
type ProcessHealthCheck struct{ ... }
    func NewProcessHealthCheck(log *logger.Logger, startTime time.Time) *ProcessHealthCheck
type SboxctlHealthCheck struct{ ... }
    func NewSboxctlHealthCheck(log *logger.Logger, service *services.SboxctlService) *SboxctlHealthCheck
type SystemHealthCheck struct{ ... }
    func NewSystemHealthCheck(log *logger.Logger) *SystemHealthCheck
```

#### logger
```
package logger // import "github.com/kpblcaoo/sboxagent/internal/logger"

type LogLevel int
    const DebugLevel LogLevel = iota ...
    func ParseLogLevel(level string) (LogLevel, error)
type Logger struct{ ... }
    func New(level string) (*Logger, error)
```

#### security
```
No documentation available
```

#### services
```
package services // import "github.com/kpblcaoo/sboxagent/internal/services"

type SboxctlEvent struct{ ... }
type SboxctlService struct{ ... }
    func NewSboxctlService(cfg config.SboxctlConfig, log *logger.Logger) (*SboxctlService, error)
```

