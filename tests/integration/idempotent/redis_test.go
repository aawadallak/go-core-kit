package idempotent

import (
	"context"
	"testing"

	idem "github.com/aawadallak/go-core-kit/plugin/idempotent"
	idemredis "github.com/aawadallak/go-core-kit/plugin/idempotent/storage/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis container host: %v", err)
	}

	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get redis container port: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	return client, func() {
		client.Close()
		redisC.Terminate(ctx)
	}
}

func TestIdempotentFlow(t *testing.T) {
	client, teardown := setupRedisContainer(t)
	defer teardown()

	type Sample struct {
		Value string `json:"value"`
	}

	repo, err := idemredis.NewRepository(client)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	idFn, err := idem.NewHandler[Sample](repo)
	assert.NoError(t, err)

	ctx := context.Background()
	key := "test-key"

	given := Sample{
		Value: "test-value",
	}

	hookExecutionCount := 0

	hook := func(ctx context.Context) (Sample, error) {
		hookExecutionCount++
		return Sample{
			Value: "test-value",
		}, nil
	}

	result, err := idFn.Wrap(ctx, key, hook)
	assert.NoError(t, err)
	assert.Equal(t, given.Value, result.Value)

	result, err = idFn.Wrap(ctx, key, hook)
	assert.NoError(t, err)
	assert.Equal(t, given.Value, result.Value)

	assert.Equal(t, 1, hookExecutionCount, "hook should be executed exactly once")
}
