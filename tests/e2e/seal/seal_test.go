package seal_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aawadallak/go-core-kit/plugin/seal"
	sealgorm "github.com/aawadallak/go-core-kit/plugin/seal/provider/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&seal.SealedMessage{})
	require.NoError(t, err)

	return db
}

func TestSealUnsealFlow(t *testing.T) {
	db := setupPostgres(t)
	ctx := context.Background()

	repo, err := sealgorm.NewRepository(db)
	require.NoError(t, err)

	svc := seal.New(&seal.SealerDependencies{
		RepoSealedMessage: repo,
	})

	payload := json.RawMessage(`{"user":"alice","amount":100}`)

	// Seal a payload and get a signature.
	sealOut, err := svc.Seal(ctx, &seal.SealInput{
		ExternalID: "order-123",
		Payload:    payload,
		ExpiresIn:  5 * time.Minute,
	})
	require.NoError(t, err)
	require.NotEmpty(t, sealOut.Signature)

	// Unseal with the signature and verify the payload matches.
	unsealOut, err := svc.Unseal(ctx, &seal.UnsealInput{
		Signature: sealOut.Signature,
	})
	require.NoError(t, err)
	assert.Equal(t, "order-123", unsealOut.ExternalID)

	// The payload is double-encoded (json.Marshal of json.RawMessage), so unmarshal once.
	var got json.RawMessage
	require.NoError(t, json.Unmarshal(unsealOut.Payload, &got))
	assert.JSONEq(t, string(payload), string(got))
}

func TestSealUnsealNonceCheck(t *testing.T) {
	db := setupPostgres(t)
	ctx := context.Background()

	repo, err := sealgorm.NewRepository(db)
	require.NoError(t, err)

	svc := seal.New(&seal.SealerDependencies{
		RepoSealedMessage: repo,
	})

	sealOut, err := svc.Seal(ctx, &seal.SealInput{
		ExternalID: "order-456",
		Payload:    json.RawMessage(`{"data":"test"}`),
		ExpiresIn:  5 * time.Minute,
	})
	require.NoError(t, err)

	// First unseal succeeds.
	_, err = svc.Unseal(ctx, &seal.UnsealInput{Signature: sealOut.Signature})
	require.NoError(t, err)

	// Second unseal with the same signature should fail with ErrUsedSignature.
	_, err = svc.Unseal(ctx, &seal.UnsealInput{Signature: sealOut.Signature})
	require.Error(t, err)
	assert.True(t, errors.Is(err, seal.ErrUsedSignature),
		fmt.Sprintf("expected ErrUsedSignature, got: %v", err))
}

func TestSealUnsealExpired(t *testing.T) {
	db := setupPostgres(t)
	ctx := context.Background()

	repo, err := sealgorm.NewRepository(db)
	require.NoError(t, err)

	svc := seal.New(&seal.SealerDependencies{
		RepoSealedMessage: repo,
	})

	// Seal with a very short expiration (1ms).
	sealOut, err := svc.Seal(ctx, &seal.SealInput{
		ExternalID: "order-789",
		Payload:    json.RawMessage(`{"data":"expires"}`),
		ExpiresIn:  time.Millisecond,
	})
	require.NoError(t, err)

	// Wait for the token to expire.
	time.Sleep(50 * time.Millisecond)

	// Unseal should fail with ErrSealSignatureExpired.
	_, err = svc.Unseal(ctx, &seal.UnsealInput{Signature: sealOut.Signature})
	require.Error(t, err)
	assert.True(t, errors.Is(err, seal.ErrSealSignatureExpired),
		fmt.Sprintf("expected ErrSealSignatureExpired, got: %v", err))
}
