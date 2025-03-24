package idempotent

import (
	"context"
	"testing"

	idem "github.com/aawadallak/go-core-kit/plugin/idempotent"
	idemgorm "github.com/aawadallak/go-core-kit/plugin/idempotent/storage/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupPostgresContainer(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := pgC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres container host: %v", err)
	}

	port, err := pgC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get postgres container port: %v", err)
	}

	dsn := "host=" + host + " user=testuser password=testpass dbname=testdb port=" + port.Port() + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres with gorm: %v", err)
	}

	return db, func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		pgC.Terminate(ctx)
	}
}

func TestIdempotentFlowPostgres(t *testing.T) {
	db, teardown := setupPostgresContainer(t)
	defer teardown()

	type Sample struct {
		Value string `json:"value"`
	}

	repo, err := idemgorm.NewRepository(db)
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
