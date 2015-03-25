package db

import (
	"testing"
)

func TestSQLDB(t *testing.T) {
	db := NewSQLDB(":memory:")
	exerciseDB(t, db)
}
