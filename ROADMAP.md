# 🗺️ Subbox Roadmap

> Version: 1.0 — Last Updated: 2025-06-27  
> Maintainer: Mikhail Stepanov (kpblcaoo@gmail.com)  
> License: GPL-3.0

## 📋 Версионирование

### Схема версионирования: Semantic Versioning (SemVer)
```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

- **MAJOR**: Несовместимые изменения API
- **MINOR**: Новая функциональность (обратно совместимая)
- **PATCH**: Исправления багов (обратно совместимые)
- **PRERELEASE**: alpha, beta, rc (release candidate)

### Примеры версий:
- `v0.1.0-alpha` - Первый alpha релиз
- `v0.2.0-beta` - Beta версия с новыми фичами
- `v1.0.0-rc.1` - Release candidate для v1.0.0
- `v1.0.0` - Первый стабильный релиз

## 🧩 Subbox-Common Improvements

### ✅ Реализовано (2025-06-27)
- **Версионирование схем**: `$id`, `$schema`, `version`, `$comment` поля
- **CLI валидация**: `cli.py verify` для YAML/JSON конфигураций
- **Тест-валидация**: `scripts/test_schema.py` для автоматического тестирования
- **Security fields**: Встроенные поля безопасности в схеме
- **API протокол**: Конкретизированный `$ref` на agent_config схему
- **Protocol versioning**: `protocol_version` поле в API запросах

### 🔄 Планируется (Phase 1C+)
- **Auto-expansion опций**: Поддержка omitempty полей в Go
- **Markdown документация**: Автогенерация docs/ из схем
- **Code generation**: Go структуры из JSON схем
- **OpenAPI/Swagger**: Автогенерация API документации
- **CI/CD**: Автоматическая проверка схем в GitHub Actions
- **Changelog generator**: Автоматическое отслеживание изменений схем

### 🛠 Рекомендации для разработки
1. **Валидация запросов**: Использовать схемы для автоматической валидации в Go
2. **HTTP API версионирование**: `/api/v1/update_config` с protocol_version
3. **Security hardening**: Обязательная проверка allowed_hosts и api_token
4. **Schema evolution**: Соблюдать backward compatibility в рамках major версий

## 🚀 Роадмап релизов

### Phase 0: Foundation (Текущий этап)
**Статус**: ✅ Завершен  
**Версия**: `v0.0.1` - `v0.0.5`

- [x] Инициализация репозитория
- [x] Архитектурные решения (ADR-0012)
- [x] План безопасности
- [x] Структура планов
- [x] ToDo для Phase 1B

---

### Phase 1B: Agent Bootstrap
**Статус**: 🚧 В разработке  
**Версия**: `v0.1.0-alpha`  
**Целевая дата**: 2025-07-10  
**Длительность**: 8-12 дней

#### 🎯 Цель
MVP sboxagent с базовой функциональностью перехвата stdout sboxctl.

#### 📋 Ключевые компоненты
- [ ] Базовая структура проекта (Go модуль)
- [ ] Конфигурация (viper + agent.yaml)
- [ ] Stdout Listener (exec.Cmd + sboxctl)
- [ ] Event Dispatcher (маршрутизация событий)
- [ ] Log Aggregator (in-memory)
- [ ] Health Checker (базовые проверки)
- [ ] Сборка и systemd service

#### ✅ Критерии готовности для v0.1.0-alpha
- [ ] Агент запускает sboxctl и перехватывает stdout
- [ ] Парсит JSON события и обрабатывает их
- [ ] Сохраняет логи в памяти
- [ ] Выполняет базовые health check'и
- [ ] Собирается и устанавливается
- [ ] Работает как systemd service

#### 🔗 Интеграция с sboxctl
- **Тип**: Односторонняя (агент → sboxctl)
- **Механизм**: exec.Cmd с pipe для stdout
- **События**: JSON через stdout от sboxctl
- **Статус**: Готов к реализации

---

### Phase 1C: Production Readiness
**Статус**: 📋 Планируется  
**Версия**: `v0.2.0-beta`  
**Целевая дата**: 2025-07-25  
**Длительность**: 10-15 дней

#### 🎯 Цель
Улучшение надежности и production-готовность.

#### 📋 Новые компоненты
- [ ] File-based логирование (вместо in-memory)
- [ ] HTTP API (health, metrics, logs)
- [ ] Graceful shutdown (context + WaitGroup)
- [ ] Event replay механизм
- [ ] Расширенные health check'и
- [ ] Метрики (Prometheus)
- [ ] Конфигурация hot-reload

#### ✅ Критерии готовности для v0.2.0-beta
- [ ] Логи сохраняются в файлы с ротацией
- [ ] HTTP API доступен на /health, /metrics, /logs
- [ ] Graceful shutdown работает корректно
- [ ] Метрики экспортируются в Prometheus формате
- [ ] Hot-reload конфигурации работает
- [ ] Интеграционные тесты проходят

---

### Phase 2: Full Integration
**Статус**: 📋 Планируется  
**Версия**: `v0.3.0-beta`  
**Целевая дата**: 2025-08-15  
**Длительность**: 15-20 дней

#### 🎯 Цель
Полная интеграция с sboxctl и продвинутые возможности.

#### 📋 Новые компоненты
- [ ] CLI-side компоненты в sboxctl
- [ ] Автоматическое обнаружение агента
- [ ] Координация конфигурации
- [ ] Web UI (базовый)
- [ ] Алертинг (email, webhook)
- [ ] Backup и восстановление
- [ ] Распределенное развертывание

#### ✅ Критерии готовности для v0.3.0-beta
- [ ] sboxctl автоматически обнаруживает агента
- [ ] Конфигурация синхронизируется между компонентами
- [ ] Web UI доступен для управления
- [ ] Алертинг работает
- [ ] Backup/restore функциональность

---

### Phase 3: Production Release
**Статус**: 📋 Планируется  
**Версия**: `v1.0.0`  
**Целевая дата**: 2025-09-01  
**Длительность**: 10-15 дней

#### 🎯 Цель
Первый стабильный релиз для production.

#### 📋 Финальные компоненты
- [ ] Полная документация
- [ ] Performance оптимизации
- [ ] Security audit
- [ ] Load testing
- [ ] Production deployment guide
- [ ] Migration guide с Path A

#### ✅ Критерии готовности для v1.0.0
- [ ] 100% покрытие тестами критических компонентов
- [ ] Performance тесты пройдены
- [ ] Security audit завершен
- [ ] Документация полная
- [ ] Production deployment guide готов
- [ ] Migration guide готов

---

## 🎯 Initial Release Criteria

### Минимально готов для работы с sboxctl: `v0.1.0-alpha`

**Дата**: 2025-07-10  
**Статус**: MVP для тестирования

#### ✅ Что должно работать:
1. **Запуск sboxctl**: Агент запускает `sboxctl update` через exec.Cmd
2. **Перехват stdout**: Получает JSON события от sboxctl
3. **Обработка событий**: Парсит и обрабатывает события через dispatcher
4. **Логирование**: Сохраняет логи (in-memory)
5. **Health checks**: Базовые проверки состояния
6. **Установка**: Работает `./scripts/install.sh`
7. **Systemd service**: Запускается как systemd unit

#### 🔗 Интеграция с sboxctl:
```bash
# sboxagent запускает sboxctl
sboxagent → exec.Cmd("sboxctl", "update") → stdout pipe → JSON events

# Пример события от sboxctl
{"type":"LOG","data":{"level":"info","message":"Config updated"},"timestamp":"2025-06-27T16:30:00Z","version":"1.0"}
```

#### 🚧 Ограничения v0.1.0-alpha:
- Логи только в памяти (перезагрузка = потеря)
- Нет HTTP API
- Базовые health check'и
- Нет hot-reload конфигурации
- Нет метрик
- Только односторонняя интеграция

---

## 📊 Timeline Overview

```
Phase 0: Foundation     ✅ Завершен
├── v0.0.1 - v0.0.5    ✅ Архитектура, планы, документация

Phase 1B: Bootstrap     🚧 В разработке (8-12 дней)
├── v0.1.0-alpha       🎯 2025-07-10 - Initial Release

Phase 1C: Production    📋 Планируется (10-15 дней)
├── v0.2.0-beta        🎯 2025-07-25 - Production Ready

Phase 2: Integration    📋 Планируется (15-20 дней)
├── v0.3.0-beta        🎯 2025-08-15 - Full Integration

Phase 3: Production     📋 Планируется (10-15 дней)
├── v1.0.0             🎯 2025-09-01 - Stable Release
```

---

## 🔄 Release Process

### Alpha/Beta Releases
1. **Feature freeze** за 2 дня до релиза
2. **Testing** на staging окружении
3. **Documentation** обновление
4. **Release notes** подготовка
5. **GitHub release** создание
6. **Docker image** публикация (если применимо)

### Stable Releases (v1.0.0+)
1. **Security audit** обязателен
2. **Performance testing** обязателен
3. **Backward compatibility** проверка
4. **Migration guide** подготовка
5. **Production deployment** тестирование

---

## 🎯 Success Metrics

### Technical Metrics
- **Uptime**: > 99.9% для v1.0.0
- **Response time**: < 100ms для API endpoints
- **Memory usage**: < 50MB в idle
- **CPU usage**: < 5% в idle

### User Metrics
- **Installation success rate**: > 95%
- **Integration success rate**: > 90%
- **User satisfaction**: > 4.5/5

### Development Metrics
- **Test coverage**: > 80% для v1.0.0
- **Documentation coverage**: 100%
- **Security vulnerabilities**: 0 критических

---

**Примечание**: Этот роадмап является живым документом и будет обновляться по мере развития проекта. Даты могут корректироваться в зависимости от приоритетов и ресурсов. 