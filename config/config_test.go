package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cleanup  func()
		validate func(*testing.T, *Config)
	}{
		{
			name: "使用預設值",
			setup: func() {
				_ = os.Unsetenv("POSTGRES_DSN")
				_ = os.Unsetenv("DB_MAX_OPEN_CONNS")
				_ = os.Unsetenv("DB_MAX_IDLE_CONNS")
				_ = os.Unsetenv("PORT")
				_ = os.Unsetenv("CORS_ALLOW_ORIGINS")
			},
			cleanup: func() {},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "", cfg.Database.DSN)
				assert.Equal(t, 25, cfg.Database.MaxOpenConns)
				assert.Equal(t, 25, cfg.Database.MaxIdleConns)
				assert.Equal(t, time.Hour, cfg.Database.ConnMaxLifetime)
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.Equal(
					t,
					[]string{"http://localhost:5173", "http://localhost:4173"},
					cfg.CORS.AllowOrigins,
				)
			},
		},
		{
			name: "使用環境變數",
			setup: func() {
				_ = os.Setenv("POSTGRES_DSN", "postgres://test:test@localhost/testdb")
				_ = os.Setenv("DB_MAX_OPEN_CONNS", "50")
				_ = os.Setenv("DB_MAX_IDLE_CONNS", "10")
				_ = os.Setenv("PORT", "8080")
				_ = os.Setenv("CORS_ALLOW_ORIGINS", "https://example.com,https://app.example.com")
			},
			cleanup: func() {
				_ = os.Unsetenv("POSTGRES_DSN")
				_ = os.Unsetenv("DB_MAX_OPEN_CONNS")
				_ = os.Unsetenv("DB_MAX_IDLE_CONNS")
				_ = os.Unsetenv("PORT")
				_ = os.Unsetenv("CORS_ALLOW_ORIGINS")
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres://test:test@localhost/testdb", cfg.Database.DSN)
				assert.Equal(t, 50, cfg.Database.MaxOpenConns)
				assert.Equal(t, 10, cfg.Database.MaxIdleConns)
				assert.Equal(t, time.Hour, cfg.Database.ConnMaxLifetime)
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.Equal(
					t,
					[]string{"https://example.com", "https://app.example.com"},
					cfg.CORS.AllowOrigins,
				)
			},
		},
		{
			name: "環境變數部分設置",
			setup: func() {
				_ = os.Setenv("POSTGRES_DSN", "postgres://localhost/db")
				_ = os.Setenv("PORT", "3000")
				_ = os.Unsetenv("DB_MAX_OPEN_CONNS")
				_ = os.Unsetenv("DB_MAX_IDLE_CONNS")
				_ = os.Unsetenv("CORS_ALLOW_ORIGINS")
			},
			cleanup: func() {
				_ = os.Unsetenv("POSTGRES_DSN")
				_ = os.Unsetenv("PORT")
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres://localhost/db", cfg.Database.DSN)
				assert.Equal(t, 25, cfg.Database.MaxOpenConns)
				assert.Equal(t, 25, cfg.Database.MaxIdleConns)
				assert.Equal(t, "3000", cfg.Server.Port)
				assert.Equal(
					t,
					[]string{"http://localhost:5173", "http://localhost:4173"},
					cfg.CORS.AllowOrigins,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			cfg := Load()
			assert.NotNil(t, cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "環境變數存在",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom_value",
			setEnv:       true,
			expected:     "custom_value",
		},
		{
			name:         "環境變數不存在使用預設值",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default_value",
			envValue:     "",
			setEnv:       false,
			expected:     "default_value",
		},
		{
			name:         "環境變數為空字串使用預設值",
			key:          "TEST_KEY_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "預設值為空字串",
			key:          "TEST_KEY_DEFAULT_EMPTY",
			defaultValue: "",
			envValue:     "",
			setEnv:       false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		expected     int
	}{
		{
			name:         "環境變數為有效整數",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "100",
			setEnv:       true,
			expected:     100,
		},
		{
			name:         "環境變數不存在使用預設值",
			key:          "TEST_INT_NOT_SET",
			defaultValue: 50,
			envValue:     "",
			setEnv:       false,
			expected:     50,
		},
		{
			name:         "環境變數為無效整數使用預設值",
			key:          "TEST_INT_INVALID",
			defaultValue: 25,
			envValue:     "not_a_number",
			setEnv:       true,
			expected:     25,
		},
		{
			name:         "環境變數為負數",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 10,
			envValue:     "-5",
			setEnv:       true,
			expected:     -5,
		},
		{
			name:         "環境變數為零",
			key:          "TEST_INT_ZERO",
			defaultValue: 10,
			envValue:     "0",
			setEnv:       true,
			expected:     0,
		},
		{
			name:         "環境變數為空字串使用預設值",
			key:          "TEST_INT_EMPTY",
			defaultValue: 20,
			envValue:     "",
			setEnv:       true,
			expected:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := getEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigStructure(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			DSN:             "test_dsn",
			MaxOpenConns:    100,
			MaxIdleConns:    50,
			ConnMaxLifetime: 2 * time.Hour,
		},
		Server: ServerConfig{
			Port: "8080",
		},
		CORS: CORSConfig{
			AllowOrigins: []string{"https://example.com"},
		},
	}

	assert.Equal(t, "test_dsn", cfg.Database.DSN)
	assert.Equal(t, 100, cfg.Database.MaxOpenConns)
	assert.Equal(t, 50, cfg.Database.MaxIdleConns)
	assert.Equal(t, 2*time.Hour, cfg.Database.ConnMaxLifetime)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, []string{"https://example.com"}, cfg.CORS.AllowOrigins)
}

func TestGetEnvStringSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		envValue     string
		setEnv       bool
		expected     []string
	}{
		{
			name:         "環境變數為多個來源",
			key:          "TEST_SLICE",
			defaultValue: []string{"http://localhost:5173"},
			envValue:     "https://a.com,https://b.com",
			setEnv:       true,
			expected:     []string{"https://a.com", "https://b.com"},
		},
		{
			name:         "環境變數為單一來源",
			key:          "TEST_SLICE_SINGLE",
			defaultValue: []string{"http://localhost:5173"},
			envValue:     "https://example.com",
			setEnv:       true,
			expected:     []string{"https://example.com"},
		},
		{
			name:         "環境變數不存在使用預設值",
			key:          "TEST_SLICE_NOT_SET",
			defaultValue: []string{"http://localhost:5173", "http://localhost:4173"},
			envValue:     "",
			setEnv:       false,
			expected:     []string{"http://localhost:5173", "http://localhost:4173"},
		},
		{
			name:         "環境變數為空字串使用預設值",
			key:          "TEST_SLICE_EMPTY",
			defaultValue: []string{"http://localhost:5173"},
			envValue:     "",
			setEnv:       true,
			expected:     []string{"http://localhost:5173"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := getEnvStringSlice(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
