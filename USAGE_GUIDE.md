# SboxAgent - Руководство по использованию (Phase 2)

## 🎯 Что умеет SboxAgent на данный момент

SboxAgent Phase 2 - это полнофункциональный демон для управления конфигурациями прокси-клиентов. Вот что он может делать:

### ✅ Основные возможности

1. **Импорт конфигураций**
   - Из JSON файлов
   - Через sboxmgr CLI с подписками
   - Валидация конфигураций
   - Поддержка sing-box, clash, xray, mihomo

2. **Интеграция с sboxmgr**
   - Генерация конфигураций
   - Валидация конфигураций
   - Список доступных клиентов
   - Retry логика и обработка ошибок

3. **Управление systemd сервисами**
   - Enable/disable сервисов
   - Start/stop/restart операций
   - Мониторинг статуса
   - Пользовательский и системный режимы

4. **Мониторинг и метрики**
   - Сбор системных метрик
   - Сбор метрик сервисов
   - Система оповещений
   - Отчеты о состоянии здоровья

5. **Оркестрация**
   - Управление жизненным циклом всех сервисов
   - Graceful shutdown
   - Агрегация статуса

## 🚀 Как начать использовать

### 1. Сборка

```bash
# Клонирование и сборка
git clone https://github.com/kpblcaoo/sboxagent.git
cd sboxagent
make build
```

### 2. Базовая конфигурация

Создайте файл `agent.yaml`:

```yaml
agent:
  name: "sboxagent"
  version: "0.1.0"
  log_level: "info"

services:
  cli:
    enabled: true
    sboxmgr_path: "sboxmgr"
    timeout: "30s"
    max_retries: 3
    retry_interval: "5s"
  
  systemd:
    enabled: true
    service_name: "sboxagent"
    user_mode: false
    auto_start: true
  
  monitoring:
    enabled: true
    interval: "30s"
    metrics_enabled: true
    alerts_enabled: true
    retention_days: 30
```

### 3. Запуск

```bash
# Запуск с конфигурацией
./sboxagent -config agent.yaml

# Запуск с настройками по умолчанию
./sboxagent
```

## 📖 Практические примеры использования

### Импорт конфигурации из файла

```bash
# Создайте конфигурацию sing-box
cat > config.json << EOF
{
  "log": {"level": "info"},
  "inbounds": [
    {
      "type": "mixed",
      "tag": "mixed-in",
      "listen": "127.0.0.1",
      "listen_port": 7890
    }
  ],
  "outbounds": [
    {
      "type": "direct",
      "tag": "direct"
    }
  ]
}
EOF

# Импортируйте конфигурацию
curl -X POST http://localhost:8080/api/v1/config/import \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "./config.json",
    "client_type": "sing-box"
  }'
```

### Импорт конфигурации через sboxmgr

```bash
# Импорт с подпиской
curl -X POST http://localhost:8080/api/v1/config/import \
  -H "Content-Type: application/json" \
  -d '{
    "subscription_url": "https://your-subscription-url.com/sub",
    "client_type": "sing-box",
    "options": {
      "exclude": "reject,blocked",
      "include": "auto"
    }
  }'
```

### Управление systemd сервисами

```bash
# Включение сервиса
curl -X POST http://localhost:8080/api/v1/systemd/enable

# Запуск сервиса
curl -X POST http://localhost:8080/api/v1/systemd/start

# Проверка статуса
curl http://localhost:8080/api/v1/systemd/status

# Остановка сервиса
curl -X POST http://localhost:8080/api/v1/systemd/stop

# Перезапуск сервиса
curl -X POST http://localhost:8080/api/v1/systemd/restart
```

### Мониторинг и метрики

```bash
# Получение метрик
curl http://localhost:8080/api/v1/monitor/metrics

# Получение оповещений
curl http://localhost:8080/api/v1/monitor/alerts

# Проверка состояния здоровья
curl http://localhost:8080/api/v1/monitor/health

# Очистка оповещений
curl -X POST http://localhost:8080/api/v1/monitor/alerts/clear
```

### Статус агента и сервисов

```bash
# Общий статус агента
curl http://localhost:8080/api/v1/status

# Статус CLI сервиса
curl http://localhost:8080/api/v1/services/cli/status

# Статус systemd сервиса
curl http://localhost:8080/api/v1/services/systemd/status

# Статус мониторинга
curl http://localhost:8080/api/v1/services/monitor/status
```

## ⚙️ Конфигурация

### Полная конфигурация

```yaml
agent:
  name: "sboxagent"
  version: "0.1.0"
  log_level: "info"

server:
  port: 8080
  host: "127.0.0.1"
  timeout: "30s"

services:
  cli:
    enabled: true
    sboxmgr_path: "sboxmgr"
    timeout: "30s"
    max_retries: 3
    retry_interval: "5s"
  
  systemd:
    enabled: true
    service_name: "sboxagent"
    user_mode: false
    auto_start: true
  
  monitoring:
    enabled: true
    interval: "30s"
    metrics_enabled: true
    alerts_enabled: true
    retention_days: 30

clients:
  sing-box:
    enabled: true
    binary_path: "/usr/local/bin/sing-box"
    config_path: "/etc/sing-box/config.json"
  
  clash:
    enabled: true
    binary_path: "/usr/local/bin/clash"
    config_path: "/etc/clash/config.yaml"
  
  xray:
    enabled: true
    binary_path: "/usr/local/bin/xray"
    config_path: "/etc/xray/config.json"

logging:
  stdout_capture: true
  aggregation: true
  retention_days: 30
  max_entries: 1000

security:
  allow_remote_api: false
  allowed_hosts: ["127.0.0.1", "::1"]
  tls_enabled: false
```

### Переменные окружения

```bash
# Основные настройки
export SBOXAGENT_AGENT_NAME="my-agent"
export SBOXAGENT_AGENT_VERSION="0.1.0"
export SBOXAGENT_AGENT_LOG_LEVEL="debug"

# Сервер
export SBOXAGENT_SERVER_PORT=9090
export SBOXAGENT_SERVER_HOST="0.0.0.0"

# CLI сервис
export SBOXAGENT_SERVICES_CLI_ENABLED=true
export SBOXAGENT_SERVICES_CLI_SBOXMGR_PATH="/usr/local/bin/sboxmgr"
export SBOXAGENT_SERVICES_CLI_TIMEOUT="60s"

# Systemd сервис
export SBOXAGENT_SERVICES_SYSTEMD_ENABLED=true
export SBOXAGENT_SERVICES_SYSTEMD_SERVICE_NAME="my-service"
export SBOXAGENT_SERVICES_SYSTEMD_USER_MODE=true

# Мониторинг
export SBOXAGENT_SERVICES_MONITORING_ENABLED=true
export SBOXAGENT_SERVICES_MONITORING_INTERVAL="1m"
```

## 🔧 Разработка и отладка

### Логирование

```bash
# Запуск с debug уровнем
./sboxagent -config agent.yaml -log-level debug

# Просмотр логов в реальном времени
tail -f /var/log/sboxagent.log
```

### Тестирование

```bash
# Unit тесты
go test ./...

# Интеграционные тесты
go test ./tests/integration/...

# Smoke тесты
go test ./tests/smoke_test.go
```

### Отладка конфигурации

```bash
# Проверка конфигурации
./sboxagent -config agent.yaml -validate-only

# Вывод конфигурации по умолчанию
./sboxagent -print-default-config
```

## 📊 Мониторинг

### Метрики

SboxAgent собирает следующие метрики:

- **Системные метрики**:
  - Uptime агента
  - Использование памяти
  - Использование CPU
  - Использование диска

- **Сервисные метрики**:
  - Статус сервисов
  - Время отклика
  - Количество ошибок

- **Производительность**:
  - Response times
  - Throughput
  - Error rates

### Оповещения

Система оповещений с уровнями:

- **INFO**: Информационные сообщения
- **WARNING**: Предупреждения
- **ERROR**: Ошибки
- **CRITICAL**: Критические ошибки

### Примеры оповещений

```json
{
  "id": "alert-001",
  "level": "WARNING",
  "message": "High memory usage detected",
  "timestamp": "2025-06-28T15:30:00Z",
  "data": {
    "memory_usage": "85%",
    "threshold": "80%"
  }
}
```

## 🔒 Безопасность

### Рекомендации

1. **Запуск под непривилегированным пользователем**:
   ```bash
   sudo useradd -r -s /bin/false sboxagent
   sudo chown sboxagent:sboxagent /etc/sboxagent/
   ```

2. **Ограничение доступа к API**:
   ```yaml
   security:
     allow_remote_api: false
     allowed_hosts: ["127.0.0.1"]
   ```

3. **TLS для удаленного доступа**:
   ```yaml
   security:
     tls_enabled: true
     tls_cert_file: "/etc/sboxagent/cert.pem"
     tls_key_file: "/etc/sboxagent/key.pem"
   ```

## 🚨 Устранение неполадок

### Частые проблемы

1. **sboxmgr не найден**:
   ```bash
   # Проверьте путь к sboxmgr
   which sboxmgr
   
   # Установите sboxmgr
   pip install sboxmgr
   ```

2. **Permission denied для systemd**:
   ```bash
   # Проверьте права пользователя
   sudo usermod -aG systemd sboxagent
   ```

3. **Порт уже занят**:
   ```bash
   # Измените порт в конфигурации
   server:
     port: 8081
   ```

### Логи для отладки

```bash
# Просмотр логов агента
journalctl -u sboxagent -f

# Просмотр логов systemd
journalctl -u sboxagent.service

# Проверка статуса сервиса
systemctl status sboxagent
```

## 📚 Дополнительные ресурсы

- [Phase 2 Implementation Summary](PHASE2_IMPLEMENTATION_SUMMARY.md)
- [Configuration Reference](docs/configuration.md)
- [API Documentation](docs/api.md)
- [Architecture Guide](docs/architecture.md)

## 🤝 Поддержка

Если у вас есть вопросы или проблемы:

1. Проверьте [документацию](docs/)
2. Посмотрите [примеры](examples/)
3. Создайте [issue](https://github.com/kpblcaoo/sboxagent/issues)

---

**Phase 2 полностью готова к использованию!** 🎉 