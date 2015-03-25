package db

import (
	"testing"
)

func TestRAM(t *testing.T) {
	db := NewRAMDB()
	exerciseDB(t, db)
}
