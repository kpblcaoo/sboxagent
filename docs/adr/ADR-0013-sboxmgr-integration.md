# ADR-0013: SBoxAgent Integration with sboxmgr

> **Status**: Accepted  
> **Date**: 2025-06-27  
> **Author**: Mikhail Stepanov  
> **Supersedes**: None

## 🎯 Context

SBoxAgent должен интегрироваться с CLI-инструментом для управления VPN конфигурациями. Необходимо определить архитектуру взаимодействия между агентом и CLI-компонентом, учитывая dual-path архитектуру sboxctl.

## 📋 Decision

### ✅ Принятое решение: Dual-Path Architecture с HTTP API интеграцией

**sboxctl** является CLI-инструментом для управления VPN конфигурациями с dual-path архитектурой:
- **Path A**: autoupdater (systemd timer + shell script)
- **Path B**: sboxagent (Go daemon) - продвинутый путь

Интеграция происходит через HTTP API с JSON протоколом, определенным в sboxctl.

### 🔗 Архитектура взаимодействия

```
Dual-Path Architecture:
┌─────────────────────────────────────────────────────────────┐
│                    Dual-Path Architecture                   │
├─────────────────────────────────────────────────────────────┤
│  Path A: autoupdater        │  Path B: sboxagent           │
│  - systemd timer            │  - Go daemon                 │
│  - shell script             │  - orchestrator              │
│  - simple logging           │  - logger                    │
│  - basic health check       │  - watchdog                  │
├─────────────────────────────────────────────────────────────┤
│                    sboxctl (oneshot CLI)                   │
│  - subscription processing                                  │
│  - config generation                                        │
│  - plugin system                                            │
│  - stdout logging (captured by agent)                       │
└─────────────────────────────────────────────────────────────┘

Integration Flow:
sboxctl CLI → stdout → sboxagent (Go)
     ↓           ↓           ↓
  config    logs/metrics  long-running
generation   capture      service logic
```

### 📊 Протокол взаимодействия

#### Тип интеграции: HTTP API + stdout capture
- **Механизм**: HTTP API endpoints + stdout capture
- **Формат**: JSON API определенный в sboxctl/agent/protocol.py
- **Направление**: sboxctl → sboxagent (команды) + stdout → sboxagent (логи)

#### Команды API:
- `validate` - валидация конфигурации VPN клиентов
- `install` - установка VPN клиентов (sing-box, xray, clash, hysteria, mihomo)
- `check` - проверка статуса клиентов
- `version` - информация о версии агента

#### Поддерживаемые VPN клиенты:
- **sing-box** - основной клиент
- **xray** - альтернативный клиент
- **clash** - альтернативный клиент
- **hysteria** - альтернативный клиент
- **mihomo** - альтернативный клиент

#### Пример запроса:
```json
{
  "command": "validate",
  "version": "1.0",
  "trace_id": "abc123",
  "config_path": "/etc/sing-box/config.json",
  "client_type": "sing-box",
  "strict": true
}
```

#### Пример ответа:
```json
{
  "success": true,
  "message": "Configuration validated successfully",
  "trace_id": "abc123",
  "errors": [],
  "client_detected": "sing-box",
  "client_version": "1.8.0"
}
```

## 🏗️ Техническая реализация

### Phase 1B (v0.1.0-alpha)
```go
// HTTP API server
func main() {
    router := gin.Default()
    
    // API endpoints
    router.POST("/api/v1/validate", handleValidate)
    router.POST("/api/v1/install", handleInstall)
    router.POST("/api/v1/check", handleCheck)
    router.POST("/api/v1/version", handleVersion)
    
    // Health endpoint
    router.GET("/health", handleHealth)
    
    // Stdout capture setup
    go capturesboxctlStdout()
    
    router.Run(":8080")
}

// Multi-client support
type ClientManager interface {
    Validate(configPath string, strict bool) (*ValidationResult, error)
    Install(version string, force bool) (*InstallResult, error)
    Check() (*CheckResult, error)
    GetVersion() (string, error)
}

// Client implementations
type SingBoxClient struct{}
type XrayClient struct{}
type ClashClient struct{}
type HysteriaClient struct{}
type MihomoClient struct{}
```

### Phase 2+ (v0.3.0-beta+)
- Web UI для управления
- Расширенные команды
- Алертинг и мониторинг
- Backup/restore функциональность

## ✅ Обоснование

### 1. Совместимость с существующей архитектурой
- sboxctl уже имеет AgentBridge для HTTP API
- JSON протокол уже определен и протестирован
- Dual-path архитектура поддерживает гибкость

### 2. Расширяемость и универсальность
- Поддержка множественных VPN клиентов
- Не привязан только к sing-box
- Легко добавлять новые клиенты

### 3. Надежность и масштабируемость
- HTTP API более надежен чем stdout pipe
- Stdout capture для логов и метрик
- Возможность добавления аутентификации

### 4. Dual-path гибкость
- Простой путь: autoupdater (systemd timer + shell)
- Продвинутый путь: sboxagent (Go daemon)
- Пользователь выбирает подходящий путь

## 🚧 Ограничения и риски

### Phase 1B ограничения:
- Только HTTP API (без Web UI)
- Базовые команды
- In-memory логирование
- Базовые health check'и

### Меры смягчения:
- Graceful shutdown для HTTP server
- Error handling и logging
- Health checks
- Systemd integration для надежности

## 🔄 Альтернативы (отклонены)

### ❌ stdout-based интеграция
- **Причина отклонения**: sboxctl уже использует HTTP API
- **Статус**: Несовместимо с существующей архитектурой

### ❌ Прямая интеграция с VPN клиентами
- **Причина отклонения**: Сложность, нарушение принципа разделения ответственности
- **Статус**: sboxctl уже решает эту задачу

### ❌ gRPC API
- **Причина отклонения**: Избыточная сложность для MVP
- **Статус**: HTTP API достаточно для текущих потребностей

### ❌ Только sing-box поддержка
- **Причина отклонения**: Ограничивает универсальность
- **Статус**: Поддержка множественных клиентов важна

## 📈 Будущие расширения

### Phase 1C (v0.2.0-beta)
- File-based логирование
- Prometheus metrics
- Graceful shutdown
- Hot-reload конфигурации

### Phase 2 (v0.3.0-beta)
- Web UI
- Алертинг система
- Backup/restore
- Расширенные команды

### Phase 3 (v1.0.0)
- Production readiness
- Security audit
- Performance optimization

## 🎯 Критерии успеха

### Технические:
- [ ] HTTP API работает с sboxctl AgentBridge
- [ ] JSON протокол 100% совместим
- [ ] Все команды реализованы
- [ ] Поддержка всех VPN клиентов
- [ ] Stdout capture функционирует
- [ ] Systemd service работает

### Пользовательские:
- [ ] Простая установка и настройка
- [ ] Надежная работа в production
- [ ] Возможность мониторинга и отладки
- [ ] Гибкость выбора пути (autoupdater vs sboxagent)

## 📝 Связанные документы

- [ADR-0012: SBoxAgent Architecture](./ADR-0012-sboxagent-architecture.md)
- [Phase 1B ToDo](../plans/implementation/phase-1b-todo.md)
- [ROADMAP.md](../../ROADMAP.md)
- [sboxmgr Agent Protocol](../sboxmgr/src/sboxmgr/agent/protocol.py)
- [sboxmgr Stage 4 Plan](../sboxmgr/plans/roadmap_v1.5.0/stage4-agent-integration/README.md)

---

**Примечание**: Это решение основано на изучении существующей архитектуры sboxmgr и обеспечивает максимальную совместимость с dual-path архитектурой, поддерживая как простой autoupdater путь, так и продвинутый sboxagent путь. 