# Phase 3: API Integration & Advanced Features

## Overview
Phase 3 focuses on enhancing the ecosystem with API integration, advanced configuration management, and improved user experience while keeping SaaS plans separate.

## Goals
- REST API for sboxmgr (core functionality)
- Enhanced agent capabilities
- Advanced configuration management
- Improved monitoring and alerting
- Better error handling and recovery

## Architecture Decisions

### sboxmgr REST API
- **Keep sboxmgr as the core** - no SaaS integration yet
- **REST API interface** for programmatic access
- **CLI remains primary** for local usage
- **API for future SaaS** - but not implemented yet

### Agent Enhancements
- **API client** to connect to sboxmgr REST API
- **Fallback to file import** if API unavailable
- **Enhanced monitoring** with metrics export
- **Configuration hot-reload** capability

## Phase 3 Components

### 1. sboxmgr REST API
**Location:** `sboxmgr/api/`
**Features:**
- REST API server (FastAPI/Flask)
- Authentication (optional)
- Configuration management endpoints
- Subscription management endpoints
- Health check endpoints
- OpenAPI documentation

**Endpoints:**
```
GET  /api/v1/configs          - List configurations
POST /api/v1/configs          - Generate new config
GET  /api/v1/configs/{id}     - Get specific config
PUT  /api/v1/configs/{id}     - Update config
DELETE /api/v1/configs/{id}   - Delete config

GET  /api/v1/subscriptions    - List subscriptions
POST /api/v1/subscriptions    - Add subscription
PUT  /api/v1/subscriptions/{id} - Update subscription
DELETE /api/v1/subscriptions/{id} - Remove subscription

GET  /api/v1/health           - Health check
GET  /api/v1/status           - Service status
```

### 2. Agent API Client
**Location:** `internal/services/api/`
**Features:**
- HTTP client for sboxmgr API
- Authentication support
- Retry logic with exponential backoff
- Connection pooling
- Request/response logging

**Methods:**
```go
type APIClient interface {
    GetConfigs() ([]Config, error)
    GetConfig(id string) (*Config, error)
    GenerateConfig(request ConfigRequest) (*Config, error)
    UpdateConfig(id string, request ConfigRequest) (*Config, error)
    DeleteConfig(id string) error
    
    GetSubscriptions() ([]Subscription, error)
    AddSubscription(sub Subscription) error
    UpdateSubscription(id string, sub Subscription) error
    DeleteSubscription(id string) error
    
    HealthCheck() (*HealthStatus, error)
}
```

### 3. Enhanced Configuration Management
**Location:** `internal/config/`
**Features:**
- Configuration versioning
- Rollback capability
- Configuration templates
- Environment-specific configs
- Configuration validation rules

**New Structures:**
```go
type ConfigVersion struct {
    ID          string    `json:"id"`
    Version     int       `json:"version"`
    Timestamp   time.Time `json:"timestamp"`
    Description string    `json:"description"`
    Config      *Config   `json:"config"`
}

type ConfigTemplate struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
    Template    string                 `json:"template"`
}
```

### 4. Advanced Monitoring
**Location:** `internal/services/monitor/`
**Features:**
- Metrics export (Prometheus format)
- Custom alerting rules
- Performance profiling
- Resource usage tracking
- Historical data retention

**Metrics:**
- Agent uptime and health
- Configuration apply success/failure rates
- API request latency and success rates
- System resource usage (CPU, memory, disk)
- Network connectivity status

### 5. Hot-Reload Capability
**Location:** `internal/services/reloader/`
**Features:**
- Watch for configuration changes
- Automatic reload without restart
- Validation before apply
- Rollback on failure
- Reload notifications

## Implementation Plan

### Week 1-2: sboxmgr REST API
- [ ] Set up FastAPI/Flask server
- [ ] Implement basic CRUD endpoints
- [ ] Add authentication (optional)
- [ ] Create OpenAPI documentation
- [ ] Add health check endpoints
- [ ] Write API tests

### Week 3-4: Agent API Client
- [ ] Create HTTP client package
- [ ] Implement API client interface
- [ ] Add retry and error handling
- [ ] Integrate with existing agent
- [ ] Add fallback to file import
- [ ] Write integration tests

### Week 5-6: Enhanced Configuration
- [ ] Implement configuration versioning
- [ ] Add rollback capability
- [ ] Create configuration templates
- [ ] Add environment-specific configs
- [ ] Enhance validation rules
- [ ] Write configuration tests

### Week 7-8: Advanced Monitoring
- [ ] Implement metrics collection
- [ ] Add Prometheus export
- [ ] Create custom alerting
- [ ] Add performance profiling
- [ ] Implement historical data
- [ ] Write monitoring tests

### Week 9-10: Hot-Reload & Polish
- [ ] Implement configuration watching
- [ ] Add automatic reload
- [ ] Create rollback mechanism
- [ ] Add reload notifications
- [ ] Performance optimization
- [ ] Final integration tests

## Testing Strategy

### Unit Tests
- API client methods
- Configuration management
- Monitoring metrics
- Hot-reload logic

### Integration Tests
- sboxmgr API + agent integration
- Configuration versioning
- Monitoring data flow
- Hot-reload scenarios

### End-to-End Tests
- Full API workflow
- Configuration lifecycle
- Monitoring and alerting
- Error recovery scenarios

## Success Criteria

### Functional
- [ ] sboxmgr exposes REST API
- [ ] Agent can connect to API
- [ ] Configuration versioning works
- [ ] Hot-reload functions properly
- [ ] Advanced monitoring provides insights

### Performance
- [ ] API response time < 100ms
- [ ] Configuration reload < 5s
- [ ] Memory usage < 50MB
- [ ] CPU usage < 5% idle

### Reliability
- [ ] 99.9% uptime
- [ ] Graceful error handling
- [ ] Automatic recovery
- [ ] Data consistency

## Future Considerations (Phase 4+)

### SaaS Integration
- Multi-tenant API
- User management
- Subscription billing
- Web dashboard

### Advanced Features
- Configuration templates marketplace
- Automated testing
- CI/CD integration
- Multi-region support

### Security
- API rate limiting
- Request signing
- Audit logging
- Encryption at rest

## Dependencies

### External
- FastAPI/Flask for REST API
- Prometheus client for metrics
- Redis (optional) for caching

### Internal
- Phase 2 completion
- sbox-common schemas
- Existing agent architecture

## Risks & Mitigation

### API Complexity
- **Risk:** Over-engineering the API
- **Mitigation:** Start simple, iterate based on needs

### Performance Impact
- **Risk:** API calls slow down agent
- **Mitigation:** Implement caching and async operations

### Backward Compatibility
- **Risk:** Breaking existing CLI workflows
- **Mitigation:** Keep CLI as primary interface

### Security
- **Risk:** API security vulnerabilities
- **Mitigation:** Implement proper authentication and validation

## Timeline
- **Total Duration:** 10 weeks
- **Start Date:** After Phase 2 completion
- **End Date:** TBD based on Phase 2 completion

## Resources
- 1 Backend Developer (sboxmgr API)
- 1 Go Developer (agent enhancements)
- 1 DevOps Engineer (monitoring setup) 