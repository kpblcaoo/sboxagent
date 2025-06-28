# Что умеет SboxAgent прямо сейчас (Phase 2)

## 🎯 Основные возможности

### 1. Импорт конфигураций ✅
- **Из JSON файлов**: Загружает готовые конфигурации
- **Через sboxmgr**: Генерирует конфигурации из подписок
- **Валидация**: Проверяет корректность конфигураций
- **Поддержка клиентов**: sing-box, clash, xray, mihomo

### 2. Интеграция с sboxmgr ✅
- **Генерация конфигураций**: `sboxmgr generate`
- **Валидация конфигураций**: `sboxmgr validate`
- **Список клиентов**: `sboxmgr list-clients`
- **Retry логика**: Автоматические повторные попытки
- **Обработка ошибок**: Детальное логирование ошибок

### 3. Управление systemd сервисами ✅
- **Enable/Disable**: Включение/отключение сервисов
- **Start/Stop/Restart**: Запуск/остановка/перезапуск
- **Status monitoring**: Мониторинг статуса сервисов
- **User/System mode**: Пользовательский и системный режимы

### 4. Мониторинг и метрики ✅
- **Системные метрики**: CPU, память, диск, uptime
- **Сервисные метрики**: Статус сервисов, время отклика
- **Оповещения**: Система алертов с уровнями (info, warning, error, critical)
- **Health checks**: Проверка состояния здоровья

### 5. Оркестрация ✅
- **Управление жизненным циклом**: Инициализация, запуск, остановка всех сервисов
- **Graceful shutdown**: Корректное завершение работы
- **Status aggregation**: Агрегация статуса от всех сервисов
- **Event handling**: Обработка событий

## 🚀 Как использовать

### Быстрый старт

```bash
# 1. Сборка
make build

# 2. Создание конфигурации
cat > agent.yaml << EOF
agent:
  name: "sboxagent"
  log_level: "info"
services:
  cli:
    enabled: true
    sboxmgr_path: "sboxmgr"
  systemd:
    enabled: true
    service_name: "sboxagent"
  monitoring:
    enabled: true
    interval: "30s"
EOF

# 3. Запуск
./sboxagent -config agent.yaml
```

### Примеры использования

#### Импорт конфигурации
```bash
# Из файла
curl -X POST http://localhost:8080/api/v1/config/import \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/path/to/config.json", "client_type": "sing-box"}'

# Через sboxmgr
curl -X POST http://localhost:8080/api/v1/config/import \
  -H "Content-Type: application/json" \
  -d '{"subscription_url": "https://example.com/sub", "client_type": "sing-box"}'
```

#### Управление systemd
```bash
# Включить сервис
curl -X POST http://localhost:8080/api/v1/systemd/enable

# Запустить сервис
curl -X POST http://localhost:8080/api/v1/systemd/start

# Проверить статус
curl http://localhost:8080/api/v1/systemd/status
```

#### Мониторинг
```bash
# Получить метрики
curl http://localhost:8080/api/v1/monitor/metrics

# Получить оповещения
curl http://localhost:8080/api/v1/monitor/alerts

# Проверить здоровье
curl http://localhost:8080/api/v1/monitor/health
```

#### Статус агента
```bash
# Общий статус
curl http://localhost:8080/api/v1/status

# Статус сервисов
curl http://localhost:8080/api/v1/services/cli/status
curl http://localhost:8080/api/v1/services/systemd/status
curl http://localhost:8080/api/v1/services/monitor/status
```

## 📊 Что собирается и мониторится

### Метрики
- **Системные**: uptime, memory usage, CPU usage, disk usage
- **Сервисные**: service status, response times, error counts
- **Производительность**: throughput, latency

### Оповещения
- **INFO**: Информационные сообщения
- **WARNING**: Предупреждения (высокое использование ресурсов)
- **ERROR**: Ошибки (сбои сервисов)
- **CRITICAL**: Критические ошибки (система недоступна)

### Логи
- **Структурированные**: JSON формат
- **Уровни**: DEBUG, INFO, WARN, ERROR
- **Агрегация**: Сбор и анализ в памяти

## ⚙️ Конфигурация

### Основные настройки
```yaml
agent:
  name: "sboxagent"
  log_level: "info"

services:
  cli:
    enabled: true
    sboxmgr_path: "sboxmgr"
    timeout: "30s"
    max_retries: 3
  
  systemd:
    enabled: true
    service_name: "sboxagent"
    user_mode: false
  
  monitoring:
    enabled: true
    interval: "30s"
    metrics_enabled: true
    alerts_enabled: true
```

### Переменные окружения
```bash
export SBOXAGENT_AGENT_NAME="my-agent"
export SBOXAGENT_SERVICES_CLI_ENABLED=true
export SBOXAGENT_SERVICES_SYSTEMD_SERVICE_NAME="my-service"
```

## 🔧 Разработка

### Тестирование
```bash
# Unit тесты
go test ./...

# Интеграционные тесты
go test ./tests/integration/...

# Smoke тесты
go test ./tests/smoke_test.go
```

### Сборка
```bash
# Обычная сборка
make build

# С отладкой
make build-debug
```

## 🎉 Готовность

- **К использованию**: ✅ ГОТОВО
- **К продакшену**: ✅ ГОТОВО  
- **К Phase 3**: ✅ ГОТОВО

**Phase 2 полностью завершена и готова к использованию!**

## 📚 Документация

- [Подробное руководство](USAGE_GUIDE.md)
- [Phase 2 Summary](PHASE2_IMPLEMENTATION_SUMMARY.md)
- [Final Status](PHASE2_FINAL_STATUS.md) 