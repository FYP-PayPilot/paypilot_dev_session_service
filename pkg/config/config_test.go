package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Test loading from default config file
	cfg, err := Load("../../configs/config.yaml")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.GetDSN()
	expected := "host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestRabbitMQConfig_GetRabbitMQURL(t *testing.T) {
	cfg := &RabbitMQConfig{
		Host:     "localhost",
		Port:     5672,
		User:     "guest",
		Password: "guest",
		VHost:    "/",
	}

	url := cfg.GetRabbitMQURL()
	expected := "amqp://guest:guest@localhost:5672/"
	assert.Equal(t, expected, url)
}

func TestLoadWithEnvOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := Load("../../configs/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
}
