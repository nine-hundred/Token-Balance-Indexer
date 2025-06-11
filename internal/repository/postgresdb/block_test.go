package postgresdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func TestRepository_UpsertBalance(t *testing.T) {
	dsn := "host=localhost user=postgres password=password dbname=onbloc port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.Nil(t, err)
	repo := NewRepository(db)

	err = repo.UpsertBalance(context.TODO(), "kcm", "1234", 100)
	assert.Nil(t, err)

	err = repo.UpsertBalance(context.TODO(), "kcm", "1234", 100)
	assert.Nil(t, err)
}
