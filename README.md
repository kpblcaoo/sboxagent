# SboxAgent

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-GPL--3.0-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.1.0--alpha-orange.svg)](VERSION)

**SboxAgent** — это Go-демон для управления конфигурациями sing-box прокси, интегрированный с [sboxmgr](https://github.com/kpblcaoo/sboxmgr) — Python CLI-инструментом для управления подписками.

## 🚀 Возможности

- **Автоматическое управление sing-box**: мониторинг и обновление конфигураций
- **Интеграция с sboxmgr**: получение обновлений через sboxctl
- **Структурированное логирование**: JSON-формат, уровни логирования
- **Event-driven архитектура**: асинхронная обработка событий
- **Health monitoring**: проверка состояния компонентов
- **Log aggregation**: сбор и анализ логов в памяти
- **Systemd integration**: автоматический запуск и управление
- **Security-first**: запуск под непривилегированным пользователем

## 📋 Требования

- **Go 1.21+** для сборки
- **Linux** с systemd для установки
- **sboxmgr** и **sboxctl** для интеграции
- **sing-box** для управления прокси

## 🏗️ Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   sboxctl       │    │   sing-box      │    │   sboxagent     │
│   (sboxmgr)     │◄──►│   (proxy)       │◄──►│   (daemon)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   systemd       │    │   event         │
                       │   service       │    │   dispatcher    │
                       └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │   log           │
                                              │   aggregator    │
                                              └─────────────────┘
```

### Основные компоненты

- **Agent Core**: основной цикл жизни приложения
- **Sboxctl Service**: мониторинг sboxctl команд
- **Event Dispatcher**: асинхронная обработка событий
- **Log Aggregator**: сбор и хранение логов
- **Health Checker**: мониторинг состояния системы
- **Configuration Manager**: загрузка и валидация конфигурации

## 🛠️ Установка

### Быстрая установка

```bash
# Клонировать репозиторий
git clone https://github.com/kpblcaoo/sboxagent.git
cd sboxagent

# Собрать и установить
make build
sudo ./scripts/install.sh
```

### Ручная установка

```bash
# 1. Собрать бинарник
make build

# 2. Создать пользователя и группу
sudo useradd --system --no-create-home --shell /bin/false sboxagent
sudo groupadd --system sboxagent
sudo usermod -a -G sboxagent sboxagent

# 3. Скопировать файлы
sudo cp bin/sboxagent /usr/local/bin/
sudo cp scripts/sboxagent.service /etc/systemd/system/

# 4. Создать конфигурацию
sudo mkdir -p /etc/sboxagent
sudo cp examples/agent.yaml /etc/sboxagent/

# 5. Запустить сервис
sudo systemctl daemon-reload
sudo systemctl enable sboxagent
sudo systemctl start sboxagent
```

## ⚙️ Конфигурация

Основной файл конфигурации: `/etc/sboxagent/agent.yaml`

```yaml
# SboxAgent Configuration
agent:
  name: "sboxagent"
  version: "0.1.0-alpha"
  log_level: "info"
  log_format: "json"

# Sboxctl service configuration
sboxctl:
  command: ["sboxctl", "status"]
  interval: "30s"
  timeout: "10s"
  stdout_capture: true
  health_check:
    enabled: true
    interval: "60s"

# Log aggregator configuration
aggregator:
  max_entries: 1000
  max_age: "24h"

# Health checker configuration
health:
  check_interval: "30s"
  timeout: "5s"
```

## 🚀 Использование

### Команды

```bash
# Показать версию
sboxagent -version

# Запуск с дефолтной конфигурацией
sboxagent

# Запуск с кастомной конфигурацией
sboxagent -config /path/to/config.yaml

# Debug режим
sboxagent -debug

# Изменить уровень логирования
sboxagent -log-level debug
```

### Управление сервисом

```bash
# Статус сервиса
sudo systemctl status sboxagent

# Запуск/остановка
sudo systemctl start sboxagent
sudo systemctl stop sboxagent

# Перезапуск
sudo systemctl restart sboxagent

# Просмотр логов
sudo journalctl -u sboxagent -f
```

### Удаление

```bash
# Удалить сервис и бинарник
sudo ./scripts/uninstall.sh

# Удалить конфигурацию (опционально)
# Скрипт спросит о удалении конфига и пользователя
```

## 🧪 Разработка

### Сборка

```bash
# Обычная сборка
make build

# Сборка для Linux
make build-linux

# Очистка
make clean
```

### Тестирование

```bash
# Запуск тестов
make test

# Тесты с покрытием
make test-coverage

# Интеграционные тесты
make test-integration

# Бенчмарки
make benchmark
```

### Качество кода

```bash
# Форматирование
make fmt

# Линтинг
make lint

# Полная проверка
make check
```

### Документация

```bash
# Генерация документации
make docs

# Проверка покрытия докстрингами
make docs-check

# Локальный сервер документации
make docs-serve
```

## 📚 Документация

- [Архитектура](docs/README.md) — техническая документация
- [Планы разработки](plans/) — roadmap и задачи
- [Правила](.cursor/rules/) — coding standards и best practices
- [Тесты](tests/) — unit и integration тесты

## 🔧 Troubleshooting

### Сервис не запускается

```bash
# Проверить статус
sudo systemctl status sboxagent

# Посмотреть логи
sudo journalctl -u sboxagent -n 50

# Проверить конфигурацию
sboxagent -config /etc/sboxagent/agent.yaml -debug
```

### sboxctl не найден

Убедитесь, что sboxmgr и sboxctl установлены и доступны в PATH:

```bash
# Проверить установку sboxmgr
which sboxctl

# Установить sboxmgr (если не установлен)
pip install sboxmgr
```

### Проблемы с правами

```bash
# Проверить права пользователя
ls -la /usr/local/bin/sboxagent
ls -la /etc/sboxagent/

# Исправить права
sudo chown sboxagent:sboxagent /usr/local/bin/sboxagent
sudo chown -R sboxagent:sboxagent /etc/sboxagent/
```

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### Требования к коду

- Следуйте [правилам](.cursor/rules/) проекта
- Добавляйте тесты для новой функциональности
- Обновляйте документацию
- Используйте conventional commits

## 📄 Лицензия

Этот проект лицензирован под GPL-3.0 — см. файл [LICENSE](LICENSE) для деталей.

## 🔗 Ссылки

- [sboxmgr](https://github.com/kpblcaoo/sboxmgr) — Python CLI для управления подписками
- [sing-box](https://github.com/SagerNet/sing-box) — универсальный прокси-инструмент
- [Планы разработки](plans/) — roadmap и задачи проекта

---

**Версия**: 0.1.0-alpha  
**Последнее обновление**: 2025-06-27  
**Поддерживаемые платформы**: Linux (systemd) 