package abstractrepo

import (
	"context"
	"testing"

	"github.com/aawadallak/go-core-kit/core/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db.Debug(), nil
}

func Test_AbstractRepository_With_Soft_Delete(t *testing.T) {
	type Sample struct {
		repository.Entity
		Value string
	}

	tests := []struct {
		name     string
		testFunc func(t *testing.T, repo *AbstractRepository[Sample], assert *assert.Assertions)
	}{
		{
			name: "Delete_Success_With_Soft_Delete",
			testFunc: func(t *testing.T, repo *AbstractRepository[Sample], assert *assert.Assertions) {
				// Save initial sample
				sample := &Sample{Value: "test"}
				saved, err := repo.Save(context.Background(), sample)

				assert.NoError(err, "Should save without error")
				assert.NotNil(saved, "Should return saved entity")

				// Delete the sample
				err = repo.Delete(context.Background(), saved)
				assert.NoError(err, "Should delete without error")

				// Verify deletion (soft delete)
				var count int64
				err = repo.db.Model(&Sample{}).Where("id = ?", saved.ID).Count(&count).Error
				assert.NoError(err, "Should query count without error")
				assert.Equal(int64(1), count, "Record should still exist in database")

				var softDeleted Sample
				err = repo.db.Unscoped().Where("id = ?", saved.ID).First(&softDeleted).Error
				assert.NoError(err, "Should find soft-deleted record")
				assert.NotNil(softDeleted.DeletedAt, "DeletedAt should be set")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, err := setupTestDB()
			if err != nil {
				t.Fatalf("failed to setup database: %v", err)
			}

			opts := []Option{WithMigrate()}
			repo, err := NewAbstractRepository[Sample](db, opts...)
			if err != nil {
				t.Fatalf("failed to create repository: %v", err)
			}

			// Run test
			assert := assert.New(t)
			tt.testFunc(t, repo, assert)
		})
	}
}

func Test_AbstractRepository_With_Hard_Delete(t *testing.T) {
	type Sample struct {
		repository.Entity
		Value string
	}

	tests := []struct {
		name     string
		testFunc func(t *testing.T, repo *AbstractRepository[Sample], assert *assert.Assertions)
	}{
		{
			name: "Delete_Success_With_Hard_Delete",
			testFunc: func(t *testing.T, repo *AbstractRepository[Sample], assert *assert.Assertions) {
				// Save initial sample
				sample := &Sample{Value: "test"}
				saved, err := repo.Save(context.Background(), sample)

				assert.NoError(err, "Should save without error")
				assert.NotNil(saved, "Should return saved entity")

				// Delete the sample
				err = repo.Delete(context.Background(), saved)
				assert.NoError(err, "Should delete without error")

				// Verify deletion (soft delete)
				var count int64
				err = repo.db.Model(&Sample{}).Where("id = ?", saved.ID).Count(&count).Error
				assert.NoError(err, "Should query count without error")
				assert.Equal(int64(0), count, "Record should still exist in database")

				var softDeleted Sample
				err = repo.db.Unscoped().Where("id = ?", saved.ID).First(&softDeleted).Error
				assert.Error(err, "Should find soft-deleted record")
				assert.Nil(softDeleted.DeletedAt, "DeletedAt should be nil")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, err := setupTestDB()
			if err != nil {
				t.Fatalf("failed to setup database: %v", err)
			}

			opts := []Option{WithMigrate(), WithHardDelete()}
			repo, err := NewAbstractRepository[Sample](db, opts...)
			if err != nil {
				t.Fatalf("failed to create repository: %v", err)
			}

			// Run test
			assert := assert.New(t)
			tt.testFunc(t, repo, assert)
		})
	}
}
