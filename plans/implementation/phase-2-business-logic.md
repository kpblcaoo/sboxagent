# Phase 2: JSON Import & Service Management

**Длительность:** 7-10 дней
**Статус:** 📋 ПЛАНИРОВАНИЕ

## 🎯 ЦЕЛИ ФАЗЫ (согласно ADR-0001)

### Основные задачи (EXECUTOR & DAEMON FOCUS)
- [ ] **JSON Configuration Import** - импорт конфигураций от sboxmgr
- [ ] **Subbox Client Management** - управление lifecycle клиентов
- [ ] **CLI Integration** - интеграция с sboxmgr CLI
- [ ] **Service Management** - systemd интеграция
- [ ] **Status Monitoring** - мониторинг и отчетность

## 📋 ДЕТАЛЬНЫЙ ПЛАН

### Day 1-2: JSON Configuration Import
- [ ] JSON parser для конфигураций
- [ ] Schema validation против sbox-common
- [ ] Configuration storage и versioning
- [ ] Rollback mechanism

### Day 3-4: CLI Integration
- [ ] sboxmgr CLI executor
- [ ] Process management
- [ ] Error handling и logging
- [ ] Configuration generation pipeline

### Day 5-6: Subbox Client Management
- [ ] Client lifecycle management (start/stop/restart)
- [ ] Configuration deployment
- [ ] Process monitoring
- [ ] Health checking

### Day 7-8: Service Management
- [ ] systemd integration
- [ ] Service discovery
- [ ] Auto-restart mechanisms
- [ ] Resource monitoring

### Day 9-10: Status & Monitoring
- [ ] Status reporting API
- [ ] Metrics collection
- [ ] Health endpoints
- [ ] Integration testing

## 🔧 ТЕХНИЧЕСКИЕ ТРЕБОВАНИЯ (GO IMPLEMENTATION)

### JSON Configuration Import
```go
// internal/config/importer.go
type ConfigImporter struct {
    validator *JSONValidator
    storage   *ConfigStorage
    logger    *Logger
}

func (ci *ConfigImporter) ImportConfig(jsonData []byte) (*Config, error) {
    // Validate against sbox-common schemas
    // Parse JSON configuration
    // Store with versioning
    // Return parsed config
}

type Config struct {
    Client    string                 `json:"client"`
    Version   string                 `json:"version"`
    CreatedAt time.Time             `json:"created_at"`
    Config    map[string]interface{} `json:"config"`
    Metadata  ConfigMetadata         `json:"metadata"`
}
```

### CLI Integration
```go
// internal/cli/executor.go
type CLIExecutor struct {
    sboxmgrPath string
    logger      *Logger
    timeout     time.Duration
}

func (ce *CLIExecutor) GenerateConfig(params GenerateParams) (*Config, error) {
    // Build sboxmgr command
    cmd := exec.Command("sboxmgr", "config", "generate", 
        "--client", params.Client,
        "--subscription", params.Subscription,
        "--output", "json")
    
    // Execute and parse JSON output
    // Validate response
    // Return parsed config
}

func (ce *CLIExecutor) ValidateConfig(config *Config) error {
    // Call sboxmgr validate command
    // Return validation result
}
```

### Subbox Client Management
```go
// internal/clients/manager.go
type ClientManager struct {
    clients map[string]SubboxClient
    configs map[string]*Config
    logger  *Logger
}

func (cm *ClientManager) DeployConfig(clientType string, config *Config) error {
    client := cm.clients[clientType]
    if client == nil {
        return fmt.Errorf("unsupported client: %s", clientType)
    }
    
    // Stop client if running
    // Deploy new configuration
    // Start client with new config
    // Verify deployment
}

type SubboxClient interface {
    Start(config *Config) error
    Stop() error
    Restart() error
    GetStatus() ClientStatus
    GetPID() (int, error)
    IsRunning() bool
}

// Implementations
type SingBoxClient struct{}
type ClashClient struct{}
type XrayClient struct{}
type MihomoClient struct{}
```

### Service Management
```go
// internal/service/manager.go
type ServiceManager struct {
    systemd *SystemdManager
    clients *ClientManager
    logger  *Logger
}

func (sm *ServiceManager) EnableService(clientType string) error {
    // Create systemd service file
    // Enable and start service
    // Monitor service status
}

func (sm *ServiceManager) RestartService(clientType string) error {
    // Graceful restart via systemd
    // Verify service health
    // Update status
}

type SystemdManager struct {
    unitPath string
}
```

### Status Monitoring
```go
// internal/status/monitor.go
type StatusMonitor struct {
    clients   *ClientManager
    collector *MetricsCollector
    logger    *Logger
}

func (sm *StatusMonitor) GetSystemStatus() SystemStatus {
    return SystemStatus{
        Clients: sm.getClientStatuses(),
        System:  sm.getSystemMetrics(),
        Health:  sm.getHealthStatus(),
    }
}

type SystemStatus struct {
    Clients map[string]ClientStatus `json:"clients"`
    System  SystemMetrics           `json:"system"`
    Health  HealthStatus            `json:"health"`
    Updated time.Time               `json:"updated"`
}
```

## 📁 СОЗДАВАЕМЫЕ ФАЙЛЫ

### internal/config/
- `importer.go` - JSON configuration importer
- `storage.go` - configuration storage
- `validator.go` - JSON schema validation
- `versioning.go` - configuration versioning

### internal/cli/
- `executor.go` - sboxmgr CLI executor
- `commands.go` - CLI command builders
- `parser.go` - JSON response parser
- `errors.go` - CLI error handling

### internal/clients/
- `manager.go` - client manager
- `singbox.go` - sing-box client implementation
- `clash.go` - clash client implementation
- `xray.go` - xray client implementation
- `mihomo.go` - mihomo client implementation

### internal/service/
- `manager.go` - service manager
- `systemd.go` - systemd integration
- `discovery.go` - service discovery
- `monitoring.go` - service monitoring

### internal/status/
- `monitor.go` - status monitoring
- `collector.go` - metrics collection
- `health.go` - health checking
- `api.go` - status API endpoints

## 🧪 ТЕСТИРОВАНИЕ

### JSON Import Tests
- [ ] Configuration parsing tests
- [ ] Schema validation tests
- [ ] Versioning mechanism tests
- [ ] Rollback functionality tests

### CLI Integration Tests
- [ ] sboxmgr execution tests
- [ ] JSON parsing tests
- [ ] Error handling tests
- [ ] Timeout handling tests

### Client Management Tests
- [ ] Client lifecycle tests
- [ ] Configuration deployment tests
- [ ] Process monitoring tests
- [ ] Health checking tests

### Service Management Tests
- [ ] systemd integration tests
- [ ] Service discovery tests
- [ ] Auto-restart tests
- [ ] Resource monitoring tests

### Integration Tests
- [ ] End-to-end configuration flow
- [ ] Multi-client management
- [ ] Service failure recovery
- [ ] Performance under load

## 📝 КРИТЕРИИ ЗАВЕРШЕНИЯ

### JSON Configuration Import: 🔄 IN PROGRESS
- [ ] JSON parsing working
- [ ] Schema validation functional
- [ ] Storage mechanism reliable
- [ ] Versioning implemented

### CLI Integration: 🔄 IN PROGRESS
- [ ] sboxmgr execution working
- [ ] JSON response parsing
- [ ] Error handling comprehensive
- [ ] Process management stable

### Client Management: 🔄 IN PROGRESS
- [ ] All client types supported
- [ ] Lifecycle management working
- [ ] Configuration deployment reliable
- [ ] Monitoring functional

### Service Management: 🔄 IN PROGRESS
- [ ] systemd integration working
- [ ] Service discovery functional
- [ ] Auto-restart reliable
- [ ] Resource monitoring accurate

### Status Monitoring: 🔄 IN PROGRESS
- [ ] Status API working
- [ ] Metrics collection accurate
- [ ] Health checking reliable
- [ ] Real-time updates functional

## 🔄 СООТВЕТСТВИЕ ADR-0001

### Роль sboxagent:
- ✅ **Executor & Daemon** - применение конфигураций и управление сервисами
- ✅ **JSON Import** - импорт конфигураций от sboxmgr
- ✅ **Service Management** - управление subbox клиентами
- ✅ **Status Monitoring** - мониторинг и отчетность

### Архитектурные принципы:
- ✅ **Single Responsibility** - только применение и управление
- ✅ **JSON Interface** - получение конфигураций через JSON
- ✅ **CLI Integration** - вызов sboxmgr как subprocess
- ✅ **License Separation** - GPL-3.0, вызывает Apache-2.0 CLI

### Интерфейс взаимодействия:
```bash
# sboxagent вызывает sboxmgr
sboxagent generate --subscription=url --client=sing-box
# Внутренне: exec("sboxmgr config generate --output=json ...")

# sboxagent применяет конфигурацию
sboxagent apply --config=config.json --client=sing-box

# sboxagent предоставляет статус
sboxagent status --format=json
```

### Dual-Path Architecture:
- **Path A**: sboxagent не участвует
- **Path B**: sboxagent → exec(sboxmgr) → JSON → apply to clients
- **Path C**: HTTP API → sboxagent → exec(sboxmgr) → JSON → apply

**Статус:** 🔄 **PHASE 2 В ПЛАНИРОВАНИИ**
