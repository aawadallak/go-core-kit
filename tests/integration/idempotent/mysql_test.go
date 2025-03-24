package idempotent

import (
	"context"
	"testing"

	idem "github.com/aawadallak/go-core-kit/plugin/idempotent"
	idemgorm "github.com/aawadallak/go-core-kit/plugin/idempotent/storage/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMySQLContainer(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "testpass",
			"MYSQL_DATABASE":      "testdb",
			"MYSQL_USER":          "testuser",
			"MYSQL_PASSWORD":      "testpass",
		},
		WaitingFor: wait.ForListeningPort("3306/tcp"),
	}

	mysqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}

	host, err := mysqlC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get mysql container host: %v", err)
	}

	port, err := mysqlC.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("failed to get mysql container port: %v", err)
	}

	dsn := "testuser:testpass@tcp(" + host + ":" + port.Port() + ")/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to mysql with gorm: %v", err)
	}

	return db, func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		mysqlC.Terminate(ctx)
	}
}

func TestIdempotentFlowMySQL(t *testing.T) {
	db, teardown := setupMySQLContainer(t)
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
