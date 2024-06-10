//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestE2EStartupScenario(t *testing.T) {
	ctx := context.Background()

	// Run PostgreSQL containers
	xkcdPostgresContainer, err := startPostgresContainer(ctx)
	require.NoError(t, err)
	defer xkcdPostgresContainer.Terminate(ctx)

	authPostgresContainer, err := startPostgresContainer(ctx)
	require.NoError(t, err)
	defer authPostgresContainer.Terminate(ctx)

	// Run Redis container
	redisContainer, err := startRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	// Get URLs for PostgreSQL and Redis
	xkcdPostgresURL, err := getPostgresURL(ctx, xkcdPostgresContainer)
	require.NoError(t, err)
	authPostgresURL, err := getPostgresURL(ctx, authPostgresContainer)
	require.NoError(t, err)
	redisURL, err := getRedisURL(ctx, redisContainer)
	require.NoError(t, err)

	// Create config file
	createXkcdConfigFile(xkcdPostgresURL, redisURL)
	createAuthConfigFile(authPostgresURL)

	// Run the auth server
	authCmd := exec.Command(
		"go",
		"run",
		"../cmd/authserver/main.go",
		"-c",
		"config/authserver.test.yaml",
		"-p",
		"50051",
	)

	err = authCmd.Start()
	require.NoError(t, err)
	defer func() {
		_ = authCmd.Process.Kill()
	}()

	time.Sleep(5 * time.Second) // Wait for the server to start

	// Run the xkcd server
	cmd := exec.Command(
		"go",
		"run",
		"../cmd/xkcdserver/main.go",
		"-c",
		"config/xkcdserver.test.yaml",
		"-p",
		"8081",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	require.NoError(t, err)
	defer func() {
		_ = cmd.Process.Kill()
	}()

	time.Sleep(5 * time.Second) // Wait for the server to start

	// Login
	loginReqBody := map[string]string{
		"username": "admin",
		"password": "password",
	}
	loginReqBytes, _ := json.Marshal(loginReqBody)
	loginReq, err := http.NewRequest("POST", "http://localhost:8081/login", bytes.NewBuffer(loginReqBytes))
	require.NoError(t, err)

	loginResp, err := http.DefaultClient.Do(loginReq)
	require.NoError(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginRespBody map[string]string
	err = json.NewDecoder(loginResp.Body).Decode(&loginRespBody)
	require.NoError(t, err)

	token, ok := loginRespBody["token"]
	require.True(t, ok)

	// Update comics
	updateReq, err := http.NewRequest("POST", "http://localhost:8081/update", nil)
	require.NoError(t, err)
	updateReq.Header.Set("Authorization", "Bearer "+token)

	updateResp, err := http.DefaultClient.Do(updateReq)
	require.NoError(t, err)
	defer updateResp.Body.Close()
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	// Search for comics with "apple" and "doctor" and check if the expected comic is in the results
	searchReq, err := http.NewRequest("GET", "http://localhost:8081/pics?search=apple,doctor", nil)
	require.NoError(t, err)
	searchReq.Header.Set("Authorization", "Bearer "+token)

	searchResp, err := http.DefaultClient.Do(searchReq)
	require.NoError(t, err)
	defer searchResp.Body.Close()
	assert.Equal(t, http.StatusOK, searchResp.StatusCode)

	body, err := io.ReadAll(searchResp.Body)
	require.NoError(t, err)

	var results []string
	err = json.Unmarshal(body, &results)
	require.NoError(t, err)
	expectedURL := "https://imgs.xkcd.com/comics/an_apple_a_day.png"
	found := false
	for _, url := range results {
		if strings.Contains(url, expectedURL) {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected comic not found in search results")
}

func startPostgresContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func startRedisContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:6",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func getPostgresURL(ctx context.Context, container testcontainers.Container) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port()), nil
}

func getRedisURL(ctx context.Context, container testcontainers.Container) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("redis://%s:%s", host, port.Port()), nil
}

func createXkcdConfigFile(postgresURL, redisURL string) {
	config := fmt.Sprintf(`postgres_url: "%s"
redis_url: "%s"
source_url: "https://xkcd.com"
max_comics_load: 0
parallel: 20
gaps_limit: 2
update_time: "00:00"
token_max_time: 15
rate_limit: 10
max_tokens: 100
concurrency_limit: 10
auth_server_url: "localhost:50051"
`, postgresURL, redisURL)

	err := os.WriteFile("config/xkcdserver.test.yaml", []byte(config), 0644)
	if err != nil {
		panic("Writing config file failed")
	}
}

func createAuthConfigFile(postgresURL string) {
	config := fmt.Sprintf(`postgres_url: "%s"
token_max_time: 15
`, postgresURL)

	err := os.WriteFile("config/authserver.test.yaml", []byte(config), 0644)
	if err != nil {
		panic("Writing config file failed")
	}
}
