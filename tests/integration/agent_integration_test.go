package integration

import (
	"os/exec"
	"os"
	"testing"
	"time"
	"io/ioutil"
	"path/filepath"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestAgent_Integration_MockSboxctl(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "agent_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "agent.yaml")
	logPath := filepath.Join(tempDir, "agent.log")

	configContent := `
agent:
  name: "integration-test"
  version: "0.1.0"
  log_level: "debug"
server:
  port: 8080
  host: "127.0.0.1"
  timeout: "30s"
services:
  sboxctl:
    enabled: true
    command:
      - echo
      - '{"type":"LOG","data":{"level":"info","message":"IntegrationTest"},"timestamp":"2025-06-27T16:30:00Z","version":"1.0"}'
    interval: "1m"
    timeout: "10s"
    stdout_capture: true
    health_check:
      enabled: false
logging:
  stdout_capture: true
  aggregation: true
  retention_days: 1
  max_entries: 100
security:
  allow_remote_api: false
  api_token: "test-token"
  allowed_hosts: ["127.0.0.1"]
  tls_enabled: false
`

	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	sboxagentPath := "./sboxagent"
	if _, err := os.Stat(sboxagentPath); os.IsNotExist(err) {
		// Автоматически собираем бинарник если его нет
		buildCmd := exec.Command("go", "build", "-o", sboxagentPath, "../../cmd/sboxagent")
		buildCmd.Dir = "."
		if err := buildCmd.Run(); err != nil {
			t.Skipf("Failed to build sboxagent binary: %v", err)
		}
	}

	cmd := exec.Command(sboxagentPath, "-config", configPath, "-debug")
	logFile, err := os.Create(logPath)
	require.NoError(t, err)
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Запускаем агент (он выполнит echo и завершится по таймауту)
	err = cmd.Start()
	require.NoError(t, err)

	// Ждём немного, чтобы агент успел обработать событие
	time.Sleep(2 * time.Second)

	// Завершаем процесс
	_ = cmd.Process.Kill()
	cmd.Wait()

	// Читаем логи
	logData, err := os.ReadFile(logPath)
	require.NoError(t, err)
	logStr := string(logData)

	// Проверяем, что в логах есть событие
	assert.Contains(t, logStr, "Processing sboxctl event")
	assert.Contains(t, logStr, "IntegrationTest")
}

func TestAgent_InvalidConfig(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "agent_invalid_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "invalid_agent.yaml")
	logPath := filepath.Join(tempDir, "invalid_agent.log")

	// Некорректный YAML (отсутствует обязательный блок agent и синтаксическая ошибка)
	invalidConfig := `
services:
  sboxctl:
    enabled: true
    command:
      - echo
      - 'test
    interval: "1m"
    timeout: "10s"
    stdout_capture: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(invalidConfig), 0644))

	sboxagentPath := "./sboxagent"
	if _, err := os.Stat(sboxagentPath); os.IsNotExist(err) {
		// Автоматически собираем бинарник если его нет
		buildCmd := exec.Command("go", "build", "-o", sboxagentPath, "../../cmd/sboxagent")
		buildCmd.Dir = "."
		if err := buildCmd.Run(); err != nil {
			t.Skipf("Failed to build sboxagent binary: %v", err)
		}
	}

	cmd := exec.Command(sboxagentPath, "-config", configPath)
	logFile, err := os.Create(logPath)
	require.NoError(t, err)
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	require.NoError(t, err)

	// Ждём завершения процесса (он должен быстро завершиться с ошибкой)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		// Агент должен завершиться с ошибкой
		assert.Error(t, err)
	case <-time.After(3 * time.Second):
		t.Error("Agent did not exit in time with invalid config")
		_ = cmd.Process.Kill()
	}

	logData, err := os.ReadFile(logPath)
	if err == nil {
		logStr := string(logData)
		assert.Contains(t, logStr, "Failed to load configuration")
	}
}

func TestAgent_SboxctlDisabled(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "agent_sboxctl_disabled_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "agent_disabled.yaml")
	logPath := filepath.Join(tempDir, "agent_disabled.log")

	configContent := `
agent:
  name: "integration-test-disabled"
  version: "0.1.0"
  log_level: "debug"
server:
  port: 8080
  host: "127.0.0.1"
  timeout: "30s"
services:
  sboxctl:
    enabled: false
    command: ["echo", "should not run"]
    interval: "1m"
    timeout: "10s"
    stdout_capture: true
    health_check:
      enabled: false
logging:
  stdout_capture: true
  aggregation: true
  retention_days: 1
  max_entries: 100
security:
  allow_remote_api: false
  api_token: "test-token"
  allowed_hosts: ["127.0.0.1"]
  tls_enabled: false
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	sboxagentPath := "./sboxagent"
	if _, err := os.Stat(sboxagentPath); os.IsNotExist(err) {
		// Автоматически собираем бинарник если его нет
		buildCmd := exec.Command("go", "build", "-o", sboxagentPath, "../../cmd/sboxagent")
		buildCmd.Dir = "."
		if err := buildCmd.Run(); err != nil {
			t.Skipf("Failed to build sboxagent binary: %v", err)
		}
	}

	cmd := exec.Command(sboxagentPath, "-config", configPath, "-debug")
	logFile, err := os.Create(logPath)
	require.NoError(t, err)
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	require.NoError(t, err)

	// Ждём немного, чтобы агент успел стартовать
	time.Sleep(2 * time.Second)

	// Завершаем процесс
	_ = cmd.Process.Kill()
	cmd.Wait()

	logData, err := os.ReadFile(logPath)
	require.NoError(t, err)
	logStr := string(logData)

	// Проверяем, что сервис sboxctl не стартовал и не было событий
	assert.Contains(t, logStr, "Agent starting")
	assert.NotContains(t, logStr, "Sboxctl service starting")
	assert.NotContains(t, logStr, "Processing sboxctl event")
} 